package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"homelight/internal/lighter"
)

var lampController *lighter.Controller
var mqttClient mqtt.Client

type lampStateUpdate struct {
	lampID int
	on     bool
}

type deviceAvailabilityUpdate struct {
	devId     int
	available bool
}

func startPublisher(
	ctx context.Context,
	client mqtt.Client,
	lampCh <-chan lampStateUpdate,
	deviceCh <-chan deviceAvailabilityUpdate,
	done chan<- struct{},
) {
	go func() {
		defer close(done)
		for {
			select {
			case upd, ok := <-lampCh:
				if !ok {
					lampCh = nil
					continue
				}
				if err := PublishLampState(client, upd.lampID, upd.on); err != nil {
					log.Printf("Cannot publish lamp state: %s", err)
				}
			case upd, ok := <-deviceCh:
				if !ok {
					deviceCh = nil
					continue
				}
				log.Printf("Availability update: device %d available=%v", upd.devId, upd.available)
				if upd.available {
					if err := PublishHomeLightDiscovery(client); err != nil {
						log.Printf("Cannot publish device discovery: %s", err)
						continue
					}
				}
				if err := PublishDeviceAvailabilityState(client, upd.devId, upd.available); err != nil {
					log.Printf("Cannot publish device availability: %s", err)
				}
			case <-ctx.Done():
				return
			}
		}
	}()
}

func messagePubHandler(client mqtt.Client, msg mqtt.Message) {
	log.Printf("Received message: %s from topic: %s\n", msg.Payload(), msg.Topic())
	topic := msg.Topic()
	if strings.HasPrefix(topic, topicPrefix) && strings.HasSuffix(topic, setTopicSuffix) {
		lampIdStr := topic[len(topicPrefix) : len(topic)-len(setTopicSuffix)]
		lampID, err := strconv.Atoi(lampIdStr)
		if err != nil {
			log.Printf("Failed to convert '%s' to lamp id: %s\n", lampIdStr, err)
			return
		}
		var on bool
		payload := string(msg.Payload())
		switch payload {
		case "OFF":
			on = false
		case "ON":
			on = true
		default:
			log.Printf("Unknown action %s\n", msg.Payload())
			return
		}
		lampController.ChangeLamp(lampID, on)
	}
}

func connectHandler(client mqtt.Client) {
	log.Printf("Connected\n")
	if err := processConnection(client); err != nil {
		log.Printf("Error while processing connect: %s", err)
		client.Disconnect(0)
	}
}

func processConnection(client mqtt.Client) error {
	if err := PublishAutodiscoveryAndState(client); err != nil {
		return err
	}

	topic := "homelight/+/set"
	token := client.Subscribe(topic, 1, nil)
	token.Wait()
	return nil
}

func connectLostHandler(client mqtt.Client, err error) {
	log.Printf("Connect lost: %v\n", err)
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	for ctx.Err() == nil {
		if err := run(ctx); err != nil && ctx.Err() == nil {
			log.Printf("Error: %s.", err)
		}
	}
}

func run(ctx context.Context) error {
	log.Printf("homelight mqtt: starting")
	lampPublishCh := make(chan lampStateUpdate, 64)
	devicePublishCh := make(chan deviceAvailabilityUpdate, 64)

	onLampChange := func(lampID int, on bool) {
		select {
		case lampPublishCh <- lampStateUpdate{lampID: lampID, on: on}:
		case <-ctx.Done():
		}
	}
	onDeviceAvailabilityChange := func(devId int, available bool, firstSeen bool) {
		select {
		case devicePublishCh <- deviceAvailabilityUpdate{
			devId:     devId,
			available: available,
		}:
		case <-ctx.Done():
		default:
			log.Printf("Dropped availability update for device %d: channel full", devId)
		}
	}

	lampController = lighter.NewController(ctx, onLampChange, onDeviceAvailabilityChange)

	publisherDone := make(chan struct{})
	opts := mqtt.NewClientOptions()
	opts.AddBroker(fmt.Sprintf("tcp://%s:%d", "localhost", 1883))
	opts.SetClientID("go_mqtt_client")
	opts.SetDefaultPublishHandler(messagePubHandler)
	opts.OnConnect = connectHandler
	opts.OnConnectionLost = connectLostHandler
	mqttClient = mqtt.NewClient(opts)
	if token := mqttClient.Connect(); token.Wait() && token.Error() != nil {
		return token.Error()
	}
	defer mqttClient.Disconnect(250)

	startPublisher(ctx, mqttClient, lampPublishCh, devicePublishCh, publisherDone)
	startAvailabilitySync(ctx, mqttClient)
	defer func() {
		<-publisherDone
	}()

	return lampController.Run()
}

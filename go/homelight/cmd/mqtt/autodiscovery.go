package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"homelight/internal/lighter"
)

const homeLightDiscoveryTopic = "homeassistant/device/homelight/config"

// Possible RS485 slave ids we may have ever published availability for.
const maxAvailabilityDevId = 32

var discoveryMu sync.Mutex
var clearedLegacyDiscovery bool

func homeLightDevice() map[string]any {
	return map[string]any{
		"identifiers": []string{"homelight"},
		"name":        "HomeLight",
	}
}

func homeLightOrigin() map[string]any {
	return map[string]any{
		"name": "HomeLight",
	}
}

func clearDiscovery(client mqtt.Client, topic string) error {
	token := client.Publish(topic, 0, true, "")
	token.Wait()
	return token.Error()
}

func clearLegacyDiscoveryOnce(client mqtt.Client) error {
	discoveryMu.Lock()
	defer discoveryMu.Unlock()
	if clearedLegacyDiscovery {
		return nil
	}
	log.Printf("Clearing legacy MQTT discovery topics")
	for lampID := range lighter.LampNames {
		if err := clearDiscovery(client, fmt.Sprintf("homeassistant/switch/homelight_switch%d/config", lampID)); err != nil {
			return err
		}
	}
	for devId := 0; devId <= maxAvailabilityDevId; devId++ {
		if err := clearDiscovery(client, fmt.Sprintf("homeassistant/binary_sensor/homelight_device_availability%d/config", devId)); err != nil {
			return err
		}
	}
	_ = clearDiscovery(client, "homeassistant/binary_sensor/homelight_device_availabilitymaster/config")
	clearedLegacyDiscovery = true
	return nil
}

func availabilityUniqueID(devId int) string {
	return fmt.Sprintf("homelight_avail_%d", devId)
}

func PublishHomeLightDiscovery(client mqtt.Client) error {
	discoveryMu.Lock()
	defer discoveryMu.Unlock()

	knownDevIds := lampController.KnownDevices()
	cmps := make(map[string]any, len(lighter.LampNames)+len(knownDevIds))
	for lampID, lampName := range lighter.LampNames {
		cmps[fmt.Sprintf("switch_%d", lampID)] = map[string]any{
			"p":             "switch",
			"unique_id":     fmt.Sprintf("lamp_%s", lampName),
			"name":          lampName,
			"state_topic":   fmt.Sprintf("%s%d%s", topicPrefix, lampID, stateTopicSuffix),
			"command_topic": fmt.Sprintf("%s%d%s", topicPrefix, lampID, setTopicSuffix),
		}
	}
	for _, devId := range knownDevIds {
		cmps[fmt.Sprintf("availability_%d", devId)] = map[string]any{
			"p":            "binary_sensor",
			"unique_id":    availabilityUniqueID(devId),
			"name":         deviceAvailabilityName(devId),
			"state_topic":  fmt.Sprintf("%s%d%s", deviceAvailabilityPrefix, devId, stateTopicSuffix),
			"payload_on":   "online",
			"payload_off":  "offline",
			"device_class": "connectivity",
		}
	}

	payload, err := json.Marshal(map[string]any{
		"device": homeLightDevice(),
		"origin": homeLightOrigin(),
		"cmps":   cmps,
	})
	if err != nil {
		return err
	}
	log.Printf("Publishing HomeLight discovery: %d lamps, availability=%v", len(lighter.LampNames), knownDevIds)
	token := client.Publish(homeLightDiscoveryTopic, 0, true, payload)
	token.Wait()
	return token.Error()
}

func PublishLampState(client mqtt.Client, lampID int, on bool) error {
	state := "OFF"
	if on {
		state = "ON"
	}
	token := client.Publish(fmt.Sprintf("%s%d%s", topicPrefix, lampID, stateTopicSuffix), 0, true, state)
	token.Wait()
	return token.Error()
}

func PublishDeviceAvailabilityState(client mqtt.Client, devId int, available bool) error {
	state := "offline"
	if available {
		state = "online"
	}
	token := client.Publish(
		fmt.Sprintf("%s%d%s", deviceAvailabilityPrefix, devId, stateTopicSuffix),
		0, true, state,
	)
	token.Wait()
	return token.Error()
}

func PublishAutodiscoveryAndState(client mqtt.Client) error {
	if err := clearLegacyDiscoveryOnce(client); err != nil {
		return err
	}
	if err := PublishHomeLightDiscovery(client); err != nil {
		return err
	}

	lampState := lampController.GetCurLampState()
	for lampID := range lighter.LampNames {
		_, on := lampState.On[lampID]
		if err := PublishLampState(client, lampID, on); err != nil {
			return err
		}
	}

	availability := lampController.GetDeviceAvailability()
	for _, devId := range lampController.KnownDevices() {
		if err := PublishDeviceAvailabilityState(client, devId, availability[devId]); err != nil {
			return err
		}
	}
	return nil
}

func syncAvailabilityDiscovery(client mqtt.Client, announced map[int]bool) {
	known := lampController.KnownDevices()
	if len(known) == 0 {
		return
	}
	needDiscovery := false
	for _, devId := range known {
		if !announced[devId] {
			needDiscovery = true
			break
		}
	}
	if needDiscovery {
		if err := PublishHomeLightDiscovery(client); err != nil {
			log.Printf("Cannot sync availability discovery: %s", err)
			return
		}
		for _, devId := range known {
			announced[devId] = true
		}
	}
	availability := lampController.GetDeviceAvailability()
	for _, devId := range known {
		if err := PublishDeviceAvailabilityState(client, devId, availability[devId]); err != nil {
			log.Printf("Cannot sync availability state for %d: %s", devId, err)
		}
	}
}

func startAvailabilitySync(ctx context.Context, client mqtt.Client) {
	go func() {
		announced := make(map[int]bool)
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				syncAvailabilityDiscovery(client, announced)
			}
		}
	}()
}

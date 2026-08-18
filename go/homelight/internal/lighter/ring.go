package lighter

import (
	"context"
	"log"
	"time"

	"github.com/tarm/serial"
)

const (
	computerDeviceID = 0

	configIDMin = 0
	configIDMax = 6

	ringTick                 = 20 * time.Millisecond
	ringSilenceBaseTicks     = 15
	ringSilencePerIDTicks    = 3
	ringForwardWaitTicks     = 3
	ringBootstrapKnownAfter  = 4 * time.Second
	ringRecvWindow           = 5 * time.Millisecond
)

type ringState int

const (
	ringListen ringState = iota
	ringMaster
	ringWaitForward
)

type Ring struct {
	port         *serial.Port
	deviceID     int
	onDeviceSeen func(devId int)

	state          ringState
	silenceTicks   int
	forwardTicks   int
	targetID       int
	busActivity    bool
	started        time.Time
	changesApplied bool

	session lampSession
	changes LampChanges

	rxBuf  []byte
	useful int
}

func NewRing(port *serial.Port, deviceID int, onDeviceSeen func(int)) *Ring {
	return &Ring{
		port:         port,
		deviceID:     deviceID,
		onDeviceSeen: onDeviceSeen,
		rxBuf:        make([]byte, widePacketBytes*2),
		started:      time.Now(),
	}
}

func (r *Ring) SetChanges(changes LampChanges) {
	r.changes = changes
}

func (r *Ring) LampState() *LampState {
	return r.session.cloneState()
}

func (r *Ring) ChangesApplied() bool {
	return r.changesApplied
}

func silenceThreshold(deviceID int) int {
	return ringSilenceBaseTicks + ringSilencePerIDTicks*deviceID
}

func nextDeviceID(id int) int {
	id++
	if id >= configIDMax {
		return configIDMin
	}
	return id
}

func advanceTarget(target, self int) int {
	for {
		target = nextDeviceID(target)
		if target != self {
			return target
		}
	}
}

func (r *Ring) Run(ctx context.Context) error {
	ticker := time.NewTicker(ringTick)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if err := r.step(ctx); err != nil {
				return err
			}
		}
	}
}

func (r *Ring) RunStep(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	r.changesApplied = false

	packet, err := readPacket(r.port, r.rxBuf, &r.useful, time.Now().Add(ringRecvWindow))
	if err != nil {
		return err
	}
	if packet != nil {
		r.handlePacket(packet)
	} else {
		r.onSilenceTick()
		r.checkElection()
		r.checkForwardTimeout()
	}

	r.checkBootstrapKnown()
	if r.state == ringMaster {
		if err := r.sendMaster(); err != nil {
			return err
		}
	}
	return nil
}

func (r *Ring) step(ctx context.Context) error {
	if err := r.RunStep(ctx); err != nil {
		return err
	}
	return nil
}

func (r *Ring) markBusActivity() {
	r.busActivity = true
	r.silenceTicks = 0
}

func (r *Ring) onSilenceTick() {
	if r.busActivity {
		r.busActivity = false
		return
	}
	if r.state == ringListen {
		r.silenceTicks++
	}
	if r.state == ringWaitForward && r.forwardTicks > 0 {
		r.forwardTicks--
	}
}

func (r *Ring) checkElection() {
	if r.state != ringListen {
		return
	}
	if r.silenceTicks < silenceThreshold(r.deviceID) {
		return
	}
	r.state = ringMaster
	r.targetID = nextDeviceID(r.deviceID)
}

func (r *Ring) checkForwardTimeout() {
	if r.state != ringWaitForward || r.forwardTicks > 0 {
		return
	}
	r.targetID = advanceTarget(r.targetID, r.deviceID)
	r.state = ringMaster
}

func (r *Ring) checkBootstrapKnown() {
	if r.session.lampStateKnown {
		return
	}
	if time.Since(r.started) < ringBootstrapKnownAfter {
		return
	}
	r.session.lampStateKnown = true
}

func (r *Ring) releaseToken() {
	r.state = ringListen
	r.silenceTicks = 0
}

func (r *Ring) handlePacket(packet []byte) {
	if packet[posDir] != 0xAB {
		return
	}
	target := int(packet[posDevId])
	r.markBusActivity()
	if r.onDeviceSeen != nil {
		r.onDeviceSeen(target)
	}

	if target == r.deviceID {
		r.handleForUs(packet)
		return
	}

	if r.state == ringWaitForward {
		r.releaseToken()
	}
}

func (r *Ring) handleForUs(packet []byte) {
	r.session.absorbIncoming(packet)
	work := append([]byte(nil), packet...)
	if r.session.applyChanges(work, r.changes) {
		r.session.syncFromPacket(work)
	}
	r.state = ringMaster
	r.targetID = nextDeviceID(r.deviceID)
}

func (r *Ring) sendMaster() error {
	work, err := buildPacket(0xAB, r.targetID, r.session.lampStateKnown, r.session.lampState)
	if err != nil {
		return err
	}
	if r.session.applyChanges(work, r.changes) {
		r.session.syncFromPacket(work)
		r.changesApplied = true
	}
	if r.session.lampStateKnown {
		work[posLampStateKnown] = 0x01
	}
	c := NewCRC()
	c.PushBytes(work[0 : len(work)-2])
	val := c.Value()
	work[posCRCH] = byte((val / 0x100) & 0xFF)
	work[posCRCL] = byte(val & 0xFF)

	if err := writePacket(r.port, work); err != nil {
		return err
	}
	log.Printf("Ring: sent 0xAB to %d", r.targetID)
	r.state = ringWaitForward
	r.forwardTicks = ringForwardWaitTicks
	return nil
}

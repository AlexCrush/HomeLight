package lighter

import (
	"errors"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/tarm/serial"
)

const widePacketBytes = 26
const packetBytes = widePacketBytes / 2
const posDir = 0
const posDevId = 1
const posLampStateKnown = 2
const posLamp = 3
const posCRCH = packetBytes - 1
const posCRCL = packetBytes - 2
const communicateTimeout = time.Second * 5
const lampCount = 64

// masterDeviceId is the RS485 bus master; its address is not in request packets,
// so we mark it available whenever we see any master→slave frame (0xAB).
const masterDeviceId = 4

type LampAction int

var LampNames = map[int]string{
	0:  "Boileroom",
	1:  "Toilet1",
	2:  "StairsBra",
	3:  "Facade",
	4:  "Living",
	5:  "Kitchen",
	6:  "Hallway",
	7:  "Storage",
	8:  "Hall1",
	9:  "Toilet2",
	10: "Sauna",
	11: "Bedroom",
	12: "Childroom",
	13: "Cabinet",
	14: "Hall2",
	15: "Canopy1",
	16: "Terrace1",
	17: "Terrace2",
	18: "Toilet2Mirror",
	19: "TowelDrier",
	20: "Vent1",
	21: "Vent2",
	22: "KitchenLight",
	23: "LivingFloor",
	24: "Toilet1Mirror",
	25: "BedroomBra1",
	26: "BedroomBra2",
	27: "ChildroomBra1",
	28: "ChildroomBra2",
	29: "Canopy2",
	30: "CabinetLight",
	31: "BedroomLight",
	32: "ChildroomMirror",
	33: "BedroomWardrobe",
	// 34 unused
	35: "BedroomWardrobeSens",
	36: "RadioProjectorScreenUp",
	37: "RadioProjectorScreenDown",
	38: "RadioProjectorScreenStop",
	39: "CabinetSens",
	40: "ChildroomMirrorSens",
	41: "Toilet2HintOn",
}

const (
	LAMP_ACTION_NONE = iota
	LAMP_ACTION_ON
	LAMP_ACTION_OFF
	LAMP_ACTION_TOGGLE
)

type LampChanges map[int]LampAction

type LampState struct {
	On map[int]struct{}
}

func (s LampState) Clone() *LampState {
	on := make(map[int]struct{}, len(s.On))
	for lampID := range s.On {
		on[lampID] = struct{}{}
	}
	return &LampState{On: on}
}

type Lighter struct {
	port    *serial.Port
	data    []byte
	useful  int
	replied bool

	changes          LampChanges
	lampStateInitial *LampState
	lampState        LampState
	lampStateKnown   bool
	onDeviceSeen     func(devId int)
}

func (l *Lighter) communicate() error {
	started := time.Now()
	l.replied = false
	for !l.replied {
		if time.Since(started) > communicateTimeout {
			return fmt.Errorf("communication failed: timeout")
		}
		n, err := l.port.Read(l.data[l.useful:])
		if err != nil {
			if errors.Is(err, io.EOF) {
				n = 0
			} else {
				return err
			}
		}
		l.useful += n
		skip := 0
		for ; skip < l.useful; skip++ {
			if l.data[skip]&0xF0 == 0xF0 {
				break
			}
		}
		if skip != 0 && skip < l.useful {
			for i := 0; i < l.useful-skip; i++ {
				l.data[i] = l.data[i+skip]
			}
		}
		if skip != 0 {
			log.Printf("WARN: Skip %d: %s", skip, bytesToString(l.data))
		}
		l.useful -= skip
		if l.useful == widePacketBytes {
			l.processData()
			l.useful = 0
		}
	}
	return nil
}

func (l *Lighter) processData() {
	shortData, err := l.decode()
	if err != nil {
		log.Printf("Decode error: %s: %s", bytesToString(l.data), err)
		return
	}
	if shortData[posDir] == 0xAB {
		log.Printf("Master to %d", shortData[posDevId])
		if l.onDeviceSeen != nil {
			l.onDeviceSeen(masterDeviceId)
		}
	}
	if shortData[posDir] == 0xCD {
		devId := int(shortData[posDevId])
		log.Printf("Slave %d to master", devId)
		if l.onDeviceSeen != nil {
			l.onDeviceSeen(devId)
		}
		return
	}
	if shortData[posDir] != 0xAB {
		return
	}
	if shortData[posDevId] != 0 {
		return
	}
	if err := l.reply(shortData); err != nil {
		log.Printf("Reply error: %s", err)
		return
	}
}

func (l *Lighter) decode() ([]byte, error) {
	shortData, err := shrinkData(l.data)
	if err != nil {
		return nil, err
	}
	c := NewCRC()
	c.PushBytes(shortData[0 : len(shortData)-2])
	expVal := c.Value()
	actVal := uint16(shortData[posCRCH])*0x100 + uint16(shortData[posCRCL])
	if expVal != actVal {
		return nil, fmt.Errorf("CRC mismatch: Exp %X, Act %X", expVal, actVal)
	}
	return shortData, nil
}

func changeLamp(data []byte, initialState LampState, lamp int, action LampAction) bool {
	if action == LAMP_ACTION_TOGGLE {
		if _, on := initialState.On[lamp]; on {
			return changeLamp(data, initialState, lamp, LAMP_ACTION_OFF)
		} else {
			return changeLamp(data, initialState, lamp, LAMP_ACTION_ON)
		}
	}
	offset, mask := lampPosAndMask(lamp)
	switch action {
	case LAMP_ACTION_OFF:
		if data[posLamp+offset]&mask != 0 {
			data[posLamp+offset] &= 0xFF - mask
			return true
		}
	case LAMP_ACTION_ON:
		if data[posLamp+offset]&mask == 0 {
			data[posLamp+offset] |= mask
			return true
		}
	}
	return false
}

func parseLampState(data []byte) LampState {
	lampState := LampState{
		On: make(map[int]struct{}),
	}
	for lamp := 0; lamp < lampCount; lamp++ {
		offset, mask := lampPosAndMask(lamp)
		if data[posLamp+offset]&mask != 0 {
			lampState.On[lamp] = struct{}{}
		}
	}
	return lampState
}

// writeLampState puts known lamp bits into the packet. Returns true if any bit differed.
func writeLampState(data []byte, state LampState) bool {
	changed := false
	for lamp := 0; lamp < lampCount; lamp++ {
		offset, mask := lampPosAndMask(lamp)
		_, on := state.On[lamp]
		have := data[posLamp+offset]&mask != 0
		if on == have {
			continue
		}
		changed = true
		if on {
			data[posLamp+offset] |= mask
		} else {
			data[posLamp+offset] &= 0xFF - mask
		}
	}
	return changed
}

func (l *Lighter) reply(data []byte) error {
	incomingKnown := data[posLampStateKnown] == 0x01
	wasKnown := l.lampStateKnown
	if incomingKnown {
		l.lampStateKnown = true
	}
	data[posDir] = 0xCD

	changed := false
	if incomingKnown || !wasKnown {
		lampState := parseLampState(data)
		l.lampState = lampState
		if l.lampStateInitial == nil {
			l.lampStateInitial = &lampState
		}
	} else {
		// Master restarted / forgot lamps — restore from our known state.
		if writeLampState(data, l.lampState) {
			changed = true
		}
	}

	if l.changes != nil && l.lampStateInitial != nil {
		for lamp, action := range l.changes {
			if changeLamp(data, *l.lampStateInitial, lamp, action) {
				changed = true
			}
		}
	}
	if l.lampStateKnown {
		data[posLampStateKnown] = 0x01
	} else {
		data[posLampStateKnown] = 0x00
	}
	c := NewCRC()
	c.PushBytes(data[0 : len(data)-2])
	expVal := c.Value()
	data[posCRCH] = byte((expVal / 0x100) & 0xFF)
	data[posCRCL] = byte((expVal) & 0xFF)
	eData := widenData(data)
	n, err := l.port.Write(eData)
	if err != nil {
		return fmt.Errorf("write error: %s", err)
	}
	if n != len(eData) {
		return fmt.Errorf("written %d instead of %d", n, len(eData))
	}
	if changed {
		log.Printf("Replied, but some changes happened, need retry")
	} else {
		log.Printf("Replied and no changes")
		l.replied = true
	}
	return nil
}

func OpenPort() (*serial.Port, error) {
	return serial.OpenPort(&serial.Config{
		//Name:        "/dev/serial/by-id/usb-FTDI_FT232R_USB_UART_A50285BI-if00-port0",
		Name:        "/dev/ttyRS485-1", // Wirenboard
		Baud:        57600,
		Parity:      serial.ParityNone,
		StopBits:    serial.Stop2,
		ReadTimeout: time.Second,
	})
}

func Communicate(port *serial.Port, lch LampChanges) (*LampState, error) {
	l := Lighter{
		port:    port,
		data:    make([]byte, widePacketBytes),
		changes: lch,
	}
	if err := l.communicate(); err != nil {
		return nil, err
	}
	return &l.lampState, nil
}

func ControlLampsOnce(lch LampChanges) (*LampState, error) {
	port, err := OpenPort()
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = port.Close()
	}()
	return Communicate(port, lch)
}

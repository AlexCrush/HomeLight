package lighter

import (
	"fmt"
	"io"
	"time"

	"github.com/tarm/serial"
)

type lampSession struct {
	lampState        LampState
	lampStateKnown   bool
	lampStateInitial *LampState
}

func (s *lampSession) cloneState() *LampState {
	return s.lampState.Clone()
}

func (s *lampSession) absorbIncoming(data []byte) {
	incomingKnown := data[posLampStateKnown] == 0x01
	wasKnown := s.lampStateKnown
	if incomingKnown {
		s.lampStateKnown = true
	}
	if incomingKnown || !wasKnown {
		s.lampState = parseLampState(data)
		if s.lampStateInitial == nil {
			initial := s.lampState.Clone()
			s.lampStateInitial = initial
		}
		return
	}
	writeLampState(data, s.lampState)
}

func (s *lampSession) applyChanges(data []byte, changes LampChanges) bool {
	if changes == nil || s.lampStateInitial == nil {
		return false
	}
	changed := false
	for lamp, action := range changes {
		if changeLamp(data, *s.lampStateInitial, lamp, action) {
			changed = true
		}
	}
	return changed
}

func (s *lampSession) syncFromPacket(data []byte) {
	s.lampState = parseLampState(data)
}

func buildPacket(dir byte, targetID int, known bool, lampState LampState) ([]byte, error) {
	data := make([]byte, packetBytes)
	data[posDir] = dir
	data[posDevId] = byte(targetID)
	if known {
		data[posLampStateKnown] = 0x01
	} else {
		data[posLampStateKnown] = 0x00
	}
	writeLampState(data, lampState)
	c := NewCRC()
	c.PushBytes(data[0 : len(data)-2])
	val := c.Value()
	data[posCRCH] = byte((val / 0x100) & 0xFF)
	data[posCRCL] = byte(val & 0xFF)
	return data, nil
}

func writePacket(port *serial.Port, data []byte) error {
	wide := widenData(data)
	n, err := port.Write(wide)
	if err != nil {
		return err
	}
	if n != len(wide) {
		return fmt.Errorf("written %d instead of %d", n, len(wide))
	}
	return nil
}

func readPacket(port *serial.Port, buf []byte, useful *int, deadline time.Time) ([]byte, error) {
	for time.Now().Before(deadline) {
		n, err := port.Read(buf[*useful:])
		if err != nil && !isTimeout(err) {
			return nil, err
		}
		*useful += n
		skip := 0
		for ; skip < *useful; skip++ {
			if buf[skip]&0xF0 == 0xF0 {
				break
			}
		}
		if skip != 0 && skip < *useful {
			copy(buf, buf[skip:*useful])
		}
		*useful -= skip
		if *useful < widePacketBytes {
			continue
		}
		if *useful > widePacketBytes {
			copy(buf, buf[*useful-widePacketBytes:*useful])
			*useful = widePacketBytes
		}
		shortData, err := shrinkData(buf[:widePacketBytes])
		if err != nil {
			*useful = 0
			continue
		}
		c := NewCRC()
		c.PushBytes(shortData[0 : len(shortData)-2])
		expVal := c.Value()
		actVal := uint16(shortData[posCRCH])*0x100 + uint16(shortData[posCRCL])
		if expVal != actVal {
			*useful = 0
			continue
		}
		*useful = 0
		return shortData, nil
	}
	return nil, nil
}

func isTimeout(err error) bool {
	if err == nil {
		return false
	}
	return err == io.EOF
}

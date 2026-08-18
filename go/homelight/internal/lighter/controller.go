package lighter

import (
	"context"
	"log"
	"sync"

	"github.com/tarm/serial"
)

type Controller struct {
	ctx context.Context

	onLampStateChange          OnLampStateChange
	onDeviceAvailabilityChange OnDeviceAvailabilityChange

	changesCh      chan lampChange
	deviceTracker  *DeviceTracker

	stateMu  sync.RWMutex
	curState *LampState
}

type lampChange struct {
	lampID int
	action LampAction
}

type OnLampStateChange func(lampID int, on bool)

func NewController(
	ctx context.Context,
	onLampStateChange OnLampStateChange,
	onDeviceAvailabilityChange OnDeviceAvailabilityChange,
) *Controller {
	c := &Controller{
		ctx:                        ctx,
		onLampStateChange:          onLampStateChange,
		onDeviceAvailabilityChange: onDeviceAvailabilityChange,
		changesCh:                  make(chan lampChange, 64),
	}
	c.deviceTracker = NewDeviceTracker(onDeviceAvailabilityChange)
	return c
}

func (c *Controller) Run() error {
	port, err := OpenPort()
	if err != nil {
		return err
	}

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		c.deviceTracker.Run(c.ctx)
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer func() { _ = port.Close() }()
		c.runWorker(port)
	}()

	<-c.ctx.Done()
	_ = port.Close()
	wg.Wait()
	return c.ctx.Err()
}

func (c *Controller) runWorker(port *serial.Port) {
	l := &Lighter{
		port:         port,
		data:         make([]byte, widePacketBytes),
		onDeviceSeen: c.deviceTracker.See,
	}
	pending := make(LampChanges)

	for {
		if c.ctx.Err() != nil {
			return
		}
		c.drainChanges(pending)
		l.changes = pending
		if err := l.communicate(); err != nil {
			if c.ctx.Err() != nil {
				return
			}
			log.Printf("Error while communicate: %s", err)
			continue
		}
		prevState := c.getCurState()
		newState := l.lampState.Clone()
		c.setCurState(newState)
		c.compareAndNotifyChanges(prevState, newState)
		pending = make(LampChanges)
	}
}

func (c *Controller) drainChanges(pending LampChanges) {
	for {
		select {
		case ch := <-c.changesCh:
			pending[ch.lampID] = ch.action
		default:
			return
		}
	}
}

func (c *Controller) GetCurLampState() *LampState {
	c.stateMu.RLock()
	defer c.stateMu.RUnlock()
	if c.curState == nil {
		return &LampState{On: make(map[int]struct{})}
	}
	return c.curState
}

func (c *Controller) KnownDevices() []int {
	return c.deviceTracker.KnownDevices()
}

func (c *Controller) GetDeviceAvailability() map[int]bool {
	return c.deviceTracker.Snapshot()
}

func (c *Controller) ChangeLamp(lampID int, on bool) {
	var action LampAction
	if on {
		action = LAMP_ACTION_ON
	} else {
		action = LAMP_ACTION_OFF
	}
	select {
	case <-c.ctx.Done():
	case c.changesCh <- lampChange{lampID: lampID, action: action}:
	}
}

func (c *Controller) getCurState() *LampState {
	c.stateMu.RLock()
	defer c.stateMu.RUnlock()
	return c.curState
}

func (c *Controller) setCurState(state *LampState) {
	c.stateMu.Lock()
	defer c.stateMu.Unlock()
	c.curState = state
}

func (c *Controller) compareAndNotifyChanges(prevState, newState *LampState) {
	if prevState != nil {
		for lampID := range prevState.On {
			if _, on := newState.On[lampID]; !on {
				c.onLampStateChange(lampID, false)
			}
		}
	}
	for lampID := range newState.On {
		if prevState != nil {
			if _, on := prevState.On[lampID]; !on {
				c.onLampStateChange(lampID, true)
			}
		} else {
			c.onLampStateChange(lampID, true)
		}
	}
}

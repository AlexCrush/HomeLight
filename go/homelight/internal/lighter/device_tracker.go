package lighter

import (
	"context"
	"log"
	"sync"
	"time"
)

const deviceAvailabilityWindow = 3 * time.Second

type OnDeviceAvailabilityChange func(devId int, available bool, firstSeen bool)

type DeviceTracker struct {
	onChange OnDeviceAvailabilityChange

	mu        sync.Mutex
	lastSeen  map[int]time.Time
	available map[int]bool
}

func NewDeviceTracker(onChange OnDeviceAvailabilityChange) *DeviceTracker {
	return &DeviceTracker{
		onChange:  onChange,
		lastSeen:  make(map[int]time.Time),
		available: make(map[int]bool),
	}
}

// See marks a bus device available. Our own address (0) is ignored.
func (t *DeviceTracker) See(devId int) {
	if devId == 0 {
		return
	}
	now := time.Now()

	t.mu.Lock()
	firstSeen := t.lastSeen[devId].IsZero()
	t.lastSeen[devId] = now
	wasAvailable := t.available[devId]
	t.available[devId] = true
	t.mu.Unlock()

	if !wasAvailable {
		log.Printf("Device %d became available (firstSeen=%v)", devId, firstSeen)
		go t.onChange(devId, true, firstSeen)
	}
}

func (t *DeviceTracker) Run(ctx context.Context) {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			t.expire()
		}
	}
}

func (t *DeviceTracker) expire() {
	now := time.Now()
	var expired []int

	t.mu.Lock()
	for devId, seen := range t.lastSeen {
		if !t.available[devId] {
			continue
		}
		if now.Sub(seen) > deviceAvailabilityWindow {
			t.available[devId] = false
			expired = append(expired, devId)
		}
	}
	t.mu.Unlock()

	for _, devId := range expired {
		log.Printf("Device %d became unavailable", devId)
		go t.onChange(devId, false, false)
	}
}

func (t *DeviceTracker) Snapshot() map[int]bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	snap := make(map[int]bool, len(t.available))
	for devId, avail := range t.available {
		snap[devId] = avail
	}
	return snap
}

func (t *DeviceTracker) KnownDevices() []int {
	t.mu.Lock()
	defer t.mu.Unlock()
	devices := make([]int, 0, len(t.lastSeen))
	for devId := range t.lastSeen {
		devices = append(devices, devId)
	}
	return devices
}

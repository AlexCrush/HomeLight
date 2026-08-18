package main

import (
	"homelight/internal/lighter"
	"log"
)

func main() {
	changes := lighter.LampChanges{
		3:  lighter.LAMP_ACTION_TOGGLE,
		15: lighter.LAMP_ACTION_TOGGLE,
	}
	state, err := lighter.ControlLampsOnce(changes)
	if err != nil {
		log.Fatalf("Cannot control lamps: %s", err)
	}
	log.Printf("On lamps: %v", state.On)
}

package main

import (
	"flag"
	"homelight/internal/lighter"
	"log"
)

func main() {
	cfg := lighter.DefaultConfig()
	cfg.RegisterFlags(flag.CommandLine)
	flag.Parse()
	cfg.ApplyEnv()

	changes := lighter.LampChanges{
		3:  lighter.LAMP_ACTION_TOGGLE,
		15: lighter.LAMP_ACTION_TOGGLE,
	}
	state, err := lighter.ControlLampsOnce(cfg, changes)
	if err != nil {
		log.Fatalf("Cannot control lamps: %s", err)
	}
	log.Printf("On lamps: %v", state.On)
}

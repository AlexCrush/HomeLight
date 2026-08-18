package lighter

import (
	"flag"
	"os"
	"strconv"
	"strings"
)

const (
	defaultSerialPort = "/dev/ttyRS485-1"
	envRingProtocol   = "HOMELIGHT_RING_PROTOCOL"
)

// Config holds runtime options for the RS-485 bridge.
type Config struct {
	SerialPort   string
	RingProtocol bool
}

func DefaultConfig() Config {
	return Config{
		SerialPort: defaultSerialPort,
	}
}

func (c *Config) RegisterFlags(fs *flag.FlagSet) {
	fs.StringVar(&c.SerialPort, "serial", c.SerialPort, "RS-485 serial device path")
	fs.BoolVar(&c.RingProtocol, "ring-protocol", c.RingProtocol, "use ring token RS-485 protocol instead of fixed master")
}

func (c *Config) ApplyEnv() {
	if v := strings.TrimSpace(os.Getenv(envRingProtocol)); v != "" {
		enabled, err := strconv.ParseBool(v)
		if err != nil {
			return
		}
		c.RingProtocol = enabled
	}
	if v := strings.TrimSpace(os.Getenv("HOMELIGHT_SERIAL")); v != "" {
		c.SerialPort = v
	}
}

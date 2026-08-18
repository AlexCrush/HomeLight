package main

import "fmt"

const topicPrefix = "homelight/switch"
const setTopicSuffix = "/set"
const stateTopicSuffix = "/state"

const deviceAvailabilityPrefix = "homelight/deviceAvailability"

var deviceNames = map[int]string{
	1: "Котельная",
	2: "Кладовка",
	3: "СУ2",
	4: "Гостиная",
	5: "СУ2-управление",
	6: "Детская",
}

func deviceAvailabilityName(devId int) string {
	if name, ok := deviceNames[devId]; ok {
		return name
	}
	return fmt.Sprintf("deviceAvailability%d", devId)
}

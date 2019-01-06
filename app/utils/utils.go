package utils

import (
	"log"
	"time"
)

func TimeTrack(start time.Time, name interface{}) {
	elapsed := time.Since(start)
	log.Printf("%v took %s", name, elapsed)
}

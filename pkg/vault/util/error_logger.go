package util

import "log"

func LogWriteErr(n int, err error) {
	if err != nil {
		log.Printf("Write failed: %v\n", err)
	}
}

func LogErr(err error) {
	if err != nil {
		log.Println(err)
	}
}

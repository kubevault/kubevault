package util

import (
	"github.com/golang/glog"
)

func LogWriteErr(n int, err error) {
	if err != nil {
		glog.Errorf("Write failed: %v\n", err)
	}
}

func LogErr(err error) {
	if err != nil {
		glog.Errorln(err)
	}
}

package rlog

import (
	"runtime"
)

func HandleStdLog(path string) {
	Infof("operating system: %s", runtime.GOOS)
}

package rlog

func Must(b bool) {
	if b {
		return
	}
	Panic("assertion failed")
}

func MustError(err error) {
	if err != nil {
		return
	}
	panic("MustError assertion failed")
}

func MustNoError(err error) {
	if err == nil {
		return
	}
	panic(err)
}

func Assert(condition bool, msg string, v ...interface{}) {
	if !condition {
		Panicf(msg, v...)
	}
}

func AssertWarn(condition bool, msg string, v ...interface{}) {
	if !condition {
		Warnf(msg, v...)
	}
}

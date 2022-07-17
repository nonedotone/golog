package golog

import (
	"testing"
	"time"
)

func TestLog(t *testing.T) {
	Log().Level("debug")
	Debug("Debug", "log")
	Info("Info", "log")
	Error("Error", "log")
	Warn("Warn", "log")
	Debugf("%s-%s\n", "Debugf", "log")
	Infof("%s-%s\n", "Infof", "log")
	Errorf("%s-%s\n", "Errorf", "log")
	Warnf("%s-%s\n", "Warnf", "log")
	//	Fatal("Fatal")
	//	Fatalf("%s-%s\n", "Fatalf", "log")
}

func TestFile(t *testing.T) {
	Log().Level("debug").LogFile("./tmp/test.log")
	Info("Info")
	Error("Error")
	Debug("Debug")
	Warn("Warn")
}

func TestSizeRolling(t *testing.T) {
	Log().Level("debug").LogFile("./tmp/test.log").Rolling("size", MinSizeInterval)
	for i := 0; i < 100; i++ {
		Info("Info ", i, " 123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890")
	}
	t.Log("complete")
}

func TestTimeRolling(t *testing.T) {
	Log().Level("debug").LogFile("./tmp/test.log").Rolling("time", MinTimeInterval)
	for i := 0; i < 15; i++ {
		Info("Info", "a", i)
		time.Sleep(time.Second)
	}
	t.Log("complete")
}

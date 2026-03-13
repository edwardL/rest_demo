package gdbtmp

import (
	"testing"
)

func TestSetDefLog(t *testing.T) {
	defLog.SetLevel(InfoLogLevel)
	defLog.CtxInfo(nil, "Info")
	defLog.CtxInfof(nil, "Infof")
	defLog.CtxDebug(nil, "Debug")
	defLog.CtxDebugf(nil, "Debugf")
	defLog.CtxError(nil, "Error")
	defLog.CtxErrorf(nil, "Errorf")
}

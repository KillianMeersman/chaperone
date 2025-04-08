package log

import (
	"testing"
)

func TestGetLogLevelUndefined(t *testing.T) {
	logLevel := TranslateLogLevel("INFO")
	if logLevel != INFO {
		t.Errorf("expected default LOG_LEVEL of INFO, got '%d'", logLevel)
	}
}

func TestGetLogLevelDebug(t *testing.T) {
	logLevel := TranslateLogLevel("DEBUG")
	if logLevel != DEBUG {
		t.Errorf("expected LOG_LEVEL of %d, got %d", DEBUG, logLevel)
	}
}

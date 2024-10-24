package log

import (
	"encoding/json"
	"fmt"
	"strings"
)

func ShellColoredLevels(level LogLevel, message string, fields map[string]string) []byte {
	fieldString := strings.Builder{}

	for key, value := range fields {
		fieldString.WriteString(fmt.Sprintf("%s=%s, ", key, value))
	}

	levelStr := "???"
	levelFormatting := []ShellFormatting{Reset}
	switch level {
	case TRACE:
		levelStr = "TRACE"
		levelFormatting = []ShellFormatting{FgWhite, Faint}
	case DEBUG:
		levelStr = "DEBUG"
		levelFormatting = []ShellFormatting{FgWhite, Faint}
	case INFO:
		levelStr = "INFO"
		levelFormatting = []ShellFormatting{FgBlue}
	case WARNING:
		levelStr = "WARNING"
		levelFormatting = []ShellFormatting{FgYellow, Bold}
	case ERROR:
		levelStr = "ERROR"
		levelFormatting = []ShellFormatting{FgHiRed, Bold}
	case FATAL:
		levelStr = "FATAL"
		levelFormatting = []ShellFormatting{FgHiMagenta, Bold}
	}

	return []byte(TabulateRow().Field(levelStr, 5, levelFormatting...).Field("|", 1).Field(message, -1).Field(fieldString.String(), -1).String())
}

type jsonMessage struct {
	Message string            `json:"message"`
	Fields  map[string]string `json:"fields"`
}

func JSON(level LogLevel, message string, fields map[string]string) []byte {
	messageStruct := jsonMessage{
		message,
		fields,
	}
	data, err := json.Marshal(messageStruct)
	if err != nil {
		panic(err)
	}

	return data
}

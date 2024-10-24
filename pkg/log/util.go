package log

import (
	"bufio"
	"io"
)

// Stream every newline from reader as a new logging entry.
// Returns the result of `reader.Err()`.
func StreamReaderNewLines(logger func(message string, fields ...string), reader io.Reader) error {
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		m := scanner.Text()
		logger(m)
	}

	return scanner.Err()
}

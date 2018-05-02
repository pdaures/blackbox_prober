package pingers

import (
	"bytes"
	"errors"
	"fmt"
	"io"
)

var (
	// ErrUnsupportedScheme will returned if no pinger function exists for given scheme
	ErrUnsupportedScheme = errors.New("Scheme not supported")
)

func readSize(r io.Reader) (int, error) {
	size := 0
	buf := make([]byte, bytes.MinRead) // Since we discard the buffer, alloc only once
	for {
		n, err := r.Read(buf)
		size += n
		if err != nil {
			if err == io.EOF {
				return size, nil
			}
			return size, err
		}
	}
}

// Ping executes the matching pinger function for the url.
// If no pinger function can be found, it return ErrUnsupportedScheme.
func Ping(target *Target, reporter MetricReporter) error {
	switch target.Rule.Type {
	case "http":
		return pingerHTTP(target.URL, reporter, target.Rule)
	case "tcp":
		return pingerTCP(target.URL, reporter, target.Rule)
	case "icmp":
		return pingerICMP(target.URL, reporter, target.Rule)
	default:
		return fmt.Errorf("no handler for rule type %s", target.Rule.Type)
	}
	return nil
}

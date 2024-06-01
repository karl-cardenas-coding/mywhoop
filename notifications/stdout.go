package notifications

import "net/http"

// Stdout is a struct that contains the configuration for the sending messages to stdout.
func NewStdout() *Stdout {
	return &Stdout{}
}

func (s *Stdout) SetUp() error {
	return nil
}

func (s *Stdout) Publish(client *http.Client, data []byte, event string) error {
	return nil
}

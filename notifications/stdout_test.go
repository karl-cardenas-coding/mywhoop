package notifications

import (
	"testing"
)

func TestNewStdout(t *testing.T) {
	expect := &Stdout{}

	if got := NewStdout(); got != expect {
		t.Errorf("NewStdout() = %v, want %v", got, expect)
	}
}

func TestStdout_SetUp(t *testing.T) {
	s := NewStdout()

	if err := s.SetUp(); err != nil {
		t.Errorf("SetUp() error = %v, want nil", err)
	}
}

func TestStdout_Publish(t *testing.T) {
	s := NewStdout()

	if err := s.Publish(nil, nil, ""); err != nil {
		t.Errorf("Publish() error = %v, want nil", err)
	}
}

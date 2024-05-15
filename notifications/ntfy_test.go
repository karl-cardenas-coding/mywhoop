package notifications

import "testing"

func TestSetup(t *testing.T) {
	ntfy := Ntfy{}
	err := ntfy.SetUp()
	if err != nil {
		t.Errorf("Error setting up Ntfy service: %v", err)
	}
}

func TestSend(t *testing.T) {
	ntfy := Ntfy{}
	err := ntfy.Send([]byte("test"))
	if err != nil {
		t.Errorf("Error sending Ntfy notification: %v", err)
	}
}

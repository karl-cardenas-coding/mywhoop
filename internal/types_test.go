package internal

import "testing"

func TestEventString(t *testing.T) {

	var custom Event = "custom"

	tests := []struct {
		id     int
		event  Event
		result string
	}{
		{1, EventErrors, "errors"},
		{2, EventSuccess, "success"},
		{3, EventAll, "all"},
		{4, custom, "custom"},
	}

	for index, test := range tests {
		test.id = index + 1
		result := test.event.String()
		if result != test.result {
			t.Errorf("Test %d: Expected %s, got %s", test.id, test.result, result)
		}
	}
}

func TestEventFromString(t *testing.T) {

	tests := []struct {
		id     int
		event  string
		result string
	}{
		{1, "errors", "errors"},
		{2, "success", "success"},
		{3, "all", "all"},
		{4, "custom", "errors"},
	}

	for index, test := range tests {
		test.id = index + 1
		result := EventFromString(test.event)
		if result.String() != test.result {
			t.Errorf("Test %d: Expected %s, got %s", test.id, test.result, result)
		}
	}
}

package internal

import (
	"encoding/json"
	"os"
	"testing"
)

func TestExportData(t *testing.T) {
	// Test case 1: Successful export
	user := &User{
		UserData: UserData{
			UserID:    123456789,
			Email:     "john.doe@gmail.com",
			FirstName: "John",
			LastName:  "Doe",
		},
	}
	err := user.ExportDataToFile("data/user.json")
	if err != nil {
		t.Errorf("Expected nil error, got: %v", err)
	}

	// Verify that the file was created and contains the expected data
	data, err := os.ReadFile("data/user.json")
	if err != nil {
		t.Errorf("Failed to read file: %v", err)
	}
	expectedData, _ := json.MarshalIndent(user, "", "  ")
	if string(data) != string(expectedData) {
		t.Errorf("Expected file content: %s, got: %s", string(expectedData), string(data))
	}

	// Clean up the file
	err = os.Remove("data/user.json")
	if err != nil {
		t.Errorf("Failed to remove file: %v", err)
	}

	// Test case 2: Error when marshaling data
	user = &User{} // Invalid user object
	err = user.ExportDataToFile("data/user.json")
	if err == nil {
		t.Error("Expected non-nil error, got nil")
	}

	// Test case 3: Error when writing to file
	user2 := &User{
		UserData: UserData{
			UserID:    123456789,
			Email:     "john.doe@gmail.com",
			FirstName: "John",
			LastName:  "Doe",
		},
		UserMesaurements: UserMesaurements{
			HeightMeter:    1.778,
			WeightKilogram: 58.9,
			MaxHeartRate:   125,
		},
	}

	// Remove the data directory if it exists
	if _, err := os.Stat("data"); err == nil {
		err = os.RemoveAll("data")
		if err != nil {
			t.Errorf("Failed to remove directory: %v", err)
		}
	}
	// Create a read-only directory to simulate a write error
	err = os.Mkdir("data", 0444)
	if err != nil {
		t.Errorf("Failed to create read-only directory: %v", err)
	}
	err = user2.ExportDataToFile("data/user.json")
	if err == nil {
		t.Error("Expected non-nil error, got nil")
	}
	// Clean up the directory
	err = os.Chmod("data", 0755)
	if err != nil {
		t.Errorf("Failed to change directory permissions: %v", err)
	}
	err = os.Remove("data")
	if err != nil {
		t.Errorf("Failed to remove directory: %v", err)
	}
}

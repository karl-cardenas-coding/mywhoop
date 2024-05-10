// Copyright (c) karl-cardenas-coding
// SPDX-License-Identifier: Apache-2.0

package export

import (
	"encoding/json"
	"log/slog"
	"os"
	"os/exec"
	"path"
	"testing"
)

type User struct {
	UserData         UserData         `json:"user_data"`
	UserMesaurements UserMesaurements `json:"user_mesaurements"`
}

type UserMesaurements struct {
	HeightMeter    float64 `json:"height_meter"`
	WeightKilogram float64 `json:"weight_kilogram"`
	MaxHeartRate   int     `json:"max_heart_rate"`
}

type UserData struct {
	UserID    int    `json:"user_id"`
	Email     string `json:"email"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

func TestGenerateName(t *testing.T) {

	type test struct {
		testCase int
		file     FileExport
		want     string
	}

	tests := []test{
		{
			file: FileExport{
				FileNamePrefix: "test",
				FileName:       "user",
				FileType:       "json",
			},
			want: "test_user.json",
		},
		{
			file: FileExport{
				FileNamePrefix: "",
				FileName:       "user",
				FileType:       "json",
			},
			want: "user.json",
		},
		{
			file: FileExport{
				FileNamePrefix: "",
				FileName:       "",
				FileType:       "",
			},
			want: ".",
		},
	}

	for index, tc := range tests {
		tc.testCase = index
		got := generateName(tc.file)
		if got != tc.want {
			t.Errorf("Test Case #%d - Expected %s error, got: %v", tc.testCase, tc.want, got)
		}
	}

}

func TestSetup(t *testing.T) {

	exp := &FileExport{
		FilePath: "tests/data/",
	}

	err := exp.Setup()
	if err != nil {
		t.Errorf("Expected nil error, got: %v", err)
	}

}

func TestCleanUp(t *testing.T) {

	exp := &FileExport{
		FilePath: "tests/data/",
	}

	err := exp.CleanUp()
	if err != nil {
		t.Errorf("Expected nil error, got: %v", err)
	}

}

func TestExportData(t *testing.T) {

	defer cleanUp()

	// Test case 1: Successful export
	user := &User{
		UserData: UserData{
			UserID:    123456789,
			Email:     "john.doe@gmail.com",
			FirstName: "John",
			LastName:  "Doe",
		},
	}

	exp := &FileExport{
		FilePath: "tests/data/",
	}

	data, err := json.MarshalIndent(user, "", "  ")
	if err != nil {
		t.Errorf("Failed to marshal data: %v", err)
	}

	err = exp.Export(data)
	if err != nil {
		t.Errorf("Expected nil error, got: %v", err)
	}

	// Verify that the file was created and contains the expected data
	data, err = os.ReadFile("tests/data/user.json")
	if err != nil {
		t.Errorf("Failed to read file: %v", err)
	}
	expectedData, _ := json.MarshalIndent(user, "", "  ")
	if string(data) != string(expectedData) {
		t.Errorf("Expected file content: %s, got: %s", string(expectedData), string(data))
	}

	// Clean up the tests directory
	err = os.RemoveAll("tests/")
	if err != nil {
		t.Errorf("Failed to remove file: %v", err)
	}

}

func TestExportDataError(t *testing.T) {

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

	exp2 := &FileExport{
		FilePath: "/",
	}

	data, err := json.MarshalIndent(user2, "", "  ")
	if err != nil {
		t.Errorf("Failed to marshal data: %v", err)
	}

	err = exp2.Export(data)
	if err == nil {
		t.Error("Expected non-nil error, got nil")
	}

}

// cleanUp removes the tests directory
func cleanUp() {
	currentDir, err := os.Getwd()
	if err != nil {
		slog.Error("Failed to remove directory", "msg", err)
		os.Exit(1)
	}

	folderPath := path.Join(currentDir, "tests")
	err = os.RemoveAll(folderPath)
	if err != nil {
		slog.Error("Failed to remove directory", "msg", err)
		err := exec.Command("rm -rf export/tests/").Run()
		if err != nil {
			slog.Error("Failed to to issue remove command", "msg", err)
			os.Exit(1)
		}
		os.Exit(1)
	}
}
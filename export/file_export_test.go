// Copyright (c) karl-cardenas-coding
// SPDX-License-Identifier: Apache-2.0

package export

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
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

func TestConvertToCSV(t *testing.T) {

	t.FailNow()
}

func TestNewFileExport(t *testing.T) {

	tests := []struct {
		filePath       string
		fileType       string
		fileName       string
		fileNamePrefix string
		serverMode     bool
	}{
		{
			filePath:       "tests/data/",
			fileType:       "json",
			fileName:       "user",
			fileNamePrefix: "test",
			serverMode:     false,
		},
		{
			filePath:       "tests/data/",
			fileType:       "json",
			fileName:       "user",
			fileNamePrefix: "",
			serverMode:     false,
		},
		{
			filePath:       "tests/data/",
			fileType:       "json",
			fileName:       "",
			fileNamePrefix: "",
			serverMode:     true,
		},
		{
			filePath:       "tests/data/",
			fileType:       "",
			fileName:       "",
			fileNamePrefix: "",
			serverMode:     true,
		},
	}

	for _, tc := range tests {

		exp := NewFileExport(tc.filePath, tc.fileType, tc.fileName, tc.fileNamePrefix, tc.serverMode)
		if exp.FilePath != tc.filePath {
			t.Errorf("Expected %s error, got: %v", tc.filePath, exp.FilePath)
		}
		if exp.FileType != tc.fileType {
			t.Errorf("Expected %s error, got: %v", tc.fileType, exp.FileType)
		}
		if exp.FileName != tc.fileName {
			t.Errorf("Expected %s error, got: %v", tc.fileName, exp.FileName)
		}
		if exp.FileNamePrefix != tc.fileNamePrefix {
			t.Errorf("Expected %s error, got: %v", tc.fileNamePrefix, exp.FileNamePrefix)
		}
	}

}

func TestGenerateName(t *testing.T) {

	type test struct {
		testCase    int
		description string
		file        FileExport
		want        string
	}

	tests := []test{
		{
			description: "Test case 1: File name with custom prefix prefix",
			file: FileExport{
				FileNamePrefix: "test",
				FileName:       "user",
				FileType:       "json",
				ServerMode:     false,
			},
			want: "test_user.json",
		},
		{
			description: "Test case 2: File name with empty prefix",
			file: FileExport{
				FileNamePrefix: "",
				FileName:       "user",
				FileType:       "json",
				ServerMode:     false,
			},
			want: "user.json",
		},
		{
			description: "Test case 3: File name with empty prefix and file name",
			file: FileExport{
				FileNamePrefix: "",
				FileName:       "",
				FileType:       "",
				ServerMode:     false,
			},
			want: ".",
		},
		{
			description: "Test case 4: Server mode enabled",
			file: FileExport{
				FileNamePrefix: "",
				FileName:       "user",
				FileType:       "json",
				ServerMode:     true,
			},
			want: fmt.Sprintf("user_%s.json", getCurrentDate()),
		},
		{
			description: "Test case 5: Server mode enabled with custom prefix",
			file: FileExport{
				FileNamePrefix: "test",
				FileName:       "user",
				FileType:       "json",
				ServerMode:     true,
			},
			want: fmt.Sprintf("test_user_%s.json", getCurrentDate()),
		},
	}

	for index, tc := range tests {
		tc.testCase = index
		got := generateName(tc.file)
		if got != tc.want {
			t.Errorf("%s - Expected %s error, got: %v", tc.description, tc.want, got)
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

	err = cleanUp("../tests/data/")
	if err != nil {
		t.Errorf("Failed to remove directory: %v", err)
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

func TestWriteToFileError(t *testing.T) {

	tests := []struct {
		id int
		FileExport
		data             []byte
		errorExpected    bool
		createTestFile   bool
		checkFileCreated bool
	}{
		{
			0,
			FileExport{
				FilePath:       "../tests/data/",
				FileType:       "json",
				FileName:       "test_user",
				FileNamePrefix: "",
			},
			[]byte("test"),
			false,
			false,
			true,
		}, {
			0,
			FileExport{
				FilePath:       "../tests/data/",
				FileType:       "json",
				FileName:       "test_user",
				FileNamePrefix: "",
			},
			[]byte("test 2"),
			false,
			true,
			true,
		},
	}

	for index, tc := range tests {

		tc.id = index + 1

		if tc.createTestFile {
			err := os.MkdirAll("../tests/data/", 0755)
			if err != nil {
				t.Errorf("Test Case - %d: Failed to create directory: %v", tc.id, err)
			}

			file, err := os.Create(filepath.Join(tc.FilePath, tc.FileName+"."+tc.FileType))
			if err != nil {
				t.Errorf("Test Case - %d: Failed to create file: %v", tc.id, err)
			}

			_, err = file.Write(tc.data)
			if err != nil {
				t.Errorf("Test Case - %d: Failed to write to file: %v", tc.id, err)
			}

			err = file.Close()
			if err != nil {
				t.Errorf("Test Case - %d: Failed to close file: %v", tc.id, err)
			}
		}

		err := writeToFile(tc.FileExport, tc.data)
		if tc.errorExpected && err == nil {
			t.Errorf("Test Case - %d: Expected non-nil error, got nil", tc.id)
		}

		if tc.checkFileCreated {
			_, err := os.Stat(filepath.Join(tc.FilePath, tc.FileName+"."+tc.FileType))
			if err != nil {
				if os.IsNotExist(err) {
					t.Errorf("Test Case - %d: File was not created", tc.id)
				}
			}
		}

		if !tc.errorExpected && err != nil {
			t.Errorf("Test Case - %d: Expected nil error, got: %v", tc.id, err)
		}

		err = cleanUp("../tests/data/")
		if err != nil {
			t.Errorf("Test Case - %d: Failed to remove directory: %v", tc.id, err)
		}

	}

}

// cleanUp removes the tests directory
func cleanUp(path string) error {

	var pathToDelete string

	currentDir, err := os.Getwd()
	if err != nil {

		return err
	}

	pathToDelete = filepath.Join(currentDir, "tests", "data")

	if path != "" {
		pathToDelete = path
	}

	err = os.RemoveAll(pathToDelete)
	if err != nil {
		slog.Error("Failed to remove directory", "msg", err)
		err := exec.Command("rm -rf export/tests/").Run()
		if err != nil {
			return err
		}
		return err
	}

	return nil
}

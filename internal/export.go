package internal

import (
	"encoding/json"
	"log/slog"
	"os"
)

// ExportDataToFile exports the user data to a file
// The file path is optional, if not provided, the file will be created in the data folder in the current directory
// The file will be named user.json
// Example usage: user.ExportDataToFile("data/user.json")
func (u *User) ExportDataToFile(filePath string) error {

	// MarshalIndent the data
	jsonData, err := json.MarshalIndent(u, "", "  ")
	if err != nil {
		slog.Error("unable to marshal data", err)
		return err
	}

	if filePath == "" {
		filePath = "data/user.json"
	}

	// write the data to a file in the data folder in the current directory
	err = WriteToFile(filePath, jsonData)
	if err != nil {
		slog.Error("unable to write to file", err)
		return err
	}

	return nil

}

// WriteToFile writes data to a file
func WriteToFile(path string, data []byte) error {

	// check if the data folder exists
	if _, err := os.Stat("data"); os.IsNotExist(err) {
		// create the data folder
		err := os.Mkdir("data", 0755)
		if err != nil {
			slog.Error("unable to create data folder", err)
			return err
		}
	}

	// Remove the file if it exists
	if _, err := os.Stat(path); err == nil {
		err := os.Remove(path)
		if err != nil {
			slog.Error("unable to remove file", err)
			return err
		}
	}

	f, err := os.Create(path)
	if err != nil {
		slog.Error("unable to create file", err)
		return err
	}

	defer f.Close()

	dataPretty := string(data)

	_, err = f.WriteString(dataPretty)
	if err != nil {
		slog.Error("unable to write to file", err)
		return err
	}

	return nil
}

// Copyright (c) karl-cardenas-coding
// SPDX-License-Identifier: Apache-2.0

package export

import (
	"log/slog"
	"os"
	"path"
)

// Setup sets up the file export and any resources required
func (f *FileExport) Setup() error {
	// no setup required
	return nil
}

// ExportDataToFile exports the user data to a file
// The file path is optional, if not provided, the file will be created in the data folder in the current directory
// The file will be named user.json by default
func (f *FileExport) Export(data []byte) error {

	currentDir, err := os.Getwd()
	if err != nil {
		slog.Error("unable to get current directory", err)
		return err
	}

	if f.FilePath == "" {
		f.FilePath = path.Join(currentDir, "data")
	}

	if f.FileType == "" {
		f.FileType = "json"
	}

	if f.FileName == "" {
		f.FileName = "user"
	}

	// write the data to a file in the data folder in the current directory
	err = WriteToFile(*f, data)
	if err != nil {
		slog.Error("unable to write to file", "error", err)
		return err
	}

	return nil

}

// generateName generates the name of the file to be created
func generateName(cfg FileExport) string {

	var fileName string

	if cfg.FileNamePrefix != "" {
		fileName = cfg.FileNamePrefix + "_" + cfg.FileName + "." + cfg.FileType
	} else {
		fileName = cfg.FileName + "." + cfg.FileType
	}

	return fileName

}

// WriteToFile writes data to a file
func WriteToFile(cfg FileExport, data []byte) error {

	// check if the path folder exists
	if _, err := os.Stat(cfg.FilePath); os.IsNotExist(err) {
		// create the data folder
		err := os.MkdirAll(cfg.FilePath, 0755)
		if err != nil {
			slog.Error("unable to create data folder", "error", err)
			return err
		}
	}

	fileName := generateName(cfg)

	// Remove the file if it exists
	if _, err := os.Stat(path.Join(cfg.FilePath, fileName)); err == nil {
		err := os.RemoveAll(cfg.FilePath)
		if err != nil {
			slog.Error("unable to remove file", "error", err)
			return err
		}
	}

	f, err := os.Create(path.Join(cfg.FilePath, fileName))
	if err != nil {
		slog.Error("unable to create file", "error", err)
		return err
	}

	defer f.Close()

	dataPretty := string(data)

	_, err = f.WriteString(dataPretty)
	if err != nil {
		slog.Error("unable to write to file", "error", err)
		return err
	}

	slog.Info("data written to file", "file", path.Join(cfg.FilePath, fileName))

	return nil
}

// CleanUp cleans up the file export and any resources required
func (f *FileExport) CleanUp() error {
	// no cleanup required
	return nil
}

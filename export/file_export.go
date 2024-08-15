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

// NewFileExport creates a new file export
func NewFileExport(filePath, fileType, fileName, fileNamePrefix string, serverMode bool) *FileExport {
	return &FileExport{
		FilePath:       filePath,
		FileType:       fileType,
		FileName:       fileName,
		FileNamePrefix: fileNamePrefix,
		ServerMode:     serverMode,
	}
}

// ExportDataToFile exports the user data to a file
// The file path is optional, if not provided, the file will be created in the data folder in the current directory
// The file will be named user.json by default
func (f *FileExport) Export(data []byte) error {

	currentDir, err := os.Getwd()
	if err != nil {
		slog.Error("unable to get current directory", "error", err)
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

	switch f.FileType {
	case "json":
		// write the data to a file in the data folder in the current directory
		err = writeToFile(*f, data)
		if err != nil {
			slog.Error("unable to write to  JSON file", "error", err)
			return err
		}
	case "xlsx":
		// err = writeToExcelFile(*f, data)
		// if err != nil {
		// 	slog.Error("unable to write to Excel file", "error", err)
		// 	return err
		// }
	default:
		err = writeToFile(*f, data)
		if err != nil {
			slog.Error("unable to write to JSON file", "error", err)
			return err
		}
	}

	return nil

}

// generateName generates the name of the file to be created
func generateName(cfg FileExport) string {

	if cfg.ServerMode {

		if cfg.FileNamePrefix != "" {
			return cfg.FileNamePrefix + "_" + cfg.FileName + "_" + getCurrentDate() + "." + cfg.FileType
		}
		return cfg.FileName + "_" + getCurrentDate() + "." + cfg.FileType
	}

	if cfg.FileNamePrefix != "" {
		return cfg.FileNamePrefix + "_" + cfg.FileName + "." + cfg.FileType
	}
	slog.Warn("File Type", "Type", cfg.FileType)
	return cfg.FileName + "." + cfg.FileType

}

// writeToFile writes data to a file
func writeToFile(cfg FileExport, data []byte) error {

	fileName := generateName(cfg)

	// check if the path folder exists, if not create it
	_, err := os.Stat(cfg.FilePath)
	if err != nil {

		if os.IsNotExist(err) {
			err := os.MkdirAll(cfg.FilePath, 0755)
			if err != nil {
				slog.Error("unable to create data folder", "error", err)
				return err
			}
		}
		// Remove identical file if it exists to avoid conflicts
	} else {
		if _, err := os.Stat(path.Join(cfg.FilePath, fileName)); err == nil {
			slog.Info("file already exists, removing it", "file", path.Join(cfg.FilePath, fileName))
			err := os.Remove(path.Join(cfg.FilePath, fileName))
			if err != nil {
				slog.Error("unable to remove file", "file", path.Join(cfg.FilePath, fileName), "error", err)
				return err
			}
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
		slog.Error("unable to write the content to the file", "error", err)
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

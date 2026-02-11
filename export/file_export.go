// Copyright (c) karl-cardenas-coding
// SPDX-License-Identifier: Apache-2.0

package export

import (
	"errors"
	"log/slog"
	"os"
	"path"
)

// Setup sets up the file export and any resources required
func (f *FileExport) Setup() error {
	err := applyFileExportDefaults(f)
	if err != nil {
		return err
	}

	// sqlite exports use a stable file name across server runs to support incremental updates.
	if f.FileType == "sqlite" {
		fileDestination := path.Join(f.FilePath, generateName(*f))
		return initializeSQLiteFile(fileDestination)
	}

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
	err := applyFileExportDefaults(f)
	if err != nil {
		slog.Error("unable to apply file export defaults", "error", err)
		return err
	}

	// write the data to a file in the data folder in the current directory
	err = writeToFile(*f, data)
	if err != nil {
		slog.Error("unable to write to  JSON file", "error", err)
		return err
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

	return cfg.FileName + "." + cfg.FileType

}

// writeToFile writes data to a file
func writeToFile(cfg FileExport, data []byte) error {

	fileName := generateName(cfg)
	fileDestination := path.Join(cfg.FilePath, fileName)

	// check if the path folder exists, if not create it
	err := createExportDir(cfg.FilePath)
	if err != nil {
		slog.Error("unable to create data folder", "error", err)
		return err
	}

	fileExists, err := fileExists(fileDestination)
	if err != nil {
		return err
	}

	if cfg.FileType == "sqlite" {
		if fileExists {
			err = mergeSQLiteFile(fileDestination, data)
		} else {
			err = writeSQLiteData(fileDestination, data)
		}
		if err != nil {
			slog.Error("unable to write sqlite data to the file", "error", err)
			return err
		}

		slog.Info("data written to file", "file", fileDestination)
		return nil
	}

	// Remove identical file if it exists to avoid conflicts
	if fileExists {
		slog.Info("file already exists, removing it", "file", fileDestination)
		err := os.Remove(fileDestination)
		if err != nil {
			slog.Error("unable to remove file", "file", fileDestination, "error", err)
			return err
		}
	}

	f, err := os.Create(fileDestination)
	if err != nil {
		slog.Error("unable to create file", "error", err)
		return err
	}

	defer func() { _ = f.Close() }()

	_, err = f.Write(data)
	if err != nil {
		slog.Error("unable to write the content to the file", "error", err)
		return err
	}

	slog.Info("data written to file", "file", fileDestination)

	return nil
}

// CleanUp cleans up the file export and any resources required
func (f *FileExport) CleanUp() error {
	// no cleanup required
	return nil
}

func applyFileExportDefaults(f *FileExport) error {
	currentDir, err := os.Getwd()
	if err != nil {
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

	if f.FileType == "sqlite" {
		// sqlite exports always use a stable filename to support ongoing merges.
		f.ServerMode = false
	}

	return nil
}

func createExportDir(filePath string) error {
	if _, err := os.Stat(filePath); err != nil {
		if os.IsNotExist(err) {
			return os.MkdirAll(filePath, 0755)
		}
		return err
	}

	return nil
}

func fileExists(filePath string) (bool, error) {
	_, err := os.Stat(filePath)
	if err == nil {
		return true, nil
	}

	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}

	return false, err
}

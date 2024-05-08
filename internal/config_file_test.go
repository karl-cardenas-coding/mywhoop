// Copyright (c) karl-cardenas-coding
// SPDX-License-Identifier: MIT

package internal

import (
	"testing"
)

func TestFileNotFound(t *testing.T) {

	_, err := readConfigFileYaml("invalid_config.yaml")
	if err == nil {
		t.Log(err)
		t.Fatalf("Failed to read the Yaml file. Expected error but received %d", err)
	}
}

func TestInvalidYAML(t *testing.T) {

	c, err := readConfigFileYaml("../tests/invalid_config.yaml")
	if err != nil {
		t.Fatalf("Failed to read the Yaml file %v", err)
	}

	err = validateConfiguration(c)
	if err == nil {
		t.Log(err)
		t.Fatalf("Failed to validate the configuration. Expected %v but received %v", nil, err)
	}
}

func TestReadConfigFileYaml(t *testing.T) {

	expectedMethod := "file"
	expectedFilePath := "data/"
	expectedName := "user"
	expectedFileType := "json"
	expectedDebug := "debug"

	got, err := readConfigFileYaml("../tests/valid_config.yaml")
	if err != nil {
		t.Fatalf("Failed to read the Yaml file. Expected error but received %d", err)
	}

	if got.Export.Method != expectedMethod {
		t.Fatalf("Failed to read the Yaml file. Expected %s but received %s", expectedMethod, got.Export.Method)
	}

	if got.Export.FileExport.FilePath != expectedFilePath {
		t.Fatalf("Failed to read the Yaml file. Expected %s but received %s", expectedFilePath, got.Export.FileExport.FilePath)
	}

	if got.Export.FileExport.FileName != expectedName {
		t.Fatalf("Failed to read the Yaml file. Expected %s but received %s", expectedName, got.Export.FileExport.FileName)
	}

	if got.Export.FileExport.FileType != expectedFileType {
		t.Fatalf("Failed to read the Yaml file. Expected %s but received %s", expectedFileType, got.Export.FileExport.FileType)
	}

	if got.Debug != expectedDebug {
		t.Fatalf("Failed to read the Yaml file. Expected %s but received %s", expectedDebug, got.Debug)
	}

}

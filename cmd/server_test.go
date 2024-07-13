// Copyright (c) karl-cardenas-coding
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"testing"

	"github.com/karl-cardenas-coding/mywhoop/export"
	"github.com/karl-cardenas-coding/mywhoop/internal"
)

func TestEvaluateConfigOptions(t *testing.T) {

	dt := &internal.ConfigurationData{
		Export: internal.ConfigExport{
			Method: "file",
			FileExport: export.FileExport{
				FilePath: "tests/data/",
			},
		},
		Server: internal.Server{
			Enabled: true,
		},
	}

	expectedExporter := "file"

	err := evaluateConfigOptions(dt)
	if err != nil {
		t.Errorf("Expected nil error, got: %v", err)
	}

	if dt.Export.Method != expectedExporter {
		t.Errorf("Expected %v, got: %v", expectedExporter, dt.Export.Method)
	}

	// Second test
	dt.Export.Method = ""

	err = evaluateConfigOptions(dt)
	if err != nil {
		t.Errorf("Expected nil error, got: %v", err)
	}

	if dt.Export.Method != expectedExporter {
		t.Errorf("Expected %v, got: %v", expectedExporter, dt.Export.Method)
	}

	// Third test

	dt = &internal.ConfigurationData{}
	err = evaluateConfigOptions(dt)
	if err != nil {
		t.Errorf("Expected nil error, got: %v", err)
	}

	if dt.Export.Method != expectedExporter {
		t.Errorf("Expected %v, got: %v", expectedExporter, dt.Export.Method)
	}

	// Fourth test
	dt = &internal.ConfigurationData{
		Server: internal.Server{
			Enabled: true,
		},
	}
	err = evaluateConfigOptions(dt)
	if err != nil {
		t.Errorf("Expected nil error, got: %v", err)
	}

	if dt.Export.Method != expectedExporter {
		t.Errorf("Expected %v, got: %v", expectedExporter, dt.Export.Method)
	}

}

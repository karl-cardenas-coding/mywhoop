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
			FirstRunDownload: true,
			Enabled:          true,
		},
	}

	expectedFirstRunDownload := true
	expectedExporter := "file"

	err := evaluateConfigOptions(true, dt)
	if err != nil {
		t.Errorf("Expected nil error, got: %v", err)
	}

	if dt.Server.FirstRunDownload != expectedFirstRunDownload {
		t.Errorf("Expected %v, got: %v", expectedFirstRunDownload, dt.Server.FirstRunDownload)
	}

	if dt.Export.Method != expectedExporter {
		t.Errorf("Expected %v, got: %v", expectedExporter, dt.Export.Method)
	}

	// Second test
	dt.Export.Method = ""
	dt.Server.FirstRunDownload = false
	expectedFirstRunDownload = false

	err = evaluateConfigOptions(false, dt)
	if err != nil {
		t.Errorf("Expected nil error, got: %v", err)
	}

	if dt.Server.FirstRunDownload != expectedFirstRunDownload {
		t.Errorf("Expected %v, got: %v", expectedFirstRunDownload, dt.Server.FirstRunDownload)
	}

	if dt.Export.Method != expectedExporter {
		t.Errorf("Expected %v, got: %v", expectedExporter, dt.Export.Method)
	}

	// Third test

	dt = &internal.ConfigurationData{}
	expectedFirstRunDownload = true
	err = evaluateConfigOptions(true, dt)
	if err != nil {
		t.Errorf("Expected nil error, got: %v", err)
	}

	if dt.Server.FirstRunDownload != expectedFirstRunDownload {
		t.Errorf("Expected %v, got: %v", expectedFirstRunDownload, dt.Server.FirstRunDownload)
	}

	if dt.Export.Method != expectedExporter {
		t.Errorf("Expected %v, got: %v", expectedExporter, dt.Export.Method)
	}

	// Fourth test
	dt = &internal.ConfigurationData{
		Server: internal.Server{
			FirstRunDownload: true,
		},
	}
	expectedFirstRunDownload = true
	err = evaluateConfigOptions(false, dt)
	if err != nil {
		t.Errorf("Expected nil error, got: %v", err)
	}

	if dt.Server.FirstRunDownload != expectedFirstRunDownload {
		t.Errorf("Expected %v, got: %v", expectedFirstRunDownload, dt.Server.FirstRunDownload)
	}

	if dt.Export.Method != expectedExporter {
		t.Errorf("Expected %v, got: %v", expectedExporter, dt.Export.Method)
	}

}

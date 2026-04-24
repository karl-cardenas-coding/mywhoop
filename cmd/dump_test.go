// Copyright (c) karl-cardenas-coding
// SPDX-License-Identifier: MIT

package cmd

import (
	"strings"
	"testing"
)

// TestDumpCmd_OutputFlagValidation exercises the PreRunE hook on dumpCmd.
// PreRunE is the gate that rejects `dump -o xlxs` before any auth or API
// work happens, so it's the single most important regression test for the
// invalid-flag class of bugs.
func TestDumpCmd_OutputFlagValidation(t *testing.T) {
	if dumpCmd.PreRunE == nil {
		t.Fatal("dumpCmd.PreRunE is nil; expected the output-flag validator to be wired up")
	}

	cases := []struct {
		name      string
		input     string
		wantErr   bool
		wantLower string // when wantErr is false, the normalized value we expect
	}{
		{name: "empty passes (defaults to json downstream)", input: "", wantErr: false, wantLower: ""},
		{name: "json", input: "json", wantErr: false, wantLower: "json"},
		{name: "xlsx", input: "xlsx", wantErr: false, wantLower: "xlsx"},
		{name: "sqlite", input: "sqlite", wantErr: false, wantLower: "sqlite"},
		{name: "uppercase JSON normalizes", input: "JSON", wantErr: false, wantLower: "json"},
		{name: "mixed case XlSx normalizes", input: "XlSx", wantErr: false, wantLower: "xlsx"},
		{name: "typo xlxs rejected", input: "xlxs", wantErr: true},
		{name: "csv rejected", input: "csv", wantErr: true},
		{name: "yaml rejected", input: "yaml", wantErr: true},
		{name: "gibberish rejected", input: "not-a-format", wantErr: true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// The flag is a package-level var, so snapshot and restore it.
			saved := output
			t.Cleanup(func() { output = saved })

			output = tc.input
			err := dumpCmd.PreRunE(dumpCmd, nil)

			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error for input %q, got nil", tc.input)
				}
				if !strings.Contains(err.Error(), "invalid --output") {
					t.Errorf("error %q did not mention 'invalid --output'", err.Error())
				}
				// Error message should list the supported values so the user
				// knows what's acceptable.
				for _, want := range []string{"json", "xlsx", "sqlite"} {
					if !strings.Contains(err.Error(), want) {
						t.Errorf("error %q did not mention supported value %q", err.Error(), want)
					}
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error for input %q: %v", tc.input, err)
			}
			if output != tc.wantLower {
				t.Errorf("expected output normalized to %q, got %q", tc.wantLower, output)
			}
		})
	}
}

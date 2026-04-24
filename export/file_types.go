// Copyright (c) karl-cardenas-coding
// SPDX-License-Identifier: Apache-2.0

package export

import "slices"

// SupportedFileTypes is the canonical list of output formats the exporters
// accept. Any code that lets a user pick an output format (CLI flag, config
// file, etc.) should validate against this slice before handing the value to
// a constructor - downstream code no longer silently remaps unknown types.
var SupportedFileTypes = []string{"json", "xlsx", "sqlite"}

// IsValidFileType reports whether s is one of the SupportedFileTypes.
// Matching is case-sensitive: normalize your input (strings.ToLower) before
// calling this helper.
func IsValidFileType(s string) bool {
	return slices.Contains(SupportedFileTypes, s)
}

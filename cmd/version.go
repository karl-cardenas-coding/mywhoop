// Copyright (c) karl-cardenas-coding
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"fmt"
	"log/slog"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(VersionCmd)
}

var VersionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the current version number of mywhoop",
	Long:  `Prints the current version number of mywhoop`,
	RunE: func(cmd *cobra.Command, args []string) error {
		version := fmt.Sprintf("mywhoop v%s", VersionString)
		slog.Info(version)
		return nil
	},
}

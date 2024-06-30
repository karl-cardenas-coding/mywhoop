// Copyright (c) karl-cardenas-coding
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"embed"
	"os"

	"github.com/karl-cardenas-coding/mywhoop/cmd"
)

//go:embed all:web/*
var staticAssets embed.FS

// run is the entry point for the program.
func run(
	ctx context.Context,
	args []string,
	stdout,
	stderr *os.File,
) error {

	return cmd.Execute(ctx, args, stdout, stderr, staticAssets)
}

func main() {
	ctx := context.Background()
	err := run(ctx, os.Args, os.Stdin, os.Stderr)
	if err != nil {
		os.Exit(1)
	}

	os.Exit(0)

}

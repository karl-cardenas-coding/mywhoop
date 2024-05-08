// Copyright (c) karl-cardenas-coding
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"fmt"
	"os"

	"github.com/karl-cardenas-coding/mywhoop/cmd"
)

// run is the entry point for the program.
func run(
	ctx context.Context,
	args []string,
	stdout,
	stderr *os.File,
) error {
	return cmd.Execute(ctx, args, stdout, stderr)
}

func main() {
	ctx := context.Background()
	if err := run(ctx, os.Args, os.Stdin, os.Stderr); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
	os.Exit(0)

}

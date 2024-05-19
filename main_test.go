package main

import (
	"context"
	"testing"
)

func TestRun(t *testing.T) {
	err := run(context.Background(), []string{}, nil, nil)
	if err != nil {
		t.Errorf("Expected nil, got %v", err)
	}
}

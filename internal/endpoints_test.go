package internal

import (
	"testing"

	"github.com/cenkalti/backoff/v4"
)

func TestGenerateBackoff(t *testing.T) {

	expected := &backoff.ExponentialBackOff{
		InitialInterval:     DEFAULT_RETRY_INITIAL_INTERVAL,
		RandomizationFactor: DEFAULT_RETRY_RANDOMIZATION,
		Multiplier:          DEFAULT_RETRY_MULTIPLIER,
		MaxElapsedTime:      DEFAULT_RETRY_MAX_ELAPSED_TIME,
	}
	// Test the backoff generation
	backoff := generateBackoff()
	if backoff.InitialInterval != expected.InitialInterval {
		t.Errorf("Expected %v, got %v", expected.InitialInterval, backoff.InitialInterval)
	}
	if backoff.RandomizationFactor != expected.RandomizationFactor {
		t.Errorf("Expected %v, got %v", expected.RandomizationFactor, backoff.RandomizationFactor)
	}

	if backoff.Multiplier != expected.Multiplier {
		t.Errorf("Expected %v, got %v", expected.Multiplier, backoff.Multiplier)
	}

	if backoff.MaxElapsedTime != expected.MaxElapsedTime {
		t.Errorf("Expected %v, got %v", expected.MaxElapsedTime, backoff.MaxElapsedTime)
	}

}

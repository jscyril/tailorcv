package main

import "testing"

func TestRunVerificationCommandIgnoresNormalApplicationArguments(t *testing.T) {
	handled, err := runVerificationCommand([]string{"--unrelated"})
	if handled || err != nil {
		t.Fatalf("runVerificationCommand() = %v, %v", handled, err)
	}
}

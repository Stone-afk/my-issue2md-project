package main

import (
	"os/exec"
	"testing"
)

func TestMainPackageBuilds(t *testing.T) {
	cmd := exec.Command("go", "build", ".")
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("go build . failed: %v\n%s", err, string(output))
	}
}

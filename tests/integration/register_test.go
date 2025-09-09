package integration

import (
	"os"
	"os/exec"
	"testing"
)

// helper funtion - migrate up/ down using goose
const migrationDir = "./sql/schema"

func gooseUp(t *testing.T, dir string) {
	t.Helper()
	cmd := exec.Command("goose", "-dir", dir, "up")
	cmd.Env = os.Environ()
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("goose up: %v\n%s", err, out)
	}
}

func gooseDown(t *testing.T, dir string) {
	t.Helper()
	cmd := exec.Command("goose", "-dir", dir, "down-to", "0")
	cmd.Env = os.Environ()
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("goose down: %v\n%s", err, out)
	}
}

//Start DB
//Register - t.run success, duplicate email/username, missing(email, username, password), bad JSON, user exists

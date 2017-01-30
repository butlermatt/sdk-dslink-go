package client

import (
	"testing"
)

func TestDial(t *testing.T) {
	c, err := dial("http://localhost:8080/conn", "myTest-", "", "")

	if err != nil {
		t.Fatalf("Dial failed: %v", err)
	}

	if c == nil {
		t.Fatal("Connection is nil")
	}
}
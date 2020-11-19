package config

import (
	"fmt"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	rc := m.Run()

	// rc 0 means we've passed,
	// and CoverMode will be non empty if run with -cover
	if rc == 0 && testing.CoverMode() != "" {
		c := testing.Coverage()
		if c < 0.80 {
			fmt.Println("Tests passed but coverage failed at", c)
			rc = -1
		} else {
			fmt.Println("Coverage enforcement succeeded at ", c)
		}
	}
	os.Exit(rc)
}

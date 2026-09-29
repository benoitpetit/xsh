package core

import (
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	if os.Getenv("XSH_CONFIG_DIR") != "" {
		os.Exit(m.Run())
	}
	dir, err := os.MkdirTemp("", "xsh-core-tests-")
	if err != nil {
		panic(err)
	}
	if err := os.Setenv("XSH_CONFIG_DIR", dir); err != nil {
		panic(err)
	}
	code := m.Run()
	_ = os.RemoveAll(dir)
	os.Exit(code)
}

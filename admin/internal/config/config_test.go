package config

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/gogf/gf/v2/frame/g"
)

func TestLoadDefaultsToDevelopment(t *testing.T) {
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("cannot resolve test file path")
	}
	projectRoot := filepath.Clean(filepath.Join(filepath.Dir(currentFile), "../.."))
	workingDirectory, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err = os.Chdir(projectRoot); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(workingDirectory)
	})
	t.Setenv("APP_ENV", "")

	environment, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if environment != "dev" {
		t.Fatalf("environment = %q, want dev", environment)
	}
	if !g.Cfg().MustGet(t.Context(), "database.default.debug").Bool() {
		t.Fatal("expected development database debug logging to be enabled")
	}
}

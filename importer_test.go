package godetector

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFindPackageDefinitionDir(t *testing.T) {
	defDir, err := FindPackageDefinitionDir("net/http", ".")
	if err != nil {
		t.Error(err)
	}
	t.Log(defDir)
	if dir, err := os.Stat(defDir); err != nil || !dir.IsDir() {
		t.Error("not a dir!", err)
	}
	defDir, err = FindPackageDefinitionDir("github.com/reddec/storages", ".")
	if err != nil {
		t.Error(err)
	}
	if dir, err := os.Stat(defDir); err != nil || !dir.IsDir() {
		t.Error("not a dir!", err)
	}
	t.Log(defDir)
	defDir, err = FindPackageDefinitionDir("golang.org/x/mod", ".")
	if err != nil {
		t.Error(err)
	}
	t.Log(defDir)
	if dir, err := os.Stat(defDir); err != nil || !dir.IsDir() {
		t.Error("not a dir!", err)
	}

	impPath, err := FindImportPath(filepath.Join(os.Getenv("GOPATH"), "pkg", "mod", "github.com/reddec/fluent-amqp@v0.1.33/flags"))
	if dir, err := os.Stat(defDir); err != nil || !dir.IsDir() {
		t.Error("not a dir!", err)
	}
	t.Log(impPath)
	if impPath != "github.com/reddec/fluent-amqp/flags" {
		t.Error("oops, should be import")
	}

	impPath, err = FindImportPath(filepath.Join(os.Getenv("GOPATH"), "pkg", "mod", "github.com/reddec/fluent-amqp@v0.1.33"))
	if dir, err := os.Stat(defDir); err != nil || !dir.IsDir() {
		t.Error("not a dir!", err)
	}
	t.Log(impPath)
	if impPath != "github.com/reddec/fluent-amqp" {
		t.Error("oops, should be import")
	}
}

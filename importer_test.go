package godetector

import (
	"os"
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
}

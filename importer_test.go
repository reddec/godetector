package godetector

import (
	"github.com/stretchr/testify/assert"
	"go/build"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestFindPackageDefinitionDir(t *testing.T) {
	defDir, err := FindPackageDefinitionDir("net/http", ".")
	if !assert.NoError(t, err) {
		return
	}
	t.Log(defDir)

	if dir, err := os.Stat(defDir); err != nil || !dir.IsDir() {
		t.Error("not a dir!", err)
		return
	}
	assert.Equal(t, filepath.Join(runtime.GOROOT(), "src", "net/http"), defDir)
	defDir, err = FindPackageDefinitionDir("github.com/reddec/storages", ".")
	if !assert.NoError(t, err) {
		return
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

	impPath, err := FindImportPath(filepath.Join(build.Default.GOPATH, "pkg", "mod", "github.com/reddec/fluent-amqp@v0.1.33/flags"))
	if dir, err := os.Stat(defDir); err != nil || !dir.IsDir() {
		t.Error("not a dir!", err)
	}
	t.Log(impPath)
	if impPath != "github.com/reddec/fluent-amqp/flags" {
		t.Error("oops, should be import")
	}

	impPath, err = FindImportPath(filepath.Join(build.Default.GOPATH, "pkg", "mod", "github.com/reddec/fluent-amqp@v0.1.33"))
	if dir, err := os.Stat(defDir); err != nil || !dir.IsDir() {
		t.Error("not a dir!", err)
	}
	t.Log(impPath)
	assert.Equal(t, "github.com/reddec/fluent-amqp", impPath)

	imp, err := InspectImportByDir("./deepparser/examples")
	if !assert.NoError(t, err) {
		return
	}
	assert.Equal(t, "examples", imp.Package)
	assert.Equal(t, "github.com/reddec/godetector/deepparser/examples", imp.Path)
	cwd, _ := filepath.Abs(".")
	assert.Equal(t, filepath.Join(cwd, "deepparser", "examples"), imp.Location)
	assert.Equal(t, cwd, imp.RootPackageLocation)
	assert.Equal(t, GoMod, imp.Type)
	t.Log(imp.RootPackageLocation)
}

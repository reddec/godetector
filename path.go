package godetector

import (
	"errors"
	"golang.org/x/mod/modfile"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// Find import path assuming that directory contains Go package.
//
// For projects with 'vendor' directory it will return path in the vendor dir.
//
// It will check all upper directories checking each of them as (a) they have go.mod file (b) directory is under GOROOT/GOPATH
func FindImportPath(dir string) (string, error) {
	const Vendor = "vendor/"
	if strings.HasPrefix(dir, Vendor) {
		return dir[len(Vendor):], nil
	}
	dir, err := filepath.Abs(dir)
	if err != nil {
		return "", err
	}
	if imp, ok := isUnderModCache(dir); ok {
		return imp, nil
	}
	return findImportPath(dir)
}

func findImportPath(dir string) (string, error) {
	if dir == "" {
		return "", errors.New("undetectable dir")
	}
	if isRootPackage(dir) {
		return "", nil
	}
	pkg, ok := isVendorPackage(dir)
	if ok {
		return pkg, nil
	}
	mod := filepath.Base(dir)
	if mod == dir {
		return "", errors.New("reached top directory")
	}
	top, err := findImportPath(filepath.Dir(dir))
	if err != nil {
		return "", err
	}
	if top != "" {
		return top + "/" + mod, nil
	}
	return mod, nil
}

func isVendorPackage(path string) (string, bool) {

	data, err := ioutil.ReadFile(filepath.Join(path, "go.mod"))
	if err != nil {
		return "", false
	}
	mod, err := modfile.Parse(path, data, nil)
	if err != nil {
		return "", false
	}

	return mod.Module.Mod.Path, true
}

func isUnderModCache(path string) (string, bool) {
	GOCACHE := filepath.Join(os.Getenv("GOPATH"), "pkg", "mod")
	absPath, _ := filepath.Abs(path)
	if !relatesToPackage(GOCACHE, absPath) {
		return "", false
	}
	t := tail(GOCACHE, absPath)
	// abc@1.2.3/x/y/z
	p := strings.Split(t, "@")
	// p = {abc, 1.2.3/x/y/z}
	sl := strings.Index(p[1], "/")
	if sl > 0 {
		p[1] = p[1][sl:]
		return p[0] + p[1], true
	}
	return p[0], true
}

func isRootPackage(path string) bool {
	GOPATH := filepath.Join(os.Getenv("GOPATH"), "src")
	GOROOT := filepath.Join(runtime.GOROOT(), "src")
	return isRootOf(path, GOPATH) || isRootOf(path, GOROOT)
}

func isRootOf(path, root string) bool {
	root, _ = filepath.Abs(root)
	path, _ = filepath.Abs(path)
	return root == path
}

type packageKind int

const (
	gopath      packageKind = 0
	modules     packageKind = 1
	unspecified packageKind = gopath
)

func findPackageRootDir(absPath string) (string, packageKind) {
	parts := strings.Split(absPath, string(os.PathSeparator))
	for i := len(parts) - 1; i >= 0; i-- {
		path := string(os.PathSeparator) + filepath.Join(parts[:i]...)
		_, ok := isVendorPackage(path)
		if ok {
			return path, modules
		}
		if isRootPackage(path) {
			return path, gopath
		}
	}
	return absPath, unspecified
}

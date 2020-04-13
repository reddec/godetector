package godetector

import (
	"errors"
	"golang.org/x/mod/modfile"
	"golang.org/x/mod/module"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// Find package definition with respect to gomodules
func FindPackageDefinitionDir(importPath string, workDir string) (string, error) {
	abs, err := filepath.Abs(workDir)
	if err != nil {
		return "", err
	}
	rootDir, kind := findPackageRootDir(abs)

	packageImport, _ := FindImportPath(rootDir)

	if relatesToPackage(packageImport, importPath) {
		childDir := importPath[len(packageImport):]
		return filepath.Join(rootDir, childDir), nil
	}

	if kind == gopath || kind == unspecified {
		return findPackagePathInGoPath(importPath)
	}
	return findPackagePathInModules(importPath, rootDir)
}

func findPackagePathInGoPath(importPath string) (string, error) {
	inRootPath := filepath.Join(runtime.GOROOT(), "src", importPath)
	if _, err := os.Stat(inRootPath); err == nil {
		return inRootPath, nil
	}
	inGoPath := filepath.Join(os.Getenv("GOPATH"), "src", importPath)
	if _, err := os.Stat(inGoPath); err == nil {
		return inGoPath, nil
	}
	return "", errors.New("not found in GOROOT or in GOPATH")
}

func findPackagePathInModules(importPath, modProjectDir string) (string, error) {
	path := filepath.Join(modProjectDir, "go.mod")
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return "", err
	}
	mod, err := modfile.Parse(path, data, nil)
	if err != nil {
		return "", err
	}
	for _, req := range mod.Require {
		if relatesToPackage(req.Mod.Path, importPath) {
			ep, err := module.EscapePath(req.Mod.Path)
			if err != nil {
				return "", err
			}
			if req.Mod.Version != "" {
				ep = ep + "@" + req.Mod.Version
			}
			return filepath.Join(os.Getenv("GOPATH"), "pkg", "mod", ep, tail(req.Mod.Path, importPath)), nil
		}
	}
	// check in root

	return findPackagePathInGoPath(importPath)
}

// github.com/reddec/godetector/cmd
// is child of
// github.com/reddec/godetector
// but not of
// github.com/reddec/storages
// or
// github.com/reddec/godetect
func relatesToPackage(rootPkg, suspectPkg string) bool {
	if rootPkg == suspectPkg {
		return true
	}
	rParts := strings.Split(rootPkg, "/")
	cParts := strings.Split(suspectPkg, "/")
	if len(cParts) < len(rParts) {
		return false
	}
	for i, parent := range rParts {
		if cParts[i] != parent {
			return false
		}
	}
	return true
}

func tail(rootPkg, suspectPkg string) string {
	if rootPkg == suspectPkg {
		return ""
	}
	rParts := strings.Split(rootPkg, "/")
	cParts := strings.Split(suspectPkg, "/")
	if len(cParts) < len(rParts) {
		return ""
	}
	for i, parent := range rParts {
		if cParts[i] != parent {
			return ""
		}
	}
	return strings.Join(cParts[len(rParts):], "/")
}

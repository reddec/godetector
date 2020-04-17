package godetector

import (
	"errors"
	"go/build"
	"golang.org/x/mod/modfile"
	"golang.org/x/mod/module"
	"io/ioutil"
	"path/filepath"
	"runtime"
	"strings"
)

// Find package definition with respect to gomodules
func FindPackageDefinitionDir(importPath string, workDir string) (string, error) {
	info, err := InspectImport(importPath, workDir)
	if err != nil {
		return "", err
	}
	return info.ImportDir, err
}

func InspectImport(importPath string, workDir string) (info *importPathInfo, err error) {
	abs, err := filepath.Abs(workDir)
	if err != nil {
		return nil, err
	}

	rootInfo, err := InspectDirectory(abs)
	if err != nil {
		return nil, err
	}

	if relatesToPackage(rootInfo.Import, importPath) {
		childDir := importPath[len(rootInfo.Import):]
		return &importPathInfo{
			PackageRootDir: rootInfo.PackageRootDir,
			LocationType:   rootInfo.LocationType,
			ImportDir:      filepath.Join(rootInfo.PackageRootDir, childDir),
			Import:         importPath,
		}, nil
	}

	// check GOROOT
	if info, err := inspectDirectory(filepath.Join(runtime.GOROOT(), "src", importPath)); err == nil {
		return info, nil
	}
	// check local modules if applicable
	if rootInfo.LocationType == GoMod {
		if dir, err := findPackagePathInModules(importPath, rootInfo.PackageRootDir); err == nil {
			return inspectDirectory(dir)
		}
	}
	// check GOPATH
	return inspectDirectory(filepath.Join(build.Default.GOPATH, "src", importPath))
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
			return filepath.Join(build.Default.GOPATH, "pkg", "mod", ep, tail(req.Mod.Path, importPath)), nil
		}
	}
	// check in root

	return "", errors.New("not found in go.mod")
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

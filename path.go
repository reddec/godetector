package godetector

import (
	"errors"
	"fmt"
	"go/build"
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
	info, err := InspectDirectory(dir)
	if err != nil {
		return "", err
	}
	return info.Import, nil
}

func InspectDirectory(dir string) (info *importPathInfo, err error) {
	const vendor = "vendor/"
	if strings.HasPrefix(dir, vendor) {
		return &importPathInfo{
			PackageRootDir: "vendor",
			ImportDir:      dir,
			LocationType:   InLocalVendor,
			Import:         dir[len(vendor):],
		}, nil
	}
	dir, err = filepath.Abs(dir)
	if err != nil {
		return nil, err
	}
	return inspectDirectory(dir)
}

type importPathInfo struct {
	PackageRootDir string
	ImportDir      string
	LocationType   LocationType
	Import         string
}

func inspectDirectory(dir string) (*importPathInfo, error) {
	if dir == "" {
		return nil, errors.New("undetectable dir")
	}
	if v, err := os.Stat(dir); err != nil {
		return nil, err
	} else if !v.IsDir() {
		return nil, fmt.Errorf("%s is not a dir", dir)
	}
	if isGoRoot(dir) {
		return &importPathInfo{
			PackageRootDir: dir,
			ImportDir:      dir,
			LocationType:   GoRoot,
		}, nil
	}
	if isGoPath(dir) {
		return &importPathInfo{
			PackageRootDir: dir,
			ImportDir:      dir,
			LocationType:   GoPath,
		}, nil
	}
	if imp, root, ok := isUnderModCache(dir); ok {
		return &importPathInfo{
			PackageRootDir: root,
			ImportDir:      dir,
			LocationType:   GoCache,
			Import:         imp,
		}, nil
	}
	pkg, ok := hasModFile(dir)
	if ok {
		return &importPathInfo{
			PackageRootDir: dir,
			ImportDir:      dir,
			LocationType:   GoMod,
			Import:         pkg,
		}, nil
	}
	mod := filepath.Base(dir)
	if mod == dir {
		return nil, errors.New("reached top directory")
	}
	info, err := inspectDirectory(filepath.Dir(dir))
	if err != nil {
		return nil, err
	}
	if info.Import != "" {
		return &importPathInfo{
			ImportDir:      dir,
			PackageRootDir: info.PackageRootDir,
			LocationType:   info.LocationType,
			Import:         info.Import + "/" + mod,
		}, nil
	}
	return &importPathInfo{
		ImportDir:      dir,
		PackageRootDir: info.PackageRootDir,
		LocationType:   info.LocationType,
		Import:         mod,
	}, nil
}

func hasModFile(path string) (string, bool) {

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

func isUnderModCache(path string) (imp string, root string, ok bool) {
	GOCACHE := filepath.Join(build.Default.GOPATH, "pkg", "mod")
	absPath, _ := filepath.Abs(path)
	if !relatesToPackage(GOCACHE, absPath) {
		return "", "", false
	}
	t := tail(GOCACHE, absPath)
	// abc@1.2.3/x/y/z
	p := strings.Split(t, "@")
	// p = {abc, 1.2.3/x/y/z}
	sl := strings.Index(p[1], "/")

	root = filepath.Join(GOCACHE, strings.Split(t, "/")[0])

	if sl > 0 {
		p[1] = p[1][sl:]
		return p[0] + p[1], root, true
	}
	return p[0], root, true
}

func isGoPath(path string) bool {
	GOPATH := filepath.Join(build.Default.GOPATH, "src")
	return isRootOf(path, GOPATH)
}

func isGoRoot(path string) bool {
	GOROOT := filepath.Join(runtime.GOROOT(), "src")
	return isRootOf(path, GOROOT)
}

func isRootOf(path, root string) bool {
	root, _ = filepath.Abs(root)
	path, _ = filepath.Abs(path)
	return root == path
}

type Package struct {
	Location string
	Import   string
	Name     string
	Type     LocationType
}

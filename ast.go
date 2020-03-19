package godetector

import (
	"fmt"
	"go/ast"
	"strconv"
)

type Import struct {
	Path     string // example.com/project/alfa
	Package  string // alfa
	Location string // /opt/go/src/example.com/project/alfa
}

// Find correct import definition in a file: "lala" "net/http" will be resolve to "net/http"
func ResolveImport(alias string, file *ast.File, workdir string) (*Import, error) {
	if alias == "" {
		return nil, fmt.Errorf("aliase or import path should be defined")
	}
	// priority to aliases
	for _, imp := range file.Imports {
		if imp.Name != nil {
			if imp.Name.Name == alias {
				importPath, _ := strconv.Unquote(imp.Path.Value)
				dir, err := FindPackageDefinitionDir(importPath, workdir)
				if err != nil {
					return nil, err
				}
				realPackageName := FindPackageNameByDir(dir)
				return &Import{
					Path:     importPath,
					Package:  realPackageName,
					Location: dir,
				}, nil
			}
		}
	}
	for _, imp := range file.Imports {

		path, _ := strconv.Unquote(imp.Path.Value)

		pkgDir, err := FindPackageDefinitionDir(path, workdir)
		if err != nil {
			return nil, err
		}
		pkgName := FindPackageNameByDir(pkgDir)
		if pkgName == alias {
			return &Import{
				Path:     path,
				Package:  pkgName,
				Location: pkgDir,
			}, nil
		}
	}
	return nil, fmt.Errorf("unresolved alias or package %s", alias)
}

// Aggregated information about directory
func InspectImportByDir(pkgDir string) (*Import, error) {
	path, err := FindImportPath(pkgDir)
	if err != nil {
		return nil, err
	}
	pkgName := FindPackageNameByDir(pkgDir)
	return &Import{
		Path:     path,
		Package:  pkgName,
		Location: pkgDir,
	}, nil
}

package godetector

import (
	"fmt"
	"go/ast"
	"strconv"
)

type LocationType int

const (
	Local         LocationType = 0
	GoMod         LocationType = 1
	InLocalVendor LocationType = 2
	GoPath        LocationType = 3
	GoRoot        LocationType = 4
	GoCache       LocationType = 5
)

type Import struct {
	Path                string // example.com/project/alfa/beta/gamma
	Package             string // gamma
	Location            string // /opt/go/src/example.com/project/alfa/beta/gamma
	RootPackageLocation string // /opt/go/src/example.com/project/alfa
	Type                LocationType
}

func (ipi *importPathInfo) ToImport() *Import {
	if ipi == nil {
		return nil
	}
	return &Import{
		Path:                ipi.Import,
		Package:             FindPackageNameByDir(ipi.ImportDir),
		Location:            ipi.ImportDir,
		RootPackageLocation: ipi.PackageRootDir,
		Type:                ipi.LocationType,
	}

}

func (ipi *importPathInfo) Root() (*Import, error) {
	return InspectImportByDir(ipi.PackageRootDir)
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
				info, err := InspectImport(importPath, workdir)
				return info.ToImport(), err
			}
		}
	}
	for _, imp := range file.Imports {

		path, _ := strconv.Unquote(imp.Path.Value)

		pkgInfo, err := InspectImport(path, workdir)
		if err != nil {
			return nil, err
		}
		imp := pkgInfo.ToImport()
		if imp.Package == alias {
			return imp, nil
		}
	}
	return nil, fmt.Errorf("unresolved alias or package %s", alias)
}

// Aggregated information about directory
func InspectImportByDir(pkgDir string) (*Import, error) {
	info, err := InspectDirectory(pkgDir)
	return info.ToImport(), err
}

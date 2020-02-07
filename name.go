package godetector

import (
	"go/parser"
	"go/token"
	"io/ioutil"
	"path/filepath"
	"strings"
)

// Find package name by directory: scans go file to detect package definition and uses path detection as fail-over
func FindPackageNameByDir(dir string) string {
	abs, err := filepath.Abs(dir)
	if err != nil {
		abs = dir
	}
	guess := filepath.Base(abs)
	files, err := ioutil.ReadDir(abs)
	if err != nil {
		return guess
	}
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		if strings.HasSuffix(file.Name(), ".go") {
			var fs token.FileSet
			parsed, err := parser.ParseFile(&fs, filepath.Join(abs, file.Name()), nil, parser.PackageClauseOnly)
			if err != nil {
				continue
			}
			return parsed.Name.Name
		}
	}
	return guess
}

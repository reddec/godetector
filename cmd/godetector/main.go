package main

import (
	"flag"
	"fmt"
	"github.com/reddec/godetector"
	"os"
)

func main() {
	dir := flag.String("dir", ".", "Dirname to change")
	flag.Parse()
	fmt.Println("Current directory info")

	if path, err := godetector.FindImportPath(*dir); err == nil {
		fmt.Println("  import:", path, "pkg:", godetector.FindPackageNameByDir(*dir))
	} else {
		fmt.Fprintln(os.Stderr, "  failed detect path:", err)
	}

	if len(flag.Args()) > 0 {
		fmt.Println("Imports definitions")
	}
	for _, arg := range flag.Args() {

		if defDir, err := godetector.FindPackageDefinitionDir(arg, *dir); err == nil {
			fmt.Println(" ", arg, "=>", defDir)
		} else {
			fmt.Fprintln(os.Stderr, "  failed detect path for import", arg, ":", err)
		}
	}
}

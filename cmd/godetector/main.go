package main

import (
	"flag"
	"fmt"
	"github.com/reddec/godetector"
	"os"
)

func main() {
	dir := flag.String("dir", ".", "Dirname to scan")
	flag.Parse()
	fmt.Println("Directory info")

	if path, err := godetector.FindImportPath(*dir); err == nil {
		fmt.Println("  import:", path)
	} else {
		fmt.Fprintln(os.Stderr, "failed detect path:", err)
	}
	fmt.Println("  pkg:", godetector.FindPackageNameByDir(*dir))
}

package main

import (
	"flag"
	"log"
	"os"

	"github.com/gostaticanalysis/codegen"
	"github.com/podhmo/veritas/cmd/veritas/gen"
	"golang.org/x/tools/go/packages"
)

func main() {
	var (
		pkgPath    string
		outputFile string
	)
	flag.StringVar(&pkgPath, "pkg", "./def", "package path to parse")
	flag.StringVar(&outputFile, "o", "./validation/validator.go", "output file path")
	flag.Parse()

	cfg := &packages.Config{
		Mode: packages.LoadAllSyntax,
	}
	pkgs, err := packages.Load(cfg, pkgPath)
	if err != nil {
		log.Fatalf("failed to load packages: %v", err)
	}
	if len(pkgs) != 1 {
		log.Fatalf("expected 1 package, got %d", len(pkgs))
	}

	var output *os.File
	if outputFile != "" {
		f, err := os.Create(outputFile)
		if err != nil {
			log.Fatalf("failed to create output file: %v", err)
		}
		defer f.Close()
		output = f
	} else {
		output = os.Stdout
	}

	pass := &codegen.Pass{
		Pkg:    pkgs[0].Types,
		Output: output,
	}

	if err := gen.Generator.Run(pass); err != nil {
		log.Fatalf("failed to run generator: %v", err)
	}
}

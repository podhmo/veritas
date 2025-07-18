package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log/slog"
	"os"

	"github.com/podhmo/veritas"
	"golang.org/x/tools/go/packages"
)

func main() {
	// Setup structured logging.
	logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))

	// Define command-line flags.
	inPath := flag.String("in", "./...", "Input path for Go source files (e.g., ./...)")
	outFile := flag.String("out", "rules.json", "Output file for generated JSON rules")
	flag.Parse()

	logger.Info("Starting veritas rule generation", "input", *inPath, "output", *outFile)

	// TODO: Implement the main logic.
	// 1. Create a new parser.
	// 2. Parse the source files from the input path.
	// 3. Marshal the results into JSON.
	// 4. Write the JSON to the output file.

	if err := run(*inPath, *outFile, logger); err != nil {
		logger.Error("Rule generation failed", "error", err)
		os.Exit(1)
	}

	logger.Info("Rule generation completed successfully")
}

func run(inPath, outFile string, logger *slog.Logger) error {
	// Find all Go files in the provided path.
	cfg := &packages.Config{
		Mode: packages.NeedFiles,
	}
	pkgs, err := packages.Load(cfg, inPath)
	if err != nil {
		return fmt.Errorf("failed to load packages for path %q: %w", inPath, err)
	}
	if packages.PrintErrors(pkgs) > 0 {
		return fmt.Errorf("packages contain errors")
	}

	parser := NewParser(logger)
	allRuleSets := make(map[string]veritas.ValidationRuleSet)

	// Parse each file.
	for _, pkg := range pkgs {
		for _, goFile := range pkg.GoFiles {
			logger.Debug("parsing file", "path", goFile)
			ruleSets, err := parser.Parse(goFile)
			if err != nil {
				logger.Warn("failed to parse file, skipping", "file", goFile, "error", err)
				continue
			}
			// Merge the rule sets.
			for k, v := range ruleSets {
				allRuleSets[k] = v
			}
		}
	}

	// Marshal the combined rule sets into JSON.
	jsonData, err := json.MarshalIndent(allRuleSets, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal rule sets to JSON: %w", err)
	}

	// Write the JSON to the output file.
	if err := os.WriteFile(outFile, jsonData, 0644); err != nil {
		return fmt.Errorf("failed to write JSON to file %q: %w", outFile, err)
	}

	logger.Info("successfully wrote rules", "count", len(allRuleSets), "file", outFile)

	return nil
}

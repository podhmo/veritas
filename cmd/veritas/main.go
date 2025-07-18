package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log/slog"
	"os"
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
	logger.Debug("Initializing parser")
	parser := NewParser(logger)

	logger.Debug("Starting parsing", "path", inPath)
	ruleSets, err := parser.Parse(inPath)
	if err != nil {
		return fmt.Errorf("error parsing source files: %w", err)
	}
	logger.Info("Parsing complete", "rule_sets_found", len(ruleSets))

	if len(ruleSets) == 0 {
		logger.Info("No validation rules found, nothing to write")
		return nil
	}

	logger.Debug("Marshalling rule sets to JSON")
	jsonData, err := json.MarshalIndent(ruleSets, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal rule sets to JSON: %w", err)
	}

	logger.Debug("Writing JSON to output file", "file", outFile)
	if err := os.WriteFile(outFile, jsonData, 0644); err != nil {
		return fmt.Errorf("failed to write to output file %s: %w", outFile, err)
	}

	return nil
}

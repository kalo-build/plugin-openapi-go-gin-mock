package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/kalo-build/plugin-openapi-go-gin-mock/pkg/generate"
)

type StoreConfig struct {
	ID        uint32 `json:"id"`
	Type      string `json:"type"`
	MountPath string `json:"mountPath,omitempty"`
}

type PluginConfig struct {
	Stores     map[string]StoreConfig `json:"stores,omitempty"`
	InputPath  string                 `json:"inputPath,omitempty"`
	OutputPath string                 `json:"outputPath,omitempty"`
	Config     PluginConfigFields     `json:"config"`
	Verbose    bool                   `json:"verbose,omitempty"`
}

type PluginConfigFields struct {
	SpecFileName string `json:"specFileName,omitempty"`
	PackageName  string `json:"packageName,omitempty"`
	ModulePath   string `json:"modulePath,omitempty"`
	Port         string `json:"port,omitempty"`
	IDPrefix     *bool  `json:"idPrefix,omitempty"`
	Health       *bool  `json:"health,omitempty"`
	HealthPath   string `json:"healthPath,omitempty"`
	Dockerfile   *bool  `json:"dockerfile,omitempty"`
}

const (
	ErrMissingConfig      = 3
	ErrInvalidConfig      = 4
	ErrInputPathRequired  = 12
	ErrOutputPathRequired = 13
	ErrGenerateFailed     = 1
)

func logInfo(verbose bool, format string, args ...interface{}) {
	if verbose {
		fmt.Fprintf(os.Stdout, format+"\n", args...)
	}
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "Usage: plugin-openapi-go-gin-mock <config>")
		os.Exit(ErrMissingConfig)
	}

	rawConfig := os.Args[1]
	var pluginConfig PluginConfig
	if err := json.Unmarshal([]byte(rawConfig), &pluginConfig); err != nil {
		fmt.Fprintln(os.Stderr, "Error parsing config JSON:", err)
		os.Exit(ErrInvalidConfig)
	}

	var inputPath, outputPath string

	if pluginConfig.Stores != nil {
		for _, store := range pluginConfig.Stores {
			switch store.MountPath {
			case "/input":
				inputPath = "/input"
			case "/output":
				outputPath = "/output"
			}
		}
	}

	if inputPath == "" && pluginConfig.InputPath != "" {
		inputPath = pluginConfig.InputPath
	}
	if outputPath == "" && pluginConfig.OutputPath != "" {
		outputPath = pluginConfig.OutputPath
	}

	if inputPath == "" {
		fmt.Fprintln(os.Stderr, "Error: Input path is required (directory containing OpenAPI spec)")
		os.Exit(ErrInputPathRequired)
	}
	if outputPath == "" {
		fmt.Fprintln(os.Stderr, "Error: Output path is required")
		os.Exit(ErrOutputPathRequired)
	}

	if inputPath[0] != '/' {
		if abs, err := filepath.Abs(inputPath); err == nil {
			inputPath = abs
		}
	}
	if outputPath[0] != '/' {
		if abs, err := filepath.Abs(outputPath); err == nil {
			outputPath = abs
		}
	}

	idPrefix := true
	if pluginConfig.Config.IDPrefix != nil {
		idPrefix = *pluginConfig.Config.IDPrefix
	}
	health := true
	if pluginConfig.Config.Health != nil {
		health = *pluginConfig.Config.Health
	}
	dockerfile := true
	if pluginConfig.Config.Dockerfile != nil {
		dockerfile = *pluginConfig.Config.Dockerfile
	}

	cfg := generate.Config{
		InputDir:     inputPath,
		OutputDir:    outputPath,
		SpecFileName: pluginConfig.Config.SpecFileName,
		PackageName:  pluginConfig.Config.PackageName,
		ModulePath:   pluginConfig.Config.ModulePath,
		Port:         pluginConfig.Config.Port,
		IDPrefix:     idPrefix,
		Health:       health,
		HealthPath:   pluginConfig.Config.HealthPath,
		Dockerfile:   dockerfile,
	}
	cfg.Resolve()

	logInfo(pluginConfig.Verbose, "Reading OpenAPI spec from: '%s'", filepath.Join(cfg.InputDir, cfg.SpecFileName))
	logInfo(pluginConfig.Verbose, "Generating mock server to: '%s'", cfg.OutputDir)

	if err := generate.GenerateMockServer(cfg); err != nil {
		fmt.Fprintln(os.Stderr, "Mock server generation failed:", err)
		os.Exit(ErrGenerateFailed)
	}

	logInfo(pluginConfig.Verbose, "Mock server generated successfully")
	os.Exit(0)
}

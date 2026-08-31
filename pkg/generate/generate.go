package generate

import (
	"fmt"
	"os"
	"path/filepath"
)

// GenerateMockServer reads an OpenAPI spec and generates a complete
// Go mock server with Gin handlers and in-memory storage.
func GenerateMockServer(cfg Config) error {
	specPath := filepath.Join(cfg.InputDir, cfg.SpecFileName)
	specBytes, err := os.ReadFile(specPath)
	if err != nil {
		return fmt.Errorf("failed to read spec file '%s': %w", specPath, err)
	}

	doc, err := ParseSpec(specBytes)
	if err != nil {
		return fmt.Errorf("failed to parse OpenAPI spec: %w", err)
	}

	pkgDir := filepath.Join(cfg.OutputDir, cfg.PackageName)
	if err := os.MkdirAll(pkgDir, 0o755); err != nil {
		return fmt.Errorf("failed to create package directory: %w", err)
	}

	cmdDir := filepath.Join(cfg.OutputDir, "cmd", "server")
	if err := os.MkdirAll(cmdDir, 0o755); err != nil {
		return fmt.Errorf("failed to create cmd directory: %w", err)
	}

	types, err := GenerateTypes(doc, cfg.PackageName)
	if err != nil {
		return fmt.Errorf("failed to generate types: %w", err)
	}
	if err := writeFile(filepath.Join(pkgDir, "types.go"), types); err != nil {
		return err
	}

	store, err := GenerateStore(doc, cfg)
	if err != nil {
		return fmt.Errorf("failed to generate store: %w", err)
	}
	if err := writeFile(filepath.Join(pkgDir, "store.go"), store); err != nil {
		return err
	}

	handlers, err := GenerateHandlers(doc, cfg)
	if err != nil {
		return fmt.Errorf("failed to generate handlers: %w", err)
	}
	if err := writeFile(filepath.Join(pkgDir, "handlers.go"), handlers); err != nil {
		return err
	}

	routes, err := GenerateRoutes(doc, cfg)
	if err != nil {
		return fmt.Errorf("failed to generate routes: %w", err)
	}
	if err := writeFile(filepath.Join(pkgDir, "routes.go"), routes); err != nil {
		return err
	}

	mainGo, err := GenerateMain(doc, cfg)
	if err != nil {
		return fmt.Errorf("failed to generate main.go: %w", err)
	}
	if err := writeFile(filepath.Join(cmdDir, "main.go"), mainGo); err != nil {
		return err
	}

	goMod := GenerateGoMod(cfg)
	if err := writeFile(filepath.Join(cfg.OutputDir, "go.mod"), goMod); err != nil {
		return err
	}

	if cfg.Dockerfile {
		dockerfile := GenerateDockerfile(cfg)
		if err := writeFile(filepath.Join(cfg.OutputDir, "Dockerfile"), dockerfile); err != nil {
			return err
		}
	}

	return nil
}

func writeFile(path, content string) error {
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return fmt.Errorf("failed to write %s: %w", path, err)
	}
	return nil
}

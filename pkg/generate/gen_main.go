package generate

import (
	"fmt"
	"strings"
)

// GenerateMain produces the server entry point (main.go) for the mock server.
func GenerateMain(doc *Document, cfg Config) (string, error) {
	var b strings.Builder

	b.WriteString("package main\n\n")
	b.WriteString("import (\n")
	b.WriteString("\t\"log\"\n")
	b.WriteString("\t\"os\"\n\n")
	b.WriteString("\t\"github.com/gin-gonic/gin\"\n")
	b.WriteString(fmt.Sprintf("\t\"%s/%s\"\n", cfg.ModulePath, cfg.PackageName))
	b.WriteString(")\n\n")

	b.WriteString("func main() {\n")
	b.WriteString(fmt.Sprintf("\tport := os.Getenv(\"PORT\")\n"))
	b.WriteString("\tif port == \"\" {\n")
	b.WriteString(fmt.Sprintf("\t\tport = \"%s\"\n", cfg.Port))
	b.WriteString("\t}\n\n")
	b.WriteString(fmt.Sprintf("\tstore := %s.NewStore()\n\n", cfg.PackageName))
	b.WriteString("\tr := gin.Default()\n")
	b.WriteString(fmt.Sprintf("\t%s.RegisterRoutes(r, store)\n\n", cfg.PackageName))
	b.WriteString("\tlog.Printf(\"Mock server starting on :%s\", port)\n")
	b.WriteString("\tif err := r.Run(\":\" + port); err != nil {\n")
	b.WriteString("\t\tlog.Fatal(err)\n")
	b.WriteString("\t}\n")
	b.WriteString("}\n")

	return formatGoSource(b.String())
}

// GenerateGoMod produces a go.mod file for the generated mock server.
func GenerateGoMod(cfg Config) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("module %s\n\n", cfg.ModulePath))
	b.WriteString("go 1.21\n\n")
	b.WriteString("require github.com/gin-gonic/gin v1.9.1\n")
	return b.String()
}

package generate

import (
	"fmt"
	"strings"
)

// GenerateRoutes produces a route registration function that wires all
// generated handlers to their OpenAPI paths.
func GenerateRoutes(doc *Document, cfg Config) (string, error) {
	var b strings.Builder

	b.WriteString(fmt.Sprintf("package %s\n\n", cfg.PackageName))
	b.WriteString("import \"github.com/gin-gonic/gin\"\n\n")

	b.WriteString("// RegisterRoutes registers all generated mock handlers on the Gin engine.\n")
	b.WriteString("func RegisterRoutes(r *gin.Engine, store *Store) {\n")

	if cfg.Health {
		b.WriteString(fmt.Sprintf("\tr.GET(\"%s\", HealthHandler())\n", cfg.HealthPath))
	}

	for _, res := range doc.Resources {
		funcBase := goTypeName(res.Name)

		for _, op := range res.Operations {
			funcName := handlerFuncName(funcBase, op)
			ginPath := openAPIPathToGinPath(op.Path)
			b.WriteString(fmt.Sprintf("\tr.%s(\"%s\", %s(store))\n", op.Method, ginPath, funcName))
		}

		for _, action := range res.Actions {
			actionName := toPascalCase(action.Name)
			funcName := fmt.Sprintf("%s%sHandler", actionName, funcBase)
			ginPath := openAPIPathToGinPath(action.Path)
			b.WriteString(fmt.Sprintf("\tr.%s(\"%s\", %s(store))\n", action.Method, ginPath, funcName))
		}
	}

	b.WriteString("}\n")

	return formatGoSource(b.String())
}

// openAPIPathToGinPath converts OpenAPI path parameters {param} to Gin :param format.
func openAPIPathToGinPath(path string) string {
	result := path
	for {
		start := strings.Index(result, "{")
		if start == -1 {
			break
		}
		end := strings.Index(result, "}")
		if end == -1 {
			break
		}
		paramName := result[start+1 : end]
		result = result[:start] + ":" + paramName + result[end+1:]
	}
	return result
}

package generate

import (
	"fmt"
	"strings"
)

// GenerateHandlers produces Gin handler functions for each API operation.
func GenerateHandlers(doc *Document, cfg Config) (string, error) {
	var b strings.Builder

	b.WriteString(fmt.Sprintf("package %s\n\n", cfg.PackageName))
	b.WriteString("import (\n")
	b.WriteString("\t\"net/http\"\n")
	b.WriteString("\t\"strconv\"\n\n")
	b.WriteString("\t\"github.com/gin-gonic/gin\"\n")
	b.WriteString(")\n")

	for _, res := range doc.Resources {
		for _, op := range res.Operations {
			b.WriteString("\n")
			writeHandler(&b, res, op)
		}
		for _, action := range res.Actions {
			b.WriteString("\n")
			writeActionHandler(&b, res, action)
		}
	}

	if cfg.Health {
		b.WriteString("\n")
		writeHealthHandler(&b)
	}

	return formatGoSource(b.String())
}

func writeHandler(b *strings.Builder, res Resource, op ResourceOp) {
	funcBase := goTypeName(res.Name)
	goType := goTypeName(res.TypeName)
	funcName := handlerFuncName(funcBase, op)

	b.WriteString(fmt.Sprintf("func %s(store *Store) gin.HandlerFunc {\n", funcName))
	b.WriteString("\treturn func(c *gin.Context) {\n")

	switch op.Kind {
	case "create":
		writeCreateHandler(b, funcBase, goType, op)
	case "get":
		writeGetHandler(b, funcBase, goType, res)
	case "list":
		writeListHandler(b, funcBase, goType, op)
	case "update":
		writeUpdateHandler(b, funcBase, goType, res)
	case "delete":
		writeDeleteHandler(b, funcBase, res)
	}

	b.WriteString("\t}\n")
	b.WriteString("}\n")
}

func writeCreateHandler(b *strings.Builder, funcBase, goType string, op ResourceOp) {
	b.WriteString(fmt.Sprintf("\t\tvar req %s\n", goType))
	b.WriteString("\t\tif err := c.ShouldBindJSON(&req); err != nil {\n")
	b.WriteString("\t\t\tc.JSON(http.StatusBadRequest, gin.H{\"error\": err.Error()})\n")
	b.WriteString("\t\t\treturn\n")
	b.WriteString("\t\t}\n")
	b.WriteString(fmt.Sprintf("\t\tresult := store.Create%s(req)\n", funcBase))
	b.WriteString("\t\tc.JSON(http.StatusCreated, result)\n")
}

func writeGetHandler(b *strings.Builder, funcBase, goType string, res Resource) {
	idParam := res.IDParam
	if idParam == "" {
		idParam = "id"
	}
	b.WriteString(fmt.Sprintf("\t\tid := c.Param(\"%s\")\n", idParam))
	b.WriteString(fmt.Sprintf("\t\titem, ok := store.Get%s(id)\n", funcBase))
	b.WriteString("\t\tif !ok {\n")
	b.WriteString("\t\t\tc.JSON(http.StatusNotFound, gin.H{\"error\": \"not found\"})\n")
	b.WriteString("\t\t\treturn\n")
	b.WriteString("\t\t}\n")
	b.WriteString("\t\tc.JSON(http.StatusOK, item)\n")
}

func writeListHandler(b *strings.Builder, funcBase, goType string, op ResourceOp) {
	hasFilter := false
	filterParam := ""
	for _, p := range op.Parameters {
		if p.In == "query" && p.Name != "limit" && p.Name != "offset" && p.Name != "starting_after" && p.Name != "ending_before" {
			hasFilter = true
			filterParam = p.Name
			break
		}
	}

	b.WriteString("\t\tlimitStr := c.DefaultQuery(\"limit\", \"25\")\n")
	b.WriteString("\t\toffsetStr := c.DefaultQuery(\"offset\", \"0\")\n")
	b.WriteString("\t\tlimit, _ := strconv.Atoi(limitStr)\n")
	b.WriteString("\t\toffset, _ := strconv.Atoi(offsetStr)\n")
	b.WriteString("\t\tif limit <= 0 {\n")
	b.WriteString("\t\t\tlimit = 25\n")
	b.WriteString("\t\t}\n")

	if hasFilter {
		b.WriteString(fmt.Sprintf("\t\tfilterValue := c.Query(\"%s\")\n", filterParam))
		b.WriteString(fmt.Sprintf("\t\tdata, total := store.List%ss(limit, offset, \"%s\", filterValue)\n", funcBase, filterParam))
	} else {
		b.WriteString(fmt.Sprintf("\t\tdata, total := store.List%ss(limit, offset)\n", funcBase))
	}

	b.WriteString("\t\tc.JSON(http.StatusOK, gin.H{\n")
	b.WriteString("\t\t\t\"data\":     data,\n")
	b.WriteString("\t\t\t\"has_more\": offset+limit < total,\n")
	b.WriteString("\t\t\t\"total\":    total,\n")
	b.WriteString("\t\t})\n")
}

func writeUpdateHandler(b *strings.Builder, funcBase, goType string, res Resource) {
	idParam := res.IDParam
	if idParam == "" {
		idParam = "id"
	}
	b.WriteString(fmt.Sprintf("\t\tid := c.Param(\"%s\")\n", idParam))
	b.WriteString(fmt.Sprintf("\t\tvar req %s\n", goType))
	b.WriteString("\t\tif err := c.ShouldBindJSON(&req); err != nil {\n")
	b.WriteString("\t\t\tc.JSON(http.StatusBadRequest, gin.H{\"error\": err.Error()})\n")
	b.WriteString("\t\t\treturn\n")
	b.WriteString("\t\t}\n")
	b.WriteString(fmt.Sprintf("\t\tresult, ok := store.Update%s(id, req)\n", funcBase))
	b.WriteString("\t\tif !ok {\n")
	b.WriteString("\t\t\tc.JSON(http.StatusNotFound, gin.H{\"error\": \"not found\"})\n")
	b.WriteString("\t\t\treturn\n")
	b.WriteString("\t\t}\n")
	b.WriteString("\t\tc.JSON(http.StatusOK, result)\n")
}

func writeDeleteHandler(b *strings.Builder, funcBase string, res Resource) {
	idParam := res.IDParam
	if idParam == "" {
		idParam = "id"
	}
	b.WriteString(fmt.Sprintf("\t\tid := c.Param(\"%s\")\n", idParam))
	b.WriteString(fmt.Sprintf("\t\tok := store.Delete%s(id)\n", funcBase))
	b.WriteString("\t\tif !ok {\n")
	b.WriteString("\t\t\tc.JSON(http.StatusNotFound, gin.H{\"error\": \"not found\"})\n")
	b.WriteString("\t\t\treturn\n")
	b.WriteString("\t\t}\n")
	b.WriteString("\t\tc.Status(http.StatusNoContent)\n")
}

func writeActionHandler(b *strings.Builder, res Resource, action ResourceAction) {
	funcBase := goTypeName(res.Name)
	actionName := toPascalCase(action.Name)
	funcName := fmt.Sprintf("%s%sHandler", actionName, funcBase)

	idParam := res.IDParam
	if idParam == "" {
		idParam = "id"
	}

	b.WriteString(fmt.Sprintf("func %s(store *Store) gin.HandlerFunc {\n", funcName))
	b.WriteString("\treturn func(c *gin.Context) {\n")
	b.WriteString(fmt.Sprintf("\t\tid := c.Param(\"%s\")\n", idParam))
	b.WriteString(fmt.Sprintf("\t\tresult, ok := store.%s%s(id)\n", actionName, funcBase))
	b.WriteString("\t\tif !ok {\n")
	b.WriteString("\t\t\tc.JSON(http.StatusNotFound, gin.H{\"error\": \"not found\"})\n")
	b.WriteString("\t\t\treturn\n")
	b.WriteString("\t\t}\n")
	b.WriteString("\t\tc.JSON(http.StatusOK, result)\n")
	b.WriteString("\t}\n")
	b.WriteString("}\n")
}

func writeHealthHandler(b *strings.Builder) {
	b.WriteString("// HealthHandler returns a simple health check endpoint.\n")
	b.WriteString("func HealthHandler() gin.HandlerFunc {\n")
	b.WriteString("\treturn func(c *gin.Context) {\n")
	b.WriteString("\t\tc.JSON(http.StatusOK, gin.H{\"status\": \"ok\"})\n")
	b.WriteString("\t}\n")
	b.WriteString("}\n")
}

func handlerFuncName(funcBase string, op ResourceOp) string {
	switch op.Kind {
	case "list":
		return fmt.Sprintf("List%ssHandler", funcBase)
	case "create":
		return fmt.Sprintf("Create%sHandler", funcBase)
	case "get":
		return fmt.Sprintf("Get%sHandler", funcBase)
	case "update":
		return fmt.Sprintf("Update%sHandler", funcBase)
	case "delete":
		return fmt.Sprintf("Delete%sHandler", funcBase)
	default:
		return fmt.Sprintf("%s%sHandler", toPascalCase(op.Kind), funcBase)
	}
}

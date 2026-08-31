package generate

import (
	"fmt"
	"strings"
)

// GenerateTypes produces Go type definitions from the document's component schemas.
func GenerateTypes(doc *Document, packageName string) (string, error) {
	var b strings.Builder

	b.WriteString(fmt.Sprintf("package %s\n\n", packageName))

	needsTime := schemasNeedTime(doc)
	if needsTime {
		b.WriteString("import \"time\"\n\n")
	}

	for i, ns := range doc.Schemas {
		if !ns.Schema.HasType("object") {
			continue
		}
		if i > 0 {
			b.WriteString("\n")
		}
		writeStructType(&b, ns.Name, ns.Schema)
	}

	return formatGoSource(b.String())
}

func writeStructType(b *strings.Builder, name string, s Schema) {
	typeName := goTypeName(name)
	b.WriteString(fmt.Sprintf("type %s struct {\n", typeName))

	for _, prop := range s.Properties {
		fieldName := goTypeName(prop.Name)
		fieldType := schemaToGoType(prop.Schema, prop.Ref)

		isRequired := s.Required[prop.Name]
		jsonTag := prop.Name

		if !isRequired && !prop.Schema.ReadOnly {
			if !strings.HasPrefix(fieldType, "[]") && !strings.HasPrefix(fieldType, "map[") {
				fieldType = "*" + fieldType
			}
			jsonTag += ",omitempty"
		}

		b.WriteString(fmt.Sprintf("\t%s %s `json:\"%s\"`\n", fieldName, fieldType, jsonTag))
	}

	b.WriteString("}\n")
}

func schemasNeedTime(doc *Document) bool {
	for _, ns := range doc.Schemas {
		if schemaUsesTime(ns.Schema) {
			return true
		}
	}
	return false
}

func schemaUsesTime(s Schema) bool {
	for _, prop := range s.Properties {
		if prop.Schema.Format == "date-time" {
			return true
		}
	}
	return false
}

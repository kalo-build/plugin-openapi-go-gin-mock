package generate

import (
	"go/format"
	"strings"
	"unicode"
)

func formatGoSource(src string) (string, error) {
	formatted, err := format.Source([]byte(src))
	if err != nil {
		return src, err
	}
	return string(formatted), nil
}

func toPascalCase(s string) string {
	parts := splitIdentifier(s)
	var b strings.Builder
	for _, part := range parts {
		if part == "" {
			continue
		}
		b.WriteString(strings.ToUpper(part[:1]) + strings.ToLower(part[1:]))
	}
	return b.String()
}

func toCamelCase(s string) string {
	parts := splitIdentifier(s)
	var b strings.Builder
	for i, part := range parts {
		if part == "" {
			continue
		}
		if i == 0 {
			b.WriteString(strings.ToLower(part))
		} else {
			b.WriteString(strings.ToUpper(part[:1]) + strings.ToLower(part[1:]))
		}
	}
	return b.String()
}

func toSnakeCase(s string) string {
	var b strings.Builder
	for i, r := range s {
		if unicode.IsUpper(r) && i > 0 {
			b.WriteByte('_')
		}
		b.WriteRune(unicode.ToLower(r))
	}
	return b.String()
}

func splitIdentifier(s string) []string {
	s = strings.ReplaceAll(s, "-", "_")
	var parts []string
	for _, seg := range strings.Split(s, "_") {
		parts = append(parts, splitCamel(seg)...)
	}
	return parts
}

func splitCamel(s string) []string {
	if s == "" {
		return nil
	}
	var parts []string
	start := 0
	for i := 1; i < len(s); i++ {
		if unicode.IsUpper(rune(s[i])) && !unicode.IsUpper(rune(s[i-1])) {
			parts = append(parts, s[start:i])
			start = i
		}
	}
	parts = append(parts, s[start:])
	return parts
}

func refToTypeName(ref string) string {
	parts := strings.Split(ref, "/")
	return parts[len(parts)-1]
}

func toSingular(s string) string {
	if strings.HasSuffix(s, "ies") {
		return s[:len(s)-3] + "y"
	}
	if strings.HasSuffix(s, "ses") || strings.HasSuffix(s, "xes") || strings.HasSuffix(s, "zes") {
		return s[:len(s)-2]
	}
	if strings.HasSuffix(s, "s") && !strings.HasSuffix(s, "ss") {
		return s[:len(s)-1]
	}
	return s
}

// idPrefixFromName generates a short lowercase prefix from a resource name.
// "Customer" → "cus", "Invoice" → "inv", "Payment" → "pay"
func idPrefixFromName(name string) string {
	lower := strings.ToLower(name)
	if len(lower) <= 3 {
		return lower
	}
	return lower[:3]
}

func goTypeName(name string) string {
	return toPascalCase(name)
}

// schemaToGoType maps an OpenAPI schema to a Go type string.
func schemaToGoType(s Schema, ref string) string {
	if ref != "" {
		return goTypeName(refToTypeName(ref))
	}

	if s.HasType("array") && s.Items != nil {
		itemType := schemaToGoType(*s.Items.Schema, s.Items.Ref)
		return "[]" + itemType
	}

	if s.HasType("object") {
		if len(s.Properties) > 0 {
			return "map[string]interface{}"
		}
		return "map[string]interface{}"
	}

	if s.HasType("integer") {
		switch s.Format {
		case "int64":
			return "int64"
		default:
			return "int"
		}
	}

	if s.HasType("number") {
		return "float64"
	}

	if s.HasType("boolean") {
		return "bool"
	}

	if s.HasType("string") {
		switch s.Format {
		case "date-time":
			return "time.Time"
		default:
			return "string"
		}
	}

	return "interface{}"
}

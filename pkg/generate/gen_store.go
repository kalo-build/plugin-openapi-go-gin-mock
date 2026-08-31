package generate

import (
	"fmt"
	"strings"
)

// GenerateStore produces a thread-safe in-memory data store with CRUD
// methods for each API resource derived from the OpenAPI spec.
func GenerateStore(doc *Document, cfg Config) (string, error) {
	var b strings.Builder

	b.WriteString(fmt.Sprintf("package %s\n\n", cfg.PackageName))
	b.WriteString("import (\n")
	b.WriteString("\t\"fmt\"\n")
	b.WriteString("\t\"math/rand\"\n")
	b.WriteString("\t\"sort\"\n")
	b.WriteString("\t\"sync\"\n")
	if resourcesNeedTime(doc) {
		b.WriteString("\t\"time\"\n")
	}
	b.WriteString(")\n\n")

	writeStoreStruct(&b, doc)
	b.WriteString("\n")
	writeNewStore(&b, doc)
	b.WriteString("\n")
	writeIDGenerator(&b, cfg.IDPrefix)

	for _, res := range doc.Resources {
		b.WriteString("\n")
		writeResourceStoreMethods(&b, res, cfg.IDPrefix)
	}

	b.WriteString("\n")
	writeResetMethod(&b, doc)

	return formatGoSource(b.String())
}

func writeStoreStruct(b *strings.Builder, doc *Document) {
	b.WriteString("// Store provides thread-safe in-memory storage for all API resources.\n")
	b.WriteString("type Store struct {\n")
	b.WriteString("\tmu sync.RWMutex\n")
	for _, res := range doc.Resources {
		fieldName := toCamelCase(res.Name) + "s"
		goType := goTypeName(res.TypeName)
		b.WriteString(fmt.Sprintf("\t%s map[string]%s\n", fieldName, goType))
	}
	b.WriteString("}\n")
}

func writeNewStore(b *strings.Builder, doc *Document) {
	b.WriteString("// NewStore creates a new empty Store.\n")
	b.WriteString("func NewStore() *Store {\n")
	b.WriteString("\treturn &Store{\n")
	for _, res := range doc.Resources {
		fieldName := toCamelCase(res.Name) + "s"
		goType := goTypeName(res.TypeName)
		b.WriteString(fmt.Sprintf("\t\t%s: make(map[string]%s),\n", fieldName, goType))
	}
	b.WriteString("\t}\n")
	b.WriteString("}\n")
}

func writeIDGenerator(b *strings.Builder, idPrefix bool) {
	b.WriteString("func generateID(prefix string) string {\n")
	b.WriteString("\tconst charset = \"abcdefghijklmnopqrstuvwxyz0123456789\"\n")
	b.WriteString("\tid := make([]byte, 14)\n")
	b.WriteString("\tfor i := range id {\n")
	b.WriteString("\t\tid[i] = charset[rand.Intn(len(charset))]\n")
	b.WriteString("\t}\n")
	if idPrefix {
		b.WriteString("\treturn fmt.Sprintf(\"%s_%s\", prefix, string(id))\n")
	} else {
		b.WriteString("\treturn string(id)\n")
	}
	b.WriteString("}\n")
}

func writeResourceStoreMethods(b *strings.Builder, res Resource, idPrefix bool) {
	funcBase := goTypeName(res.Name)
	goType := goTypeName(res.TypeName)
	fieldName := toCamelCase(res.Name) + "s"
	prefix := idPrefixFromName(res.Name)

	for _, op := range res.Operations {
		switch op.Kind {
		case "create":
			writeCreateMethod(b, funcBase, goType, fieldName, prefix, res.StatusDefault)
		case "get":
			writeGetMethod(b, funcBase, goType, fieldName)
		case "list":
			writeListMethod(b, funcBase, goType, fieldName, op)
		case "update":
			writeUpdateMethod(b, funcBase, goType, fieldName)
		case "delete":
			writeDeleteMethod(b, funcBase, goType, fieldName)
		}
	}

	for _, action := range res.Actions {
		writeActionMethod(b, funcBase, goType, fieldName, action)
	}
}

func writeCreateMethod(b *strings.Builder, funcBase, goType, fieldName, prefix, statusDefault string) {
	b.WriteString(fmt.Sprintf("func (s *Store) Create%s(item %s) %s {\n", funcBase, goType, goType))
	b.WriteString("\ts.mu.Lock()\n")
	b.WriteString("\tdefer s.mu.Unlock()\n")
	b.WriteString(fmt.Sprintf("\tif item.Id == \"\" {\n"))
	b.WriteString(fmt.Sprintf("\t\titem.Id = generateID(\"%s\")\n", prefix))
	b.WriteString("\t}\n")
	if statusDefault != "" {
		b.WriteString("\tif item.Status == \"\" {\n")
		b.WriteString(fmt.Sprintf("\t\titem.Status = \"%s\"\n", statusDefault))
		b.WriteString("\t}\n")
	}
	b.WriteString(fmt.Sprintf("\titem.CreatedAt = time.Now().UTC()\n"))
	b.WriteString(fmt.Sprintf("\ts.%s[item.Id] = item\n", fieldName))
	b.WriteString("\treturn item\n")
	b.WriteString("}\n\n")
}

func writeGetMethod(b *strings.Builder, funcBase, goType, fieldName string) {
	b.WriteString(fmt.Sprintf("func (s *Store) Get%s(id string) (%s, bool) {\n", funcBase, goType))
	b.WriteString("\ts.mu.RLock()\n")
	b.WriteString("\tdefer s.mu.RUnlock()\n")
	b.WriteString(fmt.Sprintf("\titem, ok := s.%s[id]\n", fieldName))
	b.WriteString("\treturn item, ok\n")
	b.WriteString("}\n\n")
}

func writeListMethod(b *strings.Builder, funcBase, goType, fieldName string, op ResourceOp) {
	hasFilterParam := false
	filterField := ""
	for _, p := range op.Parameters {
		if p.In == "query" && p.Name != "limit" && p.Name != "offset" && p.Name != "starting_after" && p.Name != "ending_before" {
			hasFilterParam = true
			filterField = p.Name
			break
		}
	}

	if hasFilterParam {
		b.WriteString(fmt.Sprintf("func (s *Store) List%ss(limit, offset int, filterField, filterValue string) ([]%s, int) {\n", funcBase, goType))
	} else {
		b.WriteString(fmt.Sprintf("func (s *Store) List%ss(limit, offset int) ([]%s, int) {\n", funcBase, goType))
	}
	b.WriteString("\ts.mu.RLock()\n")
	b.WriteString("\tdefer s.mu.RUnlock()\n")
	b.WriteString(fmt.Sprintf("\tvar all []%s\n", goType))
	b.WriteString(fmt.Sprintf("\tfor _, item := range s.%s {\n", fieldName))

	if hasFilterParam {
		goField := goTypeName(filterField)
		b.WriteString(fmt.Sprintf("\t\tif filterValue != \"\" && item.%s != filterValue {\n", goField))
		b.WriteString("\t\t\tcontinue\n")
		b.WriteString("\t\t}\n")
	}

	b.WriteString("\t\tall = append(all, item)\n")
	b.WriteString("\t}\n")
	b.WriteString("\tsort.Slice(all, func(i, j int) bool { return all[i].Id < all[j].Id })\n")
	b.WriteString("\ttotal := len(all)\n")
	b.WriteString("\tif offset >= total {\n")
	b.WriteString(fmt.Sprintf("\t\treturn []%s{}, total\n", goType))
	b.WriteString("\t}\n")
	b.WriteString("\tend := offset + limit\n")
	b.WriteString("\tif end > total {\n")
	b.WriteString("\t\tend = total\n")
	b.WriteString("\t}\n")
	b.WriteString("\treturn all[offset:end], total\n")
	b.WriteString("}\n\n")
}

func writeUpdateMethod(b *strings.Builder, funcBase, goType, fieldName string) {
	b.WriteString(fmt.Sprintf("func (s *Store) Update%s(id string, updates %s) (%s, bool) {\n", funcBase, goType, goType))
	b.WriteString("\ts.mu.Lock()\n")
	b.WriteString("\tdefer s.mu.Unlock()\n")
	b.WriteString(fmt.Sprintf("\texisting, ok := s.%s[id]\n", fieldName))
	b.WriteString("\tif !ok {\n")
	b.WriteString(fmt.Sprintf("\t\treturn %s{}, false\n", goType))
	b.WriteString("\t}\n")
	b.WriteString("\tupdates.Id = existing.Id\n")
	b.WriteString("\tupdates.CreatedAt = existing.CreatedAt\n")
	b.WriteString(fmt.Sprintf("\ts.%s[id] = updates\n", fieldName))
	b.WriteString(fmt.Sprintf("\treturn s.%s[id], true\n", fieldName))
	b.WriteString("}\n\n")
}

func writeDeleteMethod(b *strings.Builder, funcBase, goType, fieldName string) {
	b.WriteString(fmt.Sprintf("func (s *Store) Delete%s(id string) bool {\n", funcBase))
	b.WriteString("\ts.mu.Lock()\n")
	b.WriteString("\tdefer s.mu.Unlock()\n")
	b.WriteString(fmt.Sprintf("\tif _, ok := s.%s[id]; !ok {\n", fieldName))
	b.WriteString("\t\treturn false\n")
	b.WriteString("\t}\n")
	b.WriteString(fmt.Sprintf("\tdelete(s.%s, id)\n", fieldName))
	b.WriteString("\treturn true\n")
	b.WriteString("}\n\n")
}

func writeActionMethod(b *strings.Builder, funcBase, goType, fieldName string, action ResourceAction) {
	actionName := toPascalCase(action.Name)
	b.WriteString(fmt.Sprintf("func (s *Store) %s%s(id string) (%s, bool) {\n", actionName, funcBase, goType))
	b.WriteString("\ts.mu.Lock()\n")
	b.WriteString("\tdefer s.mu.Unlock()\n")
	b.WriteString(fmt.Sprintf("\titem, ok := s.%s[id]\n", fieldName))
	b.WriteString("\tif !ok {\n")
	b.WriteString(fmt.Sprintf("\t\treturn %s{}, false\n", goType))
	b.WriteString("\t}\n")
	if action.StatusValue != "" {
		b.WriteString(fmt.Sprintf("\titem.Status = \"%s\"\n", action.StatusValue))
	}
	b.WriteString(fmt.Sprintf("\ts.%s[id] = item\n", fieldName))
	b.WriteString("\treturn item, true\n")
	b.WriteString("}\n\n")
}

func writeResetMethod(b *strings.Builder, doc *Document) {
	b.WriteString("// Reset clears all data from the store.\n")
	b.WriteString("func (s *Store) Reset() {\n")
	b.WriteString("\ts.mu.Lock()\n")
	b.WriteString("\tdefer s.mu.Unlock()\n")
	for _, res := range doc.Resources {
		fieldName := toCamelCase(res.Name) + "s"
		goType := goTypeName(res.TypeName)
		b.WriteString(fmt.Sprintf("\ts.%s = make(map[string]%s)\n", fieldName, goType))
	}
	b.WriteString("}\n")
}

func resourcesNeedTime(doc *Document) bool {
	for _, res := range doc.Resources {
		for _, op := range res.Operations {
			if op.Kind == "create" {
				return true
			}
		}
	}
	return false
}

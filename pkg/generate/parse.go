package generate

import (
	"fmt"
	"sort"
	"strings"

	"github.com/pb33f/libopenapi"
	"github.com/pb33f/libopenapi/datamodel/high/base"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
	"github.com/pb33f/libopenapi/orderedmap"
)

// ParseSpec parses an OpenAPI YAML/JSON spec and converts it to an intermediate representation.
func ParseSpec(specBytes []byte) (*Document, error) {
	document, err := libopenapi.NewDocument(specBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse OpenAPI document: %w", err)
	}

	model, errs := document.BuildV3Model()
	if errs != nil {
		return nil, fmt.Errorf("failed to build V3 model: %v", errs)
	}

	doc := convertDocument(&model.Model)
	doc.Resources = deriveResources(doc)
	return doc, nil
}

func convertDocument(v3doc *v3.Document) *Document {
	doc := &Document{}

	if v3doc.Info != nil {
		doc.Info = DocInfo{
			Title:   v3doc.Info.Title,
			Version: v3doc.Info.Version,
		}
	}

	doc.Schemas = convertSchemas(v3doc)
	doc.Paths = convertPaths(v3doc)

	return doc
}

func convertSchemas(v3doc *v3.Document) []NamedSchema {
	if v3doc.Components == nil || v3doc.Components.Schemas == nil {
		return nil
	}

	var schemas []NamedSchema
	for pair := v3doc.Components.Schemas.First(); pair != nil; pair = pair.Next() {
		s := convertSchemaProxy(pair.Value())
		schemas = append(schemas, NamedSchema{Name: pair.Key(), Schema: s})
	}

	sort.Slice(schemas, func(i, j int) bool {
		return schemas[i].Name < schemas[j].Name
	})
	return schemas
}

func convertPaths(v3doc *v3.Document) []PathEntry {
	if v3doc.Paths == nil || v3doc.Paths.PathItems == nil {
		return nil
	}

	var paths []PathEntry
	for pair := v3doc.Paths.PathItems.First(); pair != nil; pair = pair.Next() {
		entry := convertPathItem(pair.Key(), pair.Value())
		if len(entry.Operations) > 0 {
			paths = append(paths, entry)
		}
	}

	sort.Slice(paths, func(i, j int) bool {
		return paths[i].Path < paths[j].Path
	})
	return paths
}

func convertPathItem(path string, item *v3.PathItem) PathEntry {
	entry := PathEntry{Path: path}

	type methodOp struct {
		method string
		op     *v3.Operation
	}
	ops := []methodOp{
		{"GET", item.Get},
		{"POST", item.Post},
		{"PUT", item.Put},
		{"PATCH", item.Patch},
		{"DELETE", item.Delete},
	}

	for _, mo := range ops {
		if mo.op != nil {
			entry.Operations = append(entry.Operations, convertOperation(mo.method, mo.op))
		}
	}

	return entry
}

func convertOperation(method string, op *v3.Operation) Operation {
	o := Operation{
		Method:      method,
		OperationID: op.OperationId,
		Summary:     op.Summary,
		Tags:        op.Tags,
		Responses:   make(map[string]*SchemaRef),
	}

	for _, p := range op.Parameters {
		o.Parameters = append(o.Parameters, convertParameter(p))
	}

	if op.RequestBody != nil {
		o.RequestBody = extractMediaTypeSchema(op.RequestBody.Content)
	}

	if op.Responses != nil && op.Responses.Codes != nil {
		for pair := op.Responses.Codes.First(); pair != nil; pair = pair.Next() {
			resp := pair.Value()
			if resp != nil {
				ref := extractMediaTypeSchema(resp.Content)
				if ref != nil {
					o.Responses[pair.Key()] = ref
				}
			}
		}
	}

	return o
}

func convertParameter(p *v3.Parameter) Parameter {
	param := Parameter{
		Name: p.Name,
		In:   p.In,
	}
	if p.Required != nil {
		param.Required = *p.Required
	}
	if p.Schema != nil {
		s := convertSchemaProxy(p.Schema)
		param.Schema = &s
	}
	return param
}

func extractMediaTypeSchema(content *orderedmap.Map[string, *v3.MediaType]) *SchemaRef {
	if content == nil {
		return nil
	}
	jsonMT, ok := content.Get("application/json")
	if !ok || jsonMT == nil || jsonMT.Schema == nil {
		return nil
	}
	ref := jsonMT.Schema.GetReference()
	if ref != "" {
		return &SchemaRef{Ref: ref}
	}
	s := convertSchemaProxy(jsonMT.Schema)
	return &SchemaRef{Schema: &s}
}

func convertSchemaProxy(proxy *base.SchemaProxy) Schema {
	if proxy == nil {
		return Schema{}
	}
	built := proxy.Schema()
	if built == nil {
		return Schema{}
	}
	return convertSchema(built)
}

func convertSchema(s *base.Schema) Schema {
	schema := Schema{
		Types:    s.Type,
		Format:   s.Format,
		Required: make(map[string]bool),
	}

	if s.ReadOnly != nil && *s.ReadOnly {
		schema.ReadOnly = true
	}

	for _, e := range s.Enum {
		if e != nil {
			schema.Enum = append(schema.Enum, e.Value)
		}
	}

	for _, r := range s.Required {
		schema.Required[r] = true
	}

	if s.Properties != nil {
		for pair := s.Properties.First(); pair != nil; pair = pair.Next() {
			np := NamedProperty{Name: pair.Key()}
			if pair.Value() != nil {
				np.Ref = pair.Value().GetReference()
				np.Schema = convertSchemaProxy(pair.Value())
			}
			schema.Properties = append(schema.Properties, np)
		}
		sort.Slice(schema.Properties, func(i, j int) bool {
			return schema.Properties[i].Name < schema.Properties[j].Name
		})
	}

	if s.Items != nil && s.Items.A != nil {
		itemRef := s.Items.A.GetReference()
		itemSchema := convertSchemaProxy(s.Items.A)
		schema.Items = &SchemaRef{Ref: itemRef, Schema: &itemSchema}
	}

	return schema
}

// deriveResources groups path entries into logical API resources.
func deriveResources(doc *Document) []Resource {
	type resourceKey struct {
		tag      string
		basePath string
	}
	resMap := make(map[resourceKey]*Resource)
	var order []resourceKey

	for _, pe := range doc.Paths {
		for _, op := range pe.Operations {
			tag := ""
			if len(op.Tags) > 0 {
				tag = op.Tags[0]
			}

			basePath, idParam, actionName := classifyPath(pe.Path)
			key := resourceKey{tag: tag, basePath: basePath}

			res, exists := resMap[key]
			if !exists {
				name := tag
				if name == "" {
					name = resourceNameFromPath(basePath)
				}
				res = &Resource{
					Name:     name,
					Tag:      tag,
					BasePath: basePath,
					IDParam:  idParam,
				}
				resMap[key] = res
				order = append(order, key)
			}

			if idParam != "" && res.IDParam == "" {
				res.IDParam = idParam
			}

			if actionName != "" {
				reqRef := schemaRefName(op.RequestBody)
				respRef := responseRefName(op)
				res.Actions = append(res.Actions, ResourceAction{
					Name:        actionName,
					Method:      op.Method,
					Path:        pe.Path,
					OperationID: op.OperationID,
					RequestRef:  reqRef,
					ResponseRef: respRef,
				})
				continue
			}

			kind := classifyOperation(op.Method, pe.Path, basePath)
			reqRef := schemaRefName(op.RequestBody)
			respRef := responseRefName(op)

			res.Operations = append(res.Operations, ResourceOp{
				Kind:        kind,
				Method:      op.Method,
				Path:        pe.Path,
				OperationID: op.OperationID,
				RequestRef:  reqRef,
				ResponseRef: respRef,
				Parameters:  op.Parameters,
			})
		}
	}

	var resources []Resource
	for _, key := range order {
		res := resMap[key]
		resolveResourceTypeInfo(doc, res)
		resources = append(resources, *res)
	}
	return resources
}

func resolveResourceTypeInfo(doc *Document, res *Resource) {
	for _, op := range res.Operations {
		if op.ResponseRef != "" {
			res.TypeName = op.ResponseRef
			break
		}
	}
	if res.TypeName == "" {
		res.TypeName = res.Name
	}

	statusEnum := findStatusEnum(doc, res.TypeName)
	if len(statusEnum) == 0 {
		return
	}
	res.StatusDefault = statusEnum[0]
	for i := range res.Actions {
		res.Actions[i].StatusValue = matchActionToEnum(res.Actions[i].Name, statusEnum)
	}
}

func findStatusEnum(doc *Document, typeName string) []string {
	for _, ns := range doc.Schemas {
		if ns.Name == typeName {
			for _, prop := range ns.Schema.Properties {
				if prop.Name == "status" && len(prop.Schema.Enum) > 0 {
					return prop.Schema.Enum
				}
			}
			return nil
		}
	}
	return nil
}

func matchActionToEnum(actionName string, enumValues []string) string {
	lower := strings.ToLower(actionName)
	for _, v := range enumValues {
		if strings.ToLower(v) == lower {
			return v
		}
	}
	for _, v := range enumValues {
		if strings.HasPrefix(strings.ToLower(v), lower) {
			return v
		}
	}
	return actionName
}

func classifyPath(path string) (basePath, idParam, action string) {
	segments := strings.Split(strings.Trim(path, "/"), "/")
	var base []string
	for i, seg := range segments {
		if strings.HasPrefix(seg, "{") && strings.HasSuffix(seg, "}") {
			idParam = seg[1 : len(seg)-1]
			base = segments[:i]
			if i+1 < len(segments) {
				action = segments[i+1]
			}
			return "/" + strings.Join(base, "/"), idParam, action
		}
	}
	return "/" + strings.Join(segments, "/"), "", ""
}

func classifyOperation(method, fullPath, basePath string) string {
	hasID := fullPath != basePath
	switch method {
	case "GET":
		if hasID {
			return "get"
		}
		return "list"
	case "POST":
		return "create"
	case "PUT", "PATCH":
		return "update"
	case "DELETE":
		return "delete"
	default:
		return strings.ToLower(method)
	}
}

func resourceNameFromPath(basePath string) string {
	segments := strings.Split(strings.Trim(basePath, "/"), "/")
	if len(segments) == 0 {
		return "Resource"
	}
	last := segments[len(segments)-1]
	return toPascalCase(toSingular(last))
}

func schemaRefName(ref *SchemaRef) string {
	if ref == nil {
		return ""
	}
	if ref.Ref != "" {
		return refToTypeName(ref.Ref)
	}
	return ""
}

func responseRefName(op Operation) string {
	for _, code := range []string{"200", "201"} {
		if r, ok := op.Responses[code]; ok {
			if r.Ref != "" {
				return refToTypeName(r.Ref)
			}
			if r.Schema != nil && r.Schema.HasType("object") {
				for _, prop := range r.Schema.Properties {
					if prop.Name == "data" && prop.Schema.HasType("array") && prop.Schema.Items != nil && prop.Schema.Items.Ref != "" {
						return refToTypeName(prop.Schema.Items.Ref)
					}
				}
			}
		}
	}
	return ""
}

package generate

// Document is a simplified OpenAPI 3.x document for mock server generation.
type Document struct {
	Info      DocInfo
	Schemas   []NamedSchema
	Paths     []PathEntry
	Resources []Resource
}

type DocInfo struct {
	Title   string
	Version string
}

type NamedSchema struct {
	Name   string
	Schema Schema
}

type Schema struct {
	Types      []string
	Format     string
	Properties []NamedProperty
	Required   map[string]bool
	Items      *SchemaRef
	ReadOnly   bool
	Enum       []string
}

func (s Schema) HasType(t string) bool {
	for _, st := range s.Types {
		if st == t {
			return true
		}
	}
	return false
}

type NamedProperty struct {
	Name   string
	Ref    string
	Schema Schema
}

type SchemaRef struct {
	Ref    string
	Schema *Schema
}

type PathEntry struct {
	Path       string
	Operations []Operation
}

type Operation struct {
	Method      string
	OperationID string
	Tags        []string
	Summary     string
	Parameters  []Parameter
	RequestBody *SchemaRef
	Responses   map[string]*SchemaRef
}

type Parameter struct {
	Name     string
	In       string
	Required bool
	Schema   *Schema
}

// Resource represents a grouped API resource derived from paths.
type Resource struct {
	Name          string
	Tag           string
	BasePath      string
	IDParam       string
	TypeName      string // Go type name resolved from response schema ref
	StatusDefault string // Default status value (first enum value of status field)
	Operations    []ResourceOp
	Actions       []ResourceAction
}

// ResourceOp is a standard CRUD operation on a resource.
type ResourceOp struct {
	Kind        string // "list", "create", "get", "update", "delete"
	Method      string
	Path        string
	OperationID string
	RequestRef  string
	ResponseRef string
	Parameters  []Parameter
}

// ResourceAction is a non-CRUD action on a specific resource instance (e.g. /invoices/{id}/finalize).
type ResourceAction struct {
	Name        string
	Method      string
	Path        string
	OperationID string
	RequestRef  string
	ResponseRef string
	StatusValue string // Resolved status enum value for this action
}

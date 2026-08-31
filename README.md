# plugin-openapi-go-gin-mock

Generates a standalone Go mock server with Gin handlers and in-memory storage from an OpenAPI 3.1 specification. Produces a complete, runnable server project with typed models, thread-safe CRUD storage, handler functions, route registration, and a `main.go` entry point.

## What it generates

| OpenAPI artifact | Go output |
|------------------|-----------|
| **Component schemas** | Typed Go structs with JSON tags |
| **Path operations** | Gin handler functions (list, create, get, update, delete) |
| **Path actions** | Action handlers (e.g. `/invoices/{id}/finalize`) |
| **All paths** | Route registration function |
| **Full spec** | In-memory store with CRUD methods, server `main.go`, `go.mod` |

### Example output

Given an OpenAPI spec with a `Customer` schema and CRUD paths:

**types.go:**

```go
package mock

import "time"

type Customer struct {
    CreatedAt time.Time              `json:"created_at"`
    Email     string                 `json:"email"`
    Id        string                 `json:"id"`
    Metadata  map[string]interface{} `json:"metadata,omitempty"`
    Name      string                 `json:"name"`
}
```

**store.go** (excerpt):

```go
func (s *Store) CreateCustomer(item Customer) Customer {
    s.mu.Lock()
    defer s.mu.Unlock()
    if item.Id == "" {
        item.Id = generateID("cus")
    }
    item.CreatedAt = time.Now().UTC()
    s.customers[item.Id] = item
    return item
}
```

**handlers.go** (excerpt):

```go
func CreateCustomerHandler(store *Store) gin.HandlerFunc {
    return func(c *gin.Context) {
        var req Customer
        if err := c.ShouldBindJSON(&req); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
            return
        }
        result := store.CreateCustomer(req)
        c.JSON(http.StatusCreated, result)
    }
}
```

**routes.go:**

```go
func RegisterRoutes(r *gin.Engine, store *Store) {
    r.GET("/v1/customers", ListCustomersHandler(store))
    r.POST("/v1/customers", CreateCustomerHandler(store))
    r.GET("/v1/customers/:id", GetCustomerHandler(store))
    r.PUT("/v1/customers/:id", UpdateCustomerHandler(store))
}
```

### Features

- Prefixed IDs (e.g. `cus_`, `inv_`, `pay_`) matching Stripe conventions
- Thread-safe `sync.RWMutex`-based store
- Offset-based pagination with `limit`, `offset`, `has_more`, `total`
- Query parameter filtering (e.g. `?customer_id=cus_abc`)
- Action endpoints (e.g. `POST /invoices/{id}/finalize`)
- Store `Reset()` method for clearing data between test runs

### Type mappings

| OpenAPI type | Go type |
|--------------|---------|
| `string` | `string` |
| `string` + `date-time` | `time.Time` |
| `integer` | `int` |
| `integer` + `int64` | `int64` |
| `number` | `float64` |
| `boolean` | `bool` |
| `object` (no properties) | `map[string]interface{}` |
| `array` | `[]T` |
| `$ref` | Named struct |

## Input / output

| Direction | Format | Store suggestion | Description |
|-----------|--------|------------------|-------------|
| Input | `KA:OA1:YAML1` | `KA_OA_YAML` | OpenAPI 3.1 YAML specification |
| Output | `KA:OA1:GO_GIN_MOCK1` | `KA_OA_GO_GIN_MOCK` | Generated Go mock server source files |

## Configuration

| Key | Type | Required | Default | Description |
|-----|------|----------|---------|-------------|
| `config.specFileName` | string | no | `"openapi.yaml"` | Name of the OpenAPI spec file to read |
| `config.packageName` | string | no | `"mock"` | Go package name for the generated server package |
| `config.modulePath` | string | no | `"mock-server"` | Go module path for the generated server |
| `config.port` | string | no | `"8081"` | Default server port |
| `config.idPrefix` | boolean | no | `true` | Generate resource-based ID prefixes (e.g. `cus_`, `inv_`) |

## Pipeline context

```yaml
stores:
  KA_OA_YAML:
    format: "KA:OA1:YAML1"
    type: "localFileSystem"
    options:
      path: "./docs/openapi"

  KA_OA_GO_GIN_MOCK:
    format: "KA:OA1:GO_GIN_MOCK1"
    type: "localFileSystem"
    options:
      path: "./generated/mock"

plugins:
  "@kalo-build/plugin-openapi-go-gin-mock":
    version: "v1.0.0"
    inputs:
      openapi:
        format: "KA:OA1:YAML1"
        store: "KA_OA_YAML"
    output:
      format: "KA:OA1:GO_GIN_MOCK1"
      store: "KA_OA_GO_GIN_MOCK"
    config:
      specFileName: "openapi.yaml"
      packageName: "mock"
      modulePath: "stripe-lite-mock"
      port: "8081"
      idPrefix: true
```

## Project structure

```
plugin-openapi-go-gin-mock/
├── cmd/plugin/             # WASM entry point
│   └── main.go
├── pkg/generate/           # Generation logic
│   ├── generate.go         # GenerateMockServer entry point
│   ├── config.go           # Configuration struct
│   ├── types.go            # Intermediate types (Document, Resource, etc.)
│   ├── parse.go            # OpenAPI parsing (libopenapi)
│   ├── gen_types.go        # Generate Go struct types
│   ├── gen_store.go        # Generate in-memory store
│   ├── gen_handlers.go     # Generate Gin handlers
│   ├── gen_routes.go       # Generate route registration
│   ├── gen_main.go         # Generate server main.go + go.mod
│   ├── util.go             # Shared utilities (casing, type mapping)
│   ├── generate_test.go    # Unit tests
│   └── compile_test.go     # Golden-file integration test
├── testdata/
│   ├── input/              # Sample OpenAPI spec
│   │   └── openapi.yaml
│   └── ground-truth/       # Expected generated output
│       ├── mock/
│       │   ├── types.go
│       │   ├── store.go
│       │   ├── handlers.go
│       │   └── routes.go
│       ├── cmd/server/
│       │   └── main.go
│       └── go.mod
├── scripts/
│   ├── build.sh
│   └── build.bat
├── plugin.yaml             # Kalo plugin manifest
├── dist/                   # WASM output
└── go.mod
```

### Generated server structure

```
<output>/
├── mock/
│   ├── types.go            # Go structs from component schemas
│   ├── store.go            # Thread-safe in-memory CRUD store
│   ├── handlers.go         # Gin handler functions
│   └── routes.go           # Route registration
├── cmd/server/
│   └── main.go             # Server entry point
└── go.mod                  # Go module with Gin dependency
```

## Building

```bash
# Native binary
go build ./cmd/plugin

# WASM (for Kalo CLI)
GOOS=wasip1 GOARCH=wasm go build -o dist/plugin.wasm cmd/plugin/main.go
```

## Testing

```bash
go test ./...
```

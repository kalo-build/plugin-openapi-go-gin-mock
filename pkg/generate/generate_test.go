package generate

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testdataDir() string {
	_, filename, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(filename), "..", "..", "testdata")
}

func loadTestSpec(t *testing.T) *Document {
	t.Helper()
	specPath := filepath.Join(testdataDir(), "input", "openapi.yaml")
	specBytes, err := os.ReadFile(specPath)
	require.NoError(t, err)
	doc, err := ParseSpec(specBytes)
	require.NoError(t, err)
	return doc
}

func TestParseSpec(t *testing.T) {
	doc := loadTestSpec(t)

	assert.Equal(t, "Stripe Lite API", doc.Info.Title)
	assert.Equal(t, "1.0.0", doc.Info.Version)

	assert.Len(t, doc.Schemas, 3)
	assert.Equal(t, "Customer", doc.Schemas[0].Name)
	assert.Equal(t, "Invoice", doc.Schemas[1].Name)
	assert.Equal(t, "Payment", doc.Schemas[2].Name)

	assert.True(t, len(doc.Paths) > 0)
	assert.True(t, len(doc.Resources) > 0)
}

func TestParseSpec_Resources(t *testing.T) {
	doc := loadTestSpec(t)

	require.Len(t, doc.Resources, 3)

	cust := doc.Resources[0]
	assert.Equal(t, "Customer", cust.Name)
	assert.Equal(t, "/v1/customers", cust.BasePath)
	assert.Equal(t, "id", cust.IDParam)
	assert.Len(t, cust.Operations, 4)

	inv := doc.Resources[1]
	assert.Equal(t, "Invoice", inv.Name)
	assert.Len(t, inv.Operations, 3)
	assert.Len(t, inv.Actions, 2)

	pay := doc.Resources[2]
	assert.Equal(t, "Payment", pay.Name)
	assert.Len(t, pay.Operations, 3)
}

func TestGenerateTypes(t *testing.T) {
	doc := loadTestSpec(t)

	types, err := GenerateTypes(doc, "mock")
	require.NoError(t, err)

	assert.Contains(t, types, "package mock")
	assert.Contains(t, types, "type Customer struct")
	assert.Contains(t, types, "type Invoice struct")
	assert.Contains(t, types, "type Payment struct")
	assert.Contains(t, types, "time.Time")
	assert.Contains(t, types, `json:"id"`)
	assert.Contains(t, types, `json:"email"`)
}

func TestGenerateStore(t *testing.T) {
	doc := loadTestSpec(t)

	cfg := Config{PackageName: "mock", IDPrefix: true}
	store, err := GenerateStore(doc, cfg)
	require.NoError(t, err)

	assert.Contains(t, store, "package mock")
	assert.Contains(t, store, "type Store struct")
	assert.Contains(t, store, "func NewStore()")
	assert.Contains(t, store, "func (s *Store) CreateCustomer")
	assert.Contains(t, store, "func (s *Store) GetCustomer")
	assert.Contains(t, store, "func (s *Store) ListCustomers")
	assert.Contains(t, store, "func (s *Store) CreateInvoice")
	assert.Contains(t, store, "func (s *Store) CreatePayment")
	assert.Contains(t, store, "func (s *Store) Reset()")
	assert.Contains(t, store, `generateID("cus")`)
	assert.Contains(t, store, `generateID("inv")`)
	assert.Contains(t, store, `generateID("pay")`)
	assert.Contains(t, store, `item.Status = "open"`)
	assert.Contains(t, store, `item.Status = "pending"`)
}

func TestGenerateHandlers(t *testing.T) {
	doc := loadTestSpec(t)

	cfg := Config{PackageName: "mock", Health: true, HealthPath: "/health"}
	handlers, err := GenerateHandlers(doc, cfg)
	require.NoError(t, err)

	assert.Contains(t, handlers, "package mock")
	assert.Contains(t, handlers, "func ListCustomersHandler(store *Store) gin.HandlerFunc")
	assert.Contains(t, handlers, "func CreateCustomerHandler(store *Store) gin.HandlerFunc")
	assert.Contains(t, handlers, "func GetCustomerHandler(store *Store) gin.HandlerFunc")
	assert.Contains(t, handlers, "func UpdateCustomerHandler(store *Store) gin.HandlerFunc")
	assert.Contains(t, handlers, "func FinalizeInvoiceHandler(store *Store) gin.HandlerFunc")
	assert.Contains(t, handlers, "func VoidInvoiceHandler(store *Store) gin.HandlerFunc")
	assert.Contains(t, handlers, "func HealthHandler() gin.HandlerFunc")
}

func TestGenerateRoutes(t *testing.T) {
	doc := loadTestSpec(t)

	cfg := Config{PackageName: "mock", Health: true, HealthPath: "/health"}
	routes, err := GenerateRoutes(doc, cfg)
	require.NoError(t, err)

	assert.Contains(t, routes, "package mock")
	assert.Contains(t, routes, "func RegisterRoutes")
	assert.Contains(t, routes, `r.GET("/health"`)
	assert.Contains(t, routes, `r.GET("/v1/customers"`)
	assert.Contains(t, routes, `r.POST("/v1/customers"`)
	assert.Contains(t, routes, `r.GET("/v1/customers/:id"`)
	assert.Contains(t, routes, `r.PUT("/v1/customers/:id"`)
	assert.Contains(t, routes, `r.POST("/v1/invoices/:id/finalize"`)
	assert.Contains(t, routes, `r.POST("/v1/invoices/:id/void"`)
}

func TestGenerateMain(t *testing.T) {
	doc := loadTestSpec(t)

	cfg := Config{PackageName: "mock", ModulePath: "stripe-lite-mock", Port: "8081"}
	main, err := GenerateMain(doc, cfg)
	require.NoError(t, err)

	assert.Contains(t, main, "package main")
	assert.Contains(t, main, `"stripe-lite-mock/mock"`)
	assert.Contains(t, main, "mock.NewStore()")
	assert.Contains(t, main, "mock.RegisterRoutes(r, store)")
}

func TestGenerateGoMod(t *testing.T) {
	cfg := Config{ModulePath: "stripe-lite-mock"}
	gomod := GenerateGoMod(cfg)

	assert.Contains(t, gomod, "module stripe-lite-mock")
	assert.Contains(t, gomod, "github.com/gin-gonic/gin")
}

func TestGenerateMockServer_Full(t *testing.T) {
	inputDir := filepath.Join(testdataDir(), "input")
	outputDir := t.TempDir()

	cfg := Config{
		InputDir:     inputDir,
		OutputDir:    outputDir,
		SpecFileName: "openapi.yaml",
		PackageName:  "mock",
		ModulePath:   "stripe-lite-mock",
		Port:         "8081",
		IDPrefix:     true,
		Health:       true,
		HealthPath:   "/health",
		Dockerfile:   true,
	}

	err := GenerateMockServer(cfg)
	require.NoError(t, err)

	assertFileExists(t, filepath.Join(outputDir, "mock", "types.go"))
	assertFileExists(t, filepath.Join(outputDir, "mock", "store.go"))
	assertFileExists(t, filepath.Join(outputDir, "mock", "handlers.go"))
	assertFileExists(t, filepath.Join(outputDir, "mock", "routes.go"))
	assertFileExists(t, filepath.Join(outputDir, "cmd", "server", "main.go"))
	assertFileExists(t, filepath.Join(outputDir, "go.mod"))
	assertFileExists(t, filepath.Join(outputDir, "Dockerfile"))
}

func TestConfig_Resolve(t *testing.T) {
	cfg := Config{}
	cfg.Resolve()

	assert.Equal(t, "openapi.yaml", cfg.SpecFileName)
	assert.Equal(t, "mock", cfg.PackageName)
	assert.Equal(t, "mock-server", cfg.ModulePath)
	assert.Equal(t, "8081", cfg.Port)
}

func assertFileExists(t *testing.T, path string) {
	t.Helper()
	_, err := os.Stat(path)
	assert.NoError(t, err, "expected file to exist: %s", path)
}

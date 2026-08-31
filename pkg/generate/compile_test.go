package generate_test

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/kalo-build/go-util/assertfile"
	"github.com/kalo-build/plugin-openapi-go-gin-mock/pkg/generate"
	"github.com/stretchr/testify/suite"
)

type CompileTestSuite struct {
	assertfile.FileSuite

	TestDataPath    string
	GroundTruthPath string
}

func TestCompileTestSuite(t *testing.T) {
	suite.Run(t, new(CompileTestSuite))
}

func (suite *CompileTestSuite) SetupTest() {
	_, filename, _, _ := runtime.Caller(0)
	suite.TestDataPath = filepath.Join(filepath.Dir(filename), "..", "..", "testdata")
	suite.GroundTruthPath = filepath.Join(suite.TestDataPath, "ground-truth")
}

func (suite *CompileTestSuite) TestGenerateMockServer() {
	inputDir := filepath.Join(suite.TestDataPath, "input")
	outputDir := suite.T().TempDir()

	cfg := generate.Config{
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

	err := generate.GenerateMockServer(cfg)
	suite.NoError(err)

	typesPath := filepath.Join(outputDir, "mock", "types.go")
	gtTypesPath := filepath.Join(suite.GroundTruthPath, "mock", "types.go")
	suite.FileExists(typesPath)
	suite.FileEquals(typesPath, gtTypesPath)

	storePath := filepath.Join(outputDir, "mock", "store.go")
	gtStorePath := filepath.Join(suite.GroundTruthPath, "mock", "store.go")
	suite.FileExists(storePath)
	suite.FileEquals(storePath, gtStorePath)

	handlersPath := filepath.Join(outputDir, "mock", "handlers.go")
	gtHandlersPath := filepath.Join(suite.GroundTruthPath, "mock", "handlers.go")
	suite.FileExists(handlersPath)
	suite.FileEquals(handlersPath, gtHandlersPath)

	routesPath := filepath.Join(outputDir, "mock", "routes.go")
	gtRoutesPath := filepath.Join(suite.GroundTruthPath, "mock", "routes.go")
	suite.FileExists(routesPath)
	suite.FileEquals(routesPath, gtRoutesPath)

	mainPath := filepath.Join(outputDir, "cmd", "server", "main.go")
	gtMainPath := filepath.Join(suite.GroundTruthPath, "cmd", "server", "main.go")
	suite.FileExists(mainPath)
	suite.FileEquals(mainPath, gtMainPath)

	goModPath := filepath.Join(outputDir, "go.mod")
	gtGoModPath := filepath.Join(suite.GroundTruthPath, "go.mod")
	suite.FileExists(goModPath)
	suite.FileEquals(goModPath, gtGoModPath)

	dockerfilePath := filepath.Join(outputDir, "Dockerfile")
	gtDockerfilePath := filepath.Join(suite.GroundTruthPath, "Dockerfile")
	suite.FileExists(dockerfilePath)
	suite.FileEquals(dockerfilePath, gtDockerfilePath)
}

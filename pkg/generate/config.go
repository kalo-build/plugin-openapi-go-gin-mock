package generate

// Config holds configuration for mock server generation.
type Config struct {
	InputDir     string
	OutputDir    string
	SpecFileName string
	PackageName  string
	ModulePath   string
	Port         string
	IDPrefix     bool
	Health       bool
	HealthPath   string
	Dockerfile   bool
}

// Resolve applies default values for unset fields.
func (c *Config) Resolve() {
	if c.SpecFileName == "" {
		c.SpecFileName = "openapi.yaml"
	}
	if c.PackageName == "" {
		c.PackageName = "mock"
	}
	if c.ModulePath == "" {
		c.ModulePath = "mock-server"
	}
	if c.Port == "" {
		c.Port = "8081"
	}
	if c.HealthPath == "" {
		c.HealthPath = "/health"
	}
}

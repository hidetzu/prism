package fileutil_test

import (
	"testing"

	"github.com/hidetzu/prism/internal/fileutil"
)

func TestDetectLanguage(t *testing.T) {
	tests := []struct {
		filename string
		want     string
	}{
		{"main.go", "Go"},
		{"app.ts", "TypeScript"},
		{"app.tsx", "TypeScript"},
		{"index.js", "JavaScript"},
		{"script.py", "Python"},
		{"Dockerfile", "Dockerfile"},
		{"Makefile", "Makefile"},
		{"config.yaml", "YAML"},
		{"unknown.xyz", ""},
	}
	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			got := fileutil.DetectLanguage(tt.filename)
			if got != tt.want {
				t.Errorf("DetectLanguage(%q) = %q, want %q", tt.filename, got, tt.want)
			}
		})
	}
}

func TestIsTestFile(t *testing.T) {
	tests := []struct {
		path string
		want bool
	}{
		{"handler_test.go", true},
		{"handler.go", false},
		{"app.test.js", true},
		{"app.spec.ts", true},
		{"app.test.tsx", true},
		{"app.spec.jsx", true},
		{"test_utils.py", true},
		{"utils.py", false},
		{"tests/helper.go", true},
		{"__tests__/app.js", true},
		{"testdata/fixture.json", true},
		{"src/handler.go", false},
	}
	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			got := fileutil.IsTestFile(tt.path)
			if got != tt.want {
				t.Errorf("IsTestFile(%q) = %v, want %v", tt.path, got, tt.want)
			}
		})
	}
}

func TestIsConfigFile(t *testing.T) {
	tests := []struct {
		path string
		want bool
	}{
		{".env", true},
		{"config.yaml", true},
		{"tsconfig.json", true},
		{".eslintrc.json", true},
		{".prettierrc", true},
		{"settings.yml", true},
		{"main.go", false},
		{".bashrc", true},
	}
	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			got := fileutil.IsConfigFile(tt.path)
			if got != tt.want {
				t.Errorf("IsConfigFile(%q) = %v, want %v", tt.path, got, tt.want)
			}
		})
	}
}

func TestIsGeneratedFile(t *testing.T) {
	tests := []struct {
		path string
		want bool
	}{
		{"api.gen.go", true},
		{"api.generated.go", true},
		{"schema_generated.ts", true},
		{"api.pb.go", true},
		{"status_string.go", true},
		{"generated/output.go", true},
		{"gen/api.go", true},
		{"main.go", false},
	}
	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			got := fileutil.IsGeneratedFile(tt.path)
			if got != tt.want {
				t.Errorf("IsGeneratedFile(%q) = %v, want %v", tt.path, got, tt.want)
			}
		})
	}
}

func TestIsDocFile(t *testing.T) {
	tests := []struct {
		path string
		want bool
	}{
		{"README.md", true},
		{"docs/spec.rst", true},
		{"CHANGELOG", true},
		{"LICENSE", true},
		{"CONTRIBUTING", true},
		{"notes.txt", true},
		{"main.go", false},
	}
	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			got := fileutil.IsDocFile(tt.path)
			if got != tt.want {
				t.Errorf("IsDocFile(%q) = %v, want %v", tt.path, got, tt.want)
			}
		})
	}
}

func TestIsInfraFile(t *testing.T) {
	tests := []struct {
		path string
		want bool
	}{
		{".github/workflows/ci.yml", true},
		{".circleci/config.yml", true},
		{"terraform/main.tf", true},
		{"Dockerfile", true},
		{"docker-compose.yml", true},
		{"Jenkinsfile", true},
		{"k8s/deployment.yaml", true},
		{"main.go", false},
	}
	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			got := fileutil.IsInfraFile(tt.path)
			if got != tt.want {
				t.Errorf("IsInfraFile(%q) = %v, want %v", tt.path, got, tt.want)
			}
		})
	}
}

func TestIsDependencyFile(t *testing.T) {
	tests := []struct {
		path string
		want bool
	}{
		{"go.mod", true},
		{"go.sum", true},
		{"package.json", true},
		{"yarn.lock", true},
		{"Cargo.toml", true},
		{"requirements.txt", true},
		{"main.go", false},
	}
	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			got := fileutil.IsDependencyFile(tt.path)
			if got != tt.want {
				t.Errorf("IsDependencyFile(%q) = %v, want %v", tt.path, got, tt.want)
			}
		})
	}
}

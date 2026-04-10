package fileutil

import (
	"path/filepath"
	"strings"
)

// DetectLanguage returns the programming language based on file extension.
func DetectLanguage(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	if lang, ok := languageByExt[ext]; ok {
		return lang
	}

	base := filepath.Base(filename)
	switch base {
	case "Dockerfile":
		return "Dockerfile"
	case "Makefile":
		return "Makefile"
	}
	return ""
}

// IsTestFile returns true if the file is a test file based on name and path conventions.
func IsTestFile(path string) bool {
	base := filepath.Base(path)
	dir := filepath.Dir(path)

	// Go
	if strings.HasSuffix(base, "_test.go") {
		return true
	}
	// JS/TS
	if strings.HasSuffix(base, ".test.js") || strings.HasSuffix(base, ".test.ts") ||
		strings.HasSuffix(base, ".test.tsx") || strings.HasSuffix(base, ".test.jsx") ||
		strings.HasSuffix(base, ".spec.js") || strings.HasSuffix(base, ".spec.ts") ||
		strings.HasSuffix(base, ".spec.tsx") || strings.HasSuffix(base, ".spec.jsx") {
		return true
	}
	// Python
	if strings.HasPrefix(base, "test_") && strings.HasSuffix(base, ".py") {
		return true
	}
	// Directory-based
	parts := strings.Split(dir, "/")
	for _, p := range parts {
		if p == "test" || p == "tests" || p == "__tests__" || p == "testdata" {
			return true
		}
	}
	return false
}

// IsConfigFile returns true if the file is a configuration file.
func IsConfigFile(path string) bool {
	base := filepath.Base(path)
	if knownConfigFiles[base] {
		return true
	}
	if strings.HasPrefix(base, ".") && (strings.HasSuffix(base, "rc") || strings.HasSuffix(base, "rc.json") || strings.HasSuffix(base, "rc.yml")) {
		return true
	}
	// Extension-based (for classifier: .yaml, .yml, etc.)
	ext := strings.ToLower(filepath.Ext(path))
	return ext == ".yaml" || ext == ".yml" || ext == ".toml" || ext == ".ini" || ext == ".cfg"
}

// IsGeneratedFile returns true if the file appears to be auto-generated.
func IsGeneratedFile(path string) bool {
	base := filepath.Base(path)

	if strings.Contains(base, ".gen.") || strings.Contains(base, ".generated.") ||
		strings.Contains(base, "_generated") || strings.HasSuffix(base, ".pb.go") ||
		strings.HasSuffix(base, "_string.go") {
		return true
	}

	dir := filepath.Dir(path)
	parts := strings.Split(dir, "/")
	for _, p := range parts {
		if p == "generated" || p == "gen" {
			return true
		}
	}
	return false
}

// IsDocFile returns true if the file is a documentation file.
func IsDocFile(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	if ext == ".md" || ext == ".rst" || ext == ".txt" || ext == ".adoc" {
		return true
	}
	lower := strings.ToLower(filepath.Base(path))
	return lower == "license" || lower == "changelog" || lower == "contributing"
}

// IsInfraFile returns true if the file is an infrastructure or CI file.
func IsInfraFile(path string) bool {
	lower := strings.ToLower(path)
	for _, dir := range infraDirs {
		if strings.HasPrefix(lower, dir) || strings.Contains(lower, "/"+dir) {
			return true
		}
	}
	base := strings.ToLower(filepath.Base(path))
	return base == "dockerfile" || base == "docker-compose.yml" || base == "docker-compose.yaml" ||
		strings.HasSuffix(base, ".tf") || base == "jenkinsfile" || base == "cloudbuild.yaml"
}

// IsDependencyFile returns true if the file is a dependency/lock file.
func IsDependencyFile(path string) bool {
	return knownDepFiles[filepath.Base(path)]
}

var languageByExt = map[string]string{
	".go":    "Go",
	".js":    "JavaScript",
	".ts":    "TypeScript",
	".tsx":   "TypeScript",
	".jsx":   "JavaScript",
	".py":    "Python",
	".rb":    "Ruby",
	".java":  "Java",
	".kt":    "Kotlin",
	".rs":    "Rust",
	".c":     "C",
	".cpp":   "C++",
	".h":     "C",
	".hpp":   "C++",
	".cs":    "C#",
	".swift": "Swift",
	".php":   "PHP",
	".sh":    "Shell",
	".bash":  "Shell",
	".yaml":  "YAML",
	".yml":   "YAML",
	".json":  "JSON",
	".toml":  "TOML",
	".xml":   "XML",
	".html":  "HTML",
	".css":   "CSS",
	".scss":  "SCSS",
	".sql":   "SQL",
	".md":    "Markdown",
	".tf":    "Terraform",
}

var knownConfigFiles = map[string]bool{
	".env": true, ".env.example": true,
	"config.yaml": true, "config.yml": true, "config.json": true, "config.toml": true,
	".eslintrc.json": true, ".eslintrc.js": true, ".eslintrc.yml": true,
	".prettierrc": true, ".prettierrc.json": true,
	"tsconfig.json": true, "jest.config.js": true, "jest.config.ts": true,
	"webpack.config.js": true, "vite.config.ts": true, "vite.config.js": true,
	".goreleaser.yml": true, ".golangci.yml": true,
	"pyproject.toml": true, "setup.cfg": true,
}

var infraDirs = []string{
	".github/", ".circleci/", ".gitlab-ci", "terraform/", "infra/", "deploy/", "k8s/", "helm/",
}

var knownDepFiles = map[string]bool{
	"go.mod": true, "go.sum": true,
	"package.json": true, "package-lock.json": true, "yarn.lock": true, "pnpm-lock.yaml": true,
	"requirements.txt": true, "Pipfile": true, "Pipfile.lock": true, "poetry.lock": true, "pyproject.toml": true,
	"Gemfile": true, "Gemfile.lock": true,
	"Cargo.toml": true, "Cargo.lock": true,
	"build.gradle": true, "build.gradle.kts": true, "pom.xml": true,
}

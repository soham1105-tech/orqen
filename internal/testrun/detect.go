package testrun

import (
	"path/filepath"

	"github.com/soham1105-tech/orqen/internal/domain"
	"github.com/soham1105-tech/orqen/internal/scanner/golang"
	"github.com/soham1105-tech/orqen/internal/scanner/node"
	"github.com/soham1105-tech/orqen/internal/scanner/python"
	"github.com/soham1105-tech/orqen/internal/scanner/rust"
)

// Detect resolves the test command for a service from its source directory on
// disk. The scanner persists Service.TestCmd at detection time; this fallback
// covers projects initialized before test detection whose lock still has an
// empty TestCmd. It returns "" for unsupported languages and never errors.
func Detect(svc *domain.Service, repoRoot string) string {
	if svc == nil {
		return ""
	}
	dir := repoRoot
	if svc.Path != "" && svc.Path != "." {
		dir = filepath.Join(repoRoot, filepath.FromSlash(svc.Path))
	}
	switch svc.Language {
	case domain.LangNode:
		pkg, err := node.ParsePackageJSON(dir)
		if err != nil {
			return ""
		}
		pm, _ := node.DetectPackageManager(dir, pkg)
		return node.TestCommand(dir, pkg, pm)
	case domain.LangPython:
		pyproj, _ := python.ParsePyProject(dir)
		deps := python.CollectDependencies(dir, pyproj)
		return python.TestCommand(dir, deps)
	case domain.LangGo:
		if golang.IsGoDir(dir) {
			return "go test ./..."
		}
		return ""
	case domain.LangRust:
		if rust.IsRustDir(dir) {
			return "cargo test"
		}
		return ""
	default:
		return ""
	}
}

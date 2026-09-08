package openapi

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/getkin/kin-openapi/openapi3"
)

var (
	loadOnce sync.Once
	doc      *openapi3.T
	loadErr  error
)

// Load reads and caches api/openapi.yaml from the first existing candidate path.
func Load() (*openapi3.T, error) {
	loadOnce.Do(func() {
		for _, p := range candidatePaths() {
			data, err := os.ReadFile(p) //nolint:gosec // path from fixed allowlist
			if err != nil {
				continue
			}
			loader := openapi3.NewLoader()
			loader.IsExternalRefsAllowed = true
			doc, loadErr = loader.LoadFromData(data)
			if loadErr == nil {
				doc.Servers = openapi3.Servers{{URL: "/"}}
				return
			}
		}
		if loadErr == nil {
			loadErr = fmt.Errorf("openapi spec not found")
		}
	})
	return doc, loadErr
}

// ProjectRoot returns the repository root (directory containing go.mod and api/openapi.yaml).
func ProjectRoot() string {
	if root := findProjectRoot(); root != "" {
		return root
	}
	wd, _ := os.Getwd()
	return wd
}

func findProjectRoot() string {
	wd, err := os.Getwd()
	if err != nil {
		return ""
	}
	dir := wd
	for i := 0; i < 8; i++ {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			if _, err := os.Stat(filepath.Join(dir, "api", "openapi.yaml")); err == nil {
				return dir
			}
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return ""
}

func candidatePaths() []string {
	root := findProjectRoot()
	var paths []string
	if root != "" {
		paths = append(paths, filepath.Join(root, "api", "openapi.yaml"))
	}
	if wd, err := os.Getwd(); err == nil {
		paths = append(paths,
			filepath.Join(wd, "api", "openapi.yaml"),
			filepath.Join(wd, "openapi.yaml"),
		)
	}
	paths = append(paths, "api/openapi.yaml", "./api/openapi.yaml", "openapi.yaml", "./openapi.yaml")
	return paths
}

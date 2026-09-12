package cognitivecore

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDockerBuildsStageNestedModuleBeforeDownload(t *testing.T) {
	repositoryRoot := filepath.Clean(filepath.Join("..", ".."))
	for _, name := range []string{"Dockerfile", "Dockerfile.autonomous"} {
		content, err := os.ReadFile(filepath.Join(repositoryRoot, name))
		if err != nil {
			t.Fatalf("read %s: %v", name, err)
		}
		text := string(content)
		copyIndex := strings.Index(text, "COPY cognitive-core/ecco9/go.mod")
		downloadIndex := strings.Index(text, "go mod download")
		if copyIndex < 0 || downloadIndex < 0 || copyIndex > downloadIndex {
			t.Fatalf("%s must stage cognitive-core/ecco9/go.mod before go mod download", name)
		}
	}
}

func TestImporterSourceIdentityCannotBeOverridden(t *testing.T) {
	repositoryRoot := filepath.Clean(filepath.Join("..", ".."))
	content, err := os.ReadFile(filepath.Join(repositoryRoot, "scripts", "sync_ecco9_cognitive_core.py"))
	if err != nil {
		t.Fatalf("read importer: %v", err)
	}
	text := string(content)
	for _, forbidden := range []string{"--expected-repository", "--expected-commit"} {
		if strings.Contains(text, forbidden) {
			t.Fatalf("importer exposes forbidden source identity override %q", forbidden)
		}
	}
	for _, required := range []string{
		`DEFAULT_SOURCE_REPOSITORY = "https://github.com/o9nn/ecco9.git"`,
		`DEFAULT_SOURCE_COMMIT = "1b22401ee8842fd1aa2769b688d588cd9f1f9ce9"`,
		`commit != DEFAULT_SOURCE_COMMIT`,
		`canonical_repository(repository) != canonical_repository(DEFAULT_SOURCE_REPOSITORY)`,
	} {
		if !strings.Contains(text, required) {
			t.Fatalf("importer is missing pinned source guard %q", required)
		}
	}
}

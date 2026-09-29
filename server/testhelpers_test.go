package server

import (
	"cmp"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cogpy/echo9llama/envconfig"
	"github.com/cogpy/echo9llama/fs/ggml"
)

// createBinFile writes a GGUF file and links it into the models blob store,
// returning its path and digest.
func createBinFile(t *testing.T, kv map[string]any, ti []*ggml.Tensor) (string, string) {
	t.Helper()
	t.Setenv("OLLAMA_MODELS", cmp.Or(os.Getenv("OLLAMA_MODELS"), t.TempDir()))

	modelDir := envconfig.Models()

	f, err := os.CreateTemp(t.TempDir(), "")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	if err := ggml.WriteGGUF(f, kv, ti); err != nil {
		t.Fatal(err)
	}
	if _, err := f.Seek(0, 0); err != nil {
		t.Fatal(err)
	}

	digest, _ := GetSHA256Digest(f)
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}

	blob := filepath.Join(modelDir, "blobs", fmt.Sprintf("sha256-%s", strings.TrimPrefix(digest, "sha256:")))
	if err := os.MkdirAll(filepath.Dir(blob), 0o755); err != nil {
		t.Fatal(err)
	}
	_ = os.Remove(blob)
	if err := os.Symlink(f.Name(), blob); err != nil {
		b, err := os.ReadFile(f.Name())
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(blob, b, 0o644); err != nil {
			t.Fatal(err)
		}
	}

	return f.Name(), digest
}

// wordTokenize counts whitespace-separated words as tokens.
func wordTokenize(_ context.Context, s string) (tokens []int, err error) {
	for range strings.Fields(s) {
		tokens = append(tokens, len(tokens))
	}
	return
}

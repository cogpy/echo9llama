package drivers

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/EchoCog/echollama/core/llm"
)

type stubGenerator struct {
	reply  string
	err    error
	prompt string
	opts   llm.GenerateOptions
}

func (s *stubGenerator) Generate(_ context.Context, prompt string, opts llm.GenerateOptions) (string, error) {
	s.prompt = prompt
	s.opts = opts
	return s.reply, s.err
}

func clearModelEnv(t *testing.T) {
	for _, k := range []string{EnvLLMEndpoint, EnvLLMModel, EnvLLMAPIKey, "OPENAI_API_KEY", "ANTHROPIC_API_KEY", "OPENROUTER_API_KEY"} {
		t.Setenv(k, "")
	}
}

func TestLLMDeviceReturnsOnlyModelOutput(t *testing.T) {
	clearModelEnv(t)
	ctx := context.Background()

	empty := NewLLMDriver(nil)
	if err := empty.Load(nil); err != nil {
		t.Fatal(err)
	}
	dev, _ := empty.GetDevice("llm0")
	if err := dev.Initialize(ctx); !errors.Is(err, ErrNoModelBackend) {
		t.Fatalf("Initialize without model = %v, want ErrNoModelBackend", err)
	}
	if n, err := dev.Write([]byte("hello")); err == nil || n != 0 {
		t.Fatalf("Write without model = %d, %v; want error", n, err)
	}
	buf := make([]byte, 64)
	if n, err := dev.Read(buf); n != 0 || err != io.EOF {
		t.Fatalf("Read without model output = %d, %v; want 0, EOF", n, err)
	}

	gen := &stubGenerator{reply: "model words"}
	d := NewLLMDriverWithGenerator(gen)
	if err := d.Load(nil); err != nil {
		t.Fatal(err)
	}
	dev, _ = d.GetDevice("llm0")
	if err := dev.Initialize(ctx); err != nil {
		t.Fatal(err)
	}
	if _, err := dev.Write([]byte("hello")); err != nil {
		t.Fatal(err)
	}
	n, err := dev.Read(buf)
	if err != nil || string(buf[:n]) != "model words" || gen.prompt != "hello" {
		t.Fatalf("Read = %q, %v (prompt %q)", buf[:n], err, gen.prompt)
	}

	gen.err = errors.New("down")
	if _, err := dev.Write([]byte("again")); err == nil {
		t.Fatal("expected model failure to surface")
	}
	if n, err := dev.Read(buf); n != 0 || err != io.EOF {
		t.Fatalf("Read after failure = %d, %v; want 0, EOF", n, err)
	}
}

func TestNPUInferenceUsesModelBackend(t *testing.T) {
	clearModelEnv(t)
	npu := NewNPUDevice("npu0", nil)
	if err := npu.Initialize(context.Background()); err != nil {
		t.Fatal(err)
	}
	if out, err := npu.Infer("prompt", DefaultNPUSequenceConfig()); !errors.Is(err, ErrNoModelBackend) || out != "" {
		t.Fatalf("Infer without backend = %q, %v", out, err)
	}
	if !npu.HasError() {
		t.Fatal("expected error status bit without backend")
	}

	gen := &stubGenerator{reply: "npu completion"}
	npu.SetBackend(gen)
	seq := DefaultNPUSequenceConfig()
	seq.SystemPrompt = "sys"
	out, err := npu.Infer("prompt", seq)
	if err != nil || out != "npu completion" {
		t.Fatalf("Infer = %q, %v", out, err)
	}
	if gen.opts.MaxTokens != int(seq.NPredict) || gen.opts.SystemPrompt != "sys" {
		t.Fatalf("options not forwarded: %+v", gen.opts)
	}
	if _, err := npu.Write([]byte("w")); err != nil || npu.LastCompletion() != "npu completion" {
		t.Fatalf("Write = %v, last %q", err, npu.LastCompletion())
	}

	gen.reply = "   "
	if _, err := npu.Infer("prompt", seq); err == nil {
		t.Fatal("empty model output must be an error")
	}
}

func TestOpenAICompatibleGenerator(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/chat/completions" || r.Header.Get("Authorization") != "Bearer k" {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		var req chatRequest
		_ = json.NewDecoder(r.Body).Decode(&req)
		last := req.Messages[len(req.Messages)-1].Content
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"choices": []map[string]interface{}{{"message": map[string]string{"role": "assistant", "content": "echo:" + last}}},
		})
	}))
	defer srv.Close()

	t.Setenv(EnvLLMEndpoint, srv.URL+"/v1")
	t.Setenv(EnvLLMModel, "m")
	t.Setenv(EnvLLMAPIKey, "k")
	gen, err := ModelBackendFromEnv()
	if err != nil {
		t.Fatal(err)
	}
	out, err := gen.Generate(context.Background(), "hi", llm.DefaultGenerateOptions())
	if err != nil || out != "echo:hi" {
		t.Fatalf("Generate = %q, %v", out, err)
	}

	bad := NewOpenAICompatibleGenerator(srv.URL+"/v1", "m", "wrong")
	if _, err := bad.Generate(context.Background(), "hi", llm.DefaultGenerateOptions()); err == nil || !strings.Contains(err.Error(), "400") {
		t.Fatalf("expected HTTP error, got %v", err)
	}
}

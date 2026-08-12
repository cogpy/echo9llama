package backendcap

import (
	"encoding/binary"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestProbeHostMemoryReturnsExplicitKnowledgeState(t *testing.T) {
	probe := ProbeHostMemory()
	if probe.Tier == "" {
		t.Fatal("expected explicit host memory tier")
	}
	if probe.Known && (probe.TotalBytes == 0 || probe.AvailableBytes == 0) {
		t.Fatalf("known probe must contain usable values: %+v", probe)
	}
	if !probe.Known && probe.Tier != MemoryUnknown {
		t.Fatalf("unknown probe must report unknown tier: %+v", probe)
	}
}

func TestReadGGUFModelMetadataReturnsScrubbedStableIdentity(t *testing.T) {
	directory := t.TempDir()
	path := filepath.Join(directory, "tiny.gguf")
	writeTinyGGUF(t, path, "tiny echo", 2048)
	metadata, err := ReadGGUFModelMetadata(path)
	if err != nil {
		t.Fatal(err)
	}
	if metadata.Name != "tiny echo" || metadata.Architecture != "llama" || metadata.ContextLength != 2048 {
		t.Fatalf("unexpected metadata: %+v", metadata)
	}
	if metadata.Quantization != "Q4_K_M" || metadata.EstimatedMemoryBytes == 0 {
		t.Fatalf("unexpected quantization or estimate: %+v", metadata)
	}
	if metadata.ModelID == "" || strings.Contains(metadata.ModelID, directory) || strings.Contains(metadata.ModelID, string(os.PathSeparator)) {
		t.Fatalf("model ID must be stable and path-scrubbed: %q", metadata.ModelID)
	}
	second, err := ReadGGUFModelMetadata(path)
	if err != nil || second.ModelID != metadata.ModelID {
		t.Fatalf("model ID must be deterministic: first=%q second=%q err=%v", metadata.ModelID, second.ModelID, err)
	}
}

func TestReadGGUFModelMetadataRejectsInvalidAndTruncatedFiles(t *testing.T) {
	directory := t.TempDir()
	wrongMagic := filepath.Join(directory, "wrong.gguf")
	if err := os.WriteFile(wrongMagic, []byte("NOPE-not-a-model-file-padding-123456"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := ReadGGUFModelMetadata(wrongMagic); err == nil || !strings.Contains(err.Error(), "not a GGUF") {
		t.Fatalf("expected magic rejection, got %v", err)
	}

	truncated := filepath.Join(directory, "truncated.gguf")
	if err := os.WriteFile(truncated, append([]byte("GGUF"), 3, 0, 0, 0), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := ReadGGUFModelMetadata(truncated); err == nil {
		t.Fatal("expected truncated header rejection")
	}
}

func TestReadGGUFModelMetadataRejectsExcessiveMetadata(t *testing.T) {
	path := filepath.Join(t.TempDir(), "excessive.gguf")
	file, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	mustWrite(t, file, []byte("GGUF"))
	mustBinaryWrite(t, file, uint32(3))
	mustBinaryWrite(t, file, uint64(0))
	mustBinaryWrite(t, file, uint64(maxGGUFMetadataKeys+1))
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := ReadGGUFModelMetadata(path); err == nil || !strings.Contains(err.Error(), "key count") {
		t.Fatalf("expected metadata-count rejection, got %v", err)
	}
}

func TestProbeModelFileHonorsMemoryKnowledgeAndCGOGating(t *testing.T) {
	directory := t.TempDir()
	path := filepath.Join(directory, "tiny.gguf")
	writeTinyGGUF(t, path, "tiny-echo", 2048)

	unknown := HostMemoryProbe{Known: false, Tier: MemoryUnknown, Reason: "test unknown"}
	capability, err := ProbeModelFile(path, DiscoveryOptions{HostMemory: &unknown, Roots: []string{directory}})
	if err != nil {
		t.Fatal(err)
	}
	if capability.MemorySafe || capability.Available {
		t.Fatalf("unknown memory must reject native load by default: %+v", capability)
	}

	capability, err = ProbeModelFile(path, DiscoveryOptions{HostMemory: &unknown, Roots: []string{directory}, AllowUnknownMemory: true})
	if err != nil {
		t.Fatal(err)
	}
	if !capability.MemorySafe {
		t.Fatalf("explicit unknown-memory override should satisfy memory policy: %+v", capability)
	}
	if capability.Available != NativeRuntimeEnabled() {
		t.Fatalf("availability must match cgo build state: available=%v native=%v", capability.Available, NativeRuntimeEnabled())
	}

	insufficient := HostMemoryProbe{Known: true, TotalBytes: 2 * gib, AvailableBytes: 512 * 1024 * 1024, Tier: MemoryConstrained, Reason: "test low memory"}
	capability, err = ProbeModelFile(path, DiscoveryOptions{HostMemory: &insufficient, Roots: []string{directory}, MemoryReserveBytes: 128 * 1024 * 1024})
	if err != nil {
		t.Fatal(err)
	}
	if capability.MemorySafe || !strings.Contains(capability.Reason, "exceeds safe") {
		t.Fatalf("expected insufficient-memory rejection: %+v", capability)
	}
}

func TestDiscoverModelCapabilitiesEnforcesRootsAndLimit(t *testing.T) {
	root := t.TempDir()
	outside := t.TempDir()
	writeTinyGGUF(t, filepath.Join(root, "one.gguf"), "one", 1024)
	writeTinyGGUF(t, filepath.Join(root, "two.gguf"), "two", 2048)
	outsideModel := filepath.Join(outside, "outside.gguf")
	writeTinyGGUF(t, outsideModel, "outside", 2048)
	link := filepath.Join(root, "escape.gguf")
	if err := os.Symlink(outsideModel, link); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}

	host := HostMemoryProbe{Known: true, TotalBytes: 64 * gib, AvailableBytes: 48 * gib, Tier: MemoryStress, Reason: "test host"}
	capabilities := DiscoverModelCapabilities([]string{root}, DiscoveryOptions{Roots: []string{root}, MaxFiles: 1, HostMemory: &host})
	if len(capabilities) != 1 {
		t.Fatalf("expected max-files limit of one and no symlink escape, got %+v", capabilities)
	}
	if capabilities[0].ModelID == "" || capabilities[0].ModelPath == outsideModel {
		t.Fatalf("unexpected discovered capability: %+v", capabilities[0])
	}
	if _, err := ProbeModelFile(outsideModel, DiscoveryOptions{Roots: []string{root}, HostMemory: &host}); err == nil {
		t.Fatal("expected explicit outside-root model rejection")
	}
}

func TestReadCgroupValue(t *testing.T) {
	path := filepath.Join(t.TempDir(), "memory.max")
	if err := os.WriteFile(path, []byte("1073741824\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	value, ok := readCgroupValue(path)
	if !ok || value != gib {
		t.Fatalf("unexpected cgroup value %d ok=%v", value, ok)
	}
	if err := os.WriteFile(path, []byte("max\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, ok := readCgroupValue(path); ok {
		t.Fatal("unbounded cgroup value must not be treated as a numeric limit")
	}
}

func writeTinyGGUF(t *testing.T, path, name string, contextLength uint32) {
	t.Helper()
	file, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	mustWrite(t, file, []byte("GGUF"))
	mustBinaryWrite(t, file, uint32(3))
	mustBinaryWrite(t, file, uint64(0))
	mustBinaryWrite(t, file, uint64(4))
	writeKVString(t, file, "general.name", name)
	writeKVString(t, file, "general.architecture", "llama")
	writeKVUint32(t, file, "llama.context_length", contextLength)
	writeKVUint32(t, file, "general.file_type", 15)
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}
}

func writeKVString(t *testing.T, file *os.File, key, value string) {
	t.Helper()
	writeString(t, file, key)
	mustBinaryWrite(t, file, uint32(8))
	writeString(t, file, value)
}

func writeKVUint32(t *testing.T, file *os.File, key string, value uint32) {
	t.Helper()
	writeString(t, file, key)
	mustBinaryWrite(t, file, uint32(4))
	mustBinaryWrite(t, file, value)
}

func writeString(t *testing.T, file *os.File, value string) {
	t.Helper()
	mustBinaryWrite(t, file, uint64(len(value)))
	mustWrite(t, file, []byte(value))
}

func mustBinaryWrite(t *testing.T, writer *os.File, value interface{}) {
	t.Helper()
	if err := binary.Write(writer, binary.LittleEndian, value); err != nil {
		t.Fatal(err)
	}
}

func mustWrite(t *testing.T, writer *os.File, value []byte) {
	t.Helper()
	if _, err := writer.Write(value); err != nil {
		t.Fatal(err)
	}
}

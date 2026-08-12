package backendcap

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
)

const (
	gib                       = uint64(1024 * 1024 * 1024)
	defaultMaxModelProbeFiles = 128
	maxGGUFMetadataKeys       = 4096
	maxGGUFMetadataBytes      = int64(64 * 1024 * 1024)
	maxGGUFStringBytes        = uint64(16 * 1024 * 1024)
	maxGGUFArrayElements      = uint64(1_000_000)
	defaultMemorySafetyRatio  = 0.80
	defaultMemoryReserveBytes = uint64(1024 * 1024 * 1024)
)

// HostMemoryProbe describes the effective memory envelope available to local inference.
type HostMemoryProbe struct {
	Known          bool       `json:"known"`
	TotalBytes     uint64     `json:"total_bytes"`
	AvailableBytes uint64     `json:"available_bytes"`
	LimitBytes     uint64     `json:"limit_bytes,omitempty"`
	UsageBytes     uint64     `json:"usage_bytes,omitempty"`
	CgroupLimited  bool       `json:"cgroup_limited"`
	Tier           MemoryTier `json:"tier"`
	Reason         string     `json:"reason"`
}

// DiscoveryOptions controls model path and memory-safety policy.
type DiscoveryOptions struct {
	Roots              []string
	AllowUnknownMemory bool
	MemorySafetyRatio  float64
	MemoryReserveBytes uint64
	MaxFiles           int
	HostMemory         *HostMemoryProbe
}

// ModelFileMetadata captures scheduler-relevant metadata discovered from a GGUF file.
type ModelFileMetadata struct {
	Path                 string                 `json:"-"`
	ModelID              string                 `json:"model_id"`
	Name                 string                 `json:"name"`
	Architecture         string                 `json:"architecture,omitempty"`
	FileSizeBytes        uint64                 `json:"file_size_bytes"`
	ContextLength        int                    `json:"context_length,omitempty"`
	Quantization         string                 `json:"quantization,omitempty"`
	EstimatedMemoryBytes uint64                 `json:"estimated_memory_bytes"`
	Metadata             map[string]interface{} `json:"metadata,omitempty"`
}

// ProbeHostMemory inspects host and container memory without importing native backends.
func ProbeHostMemory() HostMemoryProbe {
	if runtime.GOOS != "linux" {
		return HostMemoryProbe{Known: false, Tier: MemoryUnknown, Reason: "host memory probe is not implemented for " + runtime.GOOS}
	}

	fields, err := readMemInfo("/proc/meminfo")
	if err != nil {
		return HostMemoryProbe{Known: false, Tier: MemoryUnknown, Reason: "host memory probe unavailable: " + err.Error()}
	}

	total := fields["MemTotal"] * 1024
	available := fields["MemAvailable"] * 1024
	if available == 0 {
		available = fields["MemFree"] * 1024
	}
	if total == 0 || available == 0 {
		return HostMemoryProbe{Known: false, TotalBytes: total, AvailableBytes: available, Tier: MemoryUnknown, Reason: "host memory probe returned incomplete values"}
	}

	probe := HostMemoryProbe{Known: true, TotalBytes: total, AvailableBytes: available}
	if limit, usage, ok := probeCgroupMemory(); ok && limit > 0 && limit < total {
		probe.CgroupLimited = true
		probe.LimitBytes = limit
		probe.UsageBytes = usage
		probe.TotalBytes = limit
		headroom := uint64(0)
		if usage < limit {
			headroom = limit - usage
		}
		if headroom < probe.AvailableBytes {
			probe.AvailableBytes = headroom
		}
	}
	probe.Tier = memoryTierFromBytes(probe.TotalBytes)
	probe.Reason = fmt.Sprintf("effective host envelope reports %.1f GiB total and %.1f GiB available", bytesToGiB(probe.TotalBytes), bytesToGiB(probe.AvailableBytes))
	if probe.CgroupLimited {
		probe.Reason += " after cgroup limits"
	}
	return probe
}

// HostMemoryTier returns the effective host memory tier.
func HostMemoryTier() MemoryTier {
	return ProbeHostMemory().Tier
}

func probeCgroupMemory() (uint64, uint64, bool) {
	if limit, ok := readCgroupValue("/sys/fs/cgroup/memory.max"); ok {
		usage, usageOK := readCgroupValue("/sys/fs/cgroup/memory.current")
		return limit, usage, usageOK
	}
	limit, limitOK := readCgroupValue("/sys/fs/cgroup/memory/memory.limit_in_bytes")
	usage, usageOK := readCgroupValue("/sys/fs/cgroup/memory/memory.usage_in_bytes")
	return limit, usage, limitOK && usageOK
}

func readCgroupValue(path string) (uint64, bool) {
	data, err := os.ReadFile(path)
	if err != nil {
		return 0, false
	}
	value := strings.TrimSpace(string(data))
	if value == "" || value == "max" {
		return 0, false
	}
	parsed, err := strconv.ParseUint(value, 10, 64)
	return parsed, err == nil
}

func readMemInfo(path string) (map[string]uint64, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	result := make(map[string]uint64)
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || !strings.Contains(line, ":") {
			continue
		}
		parts := strings.SplitN(line, ":", 2)
		fields := strings.Fields(parts[1])
		if len(fields) == 0 {
			continue
		}
		value, err := strconv.ParseUint(fields[0], 10, 64)
		if err == nil {
			result[strings.TrimSpace(parts[0])] = value
		}
	}
	return result, nil
}

func normalizedDiscoveryOptions(options []DiscoveryOptions) DiscoveryOptions {
	var option DiscoveryOptions
	if len(options) > 0 {
		option = options[0]
	}
	if option.MemorySafetyRatio <= 0 || option.MemorySafetyRatio > 1 {
		option.MemorySafetyRatio = defaultMemorySafetyRatio
	}
	if option.MemoryReserveBytes == 0 {
		option.MemoryReserveBytes = defaultMemoryReserveBytes
	}
	if option.MaxFiles <= 0 || option.MaxFiles > defaultMaxModelProbeFiles {
		option.MaxFiles = defaultMaxModelProbeFiles
	}
	return option
}

// ProbeModelFile returns a concrete local inference capability for a GGUF file.
func ProbeModelFile(path string, options ...DiscoveryOptions) (Capability, error) {
	option := normalizedDiscoveryOptions(options)
	canonical, roots, err := canonicalModelPath(path, option.Roots)
	if err != nil {
		return Capability{}, err
	}
	if !withinAnyRoot(canonical, roots) {
		return Capability{}, fmt.Errorf("GGUF path is outside configured model roots")
	}
	metadata, err := ReadGGUFModelMetadata(canonical)
	if err != nil {
		return Capability{}, err
	}

	host := ProbeHostMemory()
	if option.HostMemory != nil {
		host = *option.HostMemory
	}
	memorySafe, memoryReason := modelFitsHost(metadata.EstimatedMemoryBytes, host, option)
	available := cgoEnabled && memorySafe
	reason := fmt.Sprintf("GGUF model %s requires approximately %.1f GiB; %s", metadata.Name, bytesToGiB(metadata.EstimatedMemoryBytes), memoryReason)
	if !cgoEnabled {
		reason = "native GGUF runtime unavailable: cgo is disabled; " + reason
	}

	return Capability{
		Name:                 "model:" + metadata.Name,
		ProviderName:         "local_gguf",
		ModelID:              metadata.ModelID,
		Kind:                 BackendNativeCPU,
		Available:            available,
		Native:               true,
		Offline:              true,
		Concrete:             true,
		MemorySafe:           memorySafe,
		StressGrade:          modelMemoryTier(metadata.EstimatedMemoryBytes) == MemoryStress,
		MemoryTier:           modelMemoryTier(metadata.EstimatedMemoryBytes),
		ModelPath:            metadata.Path,
		ContextLength:        metadata.ContextLength,
		Quantization:         metadata.Quantization,
		EstimatedMemoryBytes: metadata.EstimatedMemoryBytes,
		MaxConcurrency:       1,
		Priority:             20,
		Reason:               reason,
		Guidance:             "Use when this concrete GGUF model satisfies context and effective host-memory constraints.",
	}, nil
}

func modelFitsHost(estimated uint64, host HostMemoryProbe, option DiscoveryOptions) (bool, string) {
	if !host.Known || host.AvailableBytes == 0 {
		if option.AllowUnknownMemory {
			return true, "effective available memory is unknown and explicit operator override is enabled"
		}
		return false, "effective available memory is unknown and override is disabled"
	}
	ratioLimit := uint64(float64(host.AvailableBytes) * option.MemorySafetyRatio)
	reserveLimit := uint64(0)
	if host.AvailableBytes > option.MemoryReserveBytes {
		reserveLimit = host.AvailableBytes - option.MemoryReserveBytes
	}
	limit := ratioLimit
	if reserveLimit < limit {
		limit = reserveLimit
	}
	if estimated > limit {
		return false, fmt.Sprintf("estimated footprint exceeds safe %.1f GiB load limit within %s", bytesToGiB(limit), host.Reason)
	}
	return true, fmt.Sprintf("estimated footprint fits safe %.1f GiB load limit within %s", bytesToGiB(limit), host.Reason)
}

// DiscoverModelCapabilities probes explicit GGUF files/directories under allowed roots.
func DiscoverModelCapabilities(paths []string, options ...DiscoveryOptions) []Capability {
	option := normalizedDiscoveryOptions(options)
	files := discoverGGUFFiles(paths, option)
	caps := make([]Capability, 0, len(files))
	for _, path := range files {
		capability, err := ProbeModelFile(path, option)
		if err == nil {
			caps = append(caps, capability)
		}
	}
	sort.SliceStable(caps, func(i, j int) bool {
		if caps[i].Name == caps[j].Name {
			return caps[i].ModelID < caps[j].ModelID
		}
		return caps[i].Name < caps[j].Name
	})
	return caps
}

func discoverGGUFFiles(paths []string, option DiscoveryOptions) []string {
	roots := canonicalRoots(option.Roots)
	if len(roots) == 0 {
		roots = rootsFromPaths(paths)
	}
	seen := make(map[string]struct{})
	files := make([]string, 0)

	addCandidate := func(candidate string) {
		if len(files) >= option.MaxFiles {
			return
		}
		canonical, err := filepath.EvalSymlinks(candidate)
		if err != nil {
			return
		}
		canonical, err = filepath.Abs(canonical)
		if err != nil || !withinAnyRoot(canonical, roots) || !strings.EqualFold(filepath.Ext(canonical), ".gguf") {
			return
		}
		info, err := os.Stat(canonical)
		if err != nil || !info.Mode().IsRegular() {
			return
		}
		if _, exists := seen[canonical]; exists {
			return
		}
		seen[canonical] = struct{}{}
		files = append(files, canonical)
	}

	for _, raw := range paths {
		if len(files) >= option.MaxFiles {
			break
		}
		path := strings.TrimSpace(raw)
		if path == "" {
			continue
		}
		canonical, err := filepath.EvalSymlinks(path)
		if err != nil {
			continue
		}
		canonical, err = filepath.Abs(canonical)
		if err != nil || !withinAnyRoot(canonical, roots) {
			continue
		}
		info, err := os.Stat(canonical)
		if err != nil {
			continue
		}
		if !info.IsDir() {
			addCandidate(canonical)
			continue
		}
		if err := filepath.WalkDir(canonical, func(candidate string, entry os.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if len(files) >= option.MaxFiles {
				return filepath.SkipAll
			}
			if entry.IsDir() {
				if candidate != canonical && strings.HasPrefix(entry.Name(), ".") {
					return filepath.SkipDir
				}
				return nil
			}
			addCandidate(candidate)
			return nil
		}); err != nil {
			continue
		}
	}
	sort.Strings(files)
	return files
}

func canonicalModelPath(path string, configuredRoots []string) (string, []string, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return "", nil, fmt.Errorf("empty GGUF model path")
	}
	canonical, err := filepath.EvalSymlinks(path)
	if err != nil {
		return "", nil, fmt.Errorf("resolve GGUF model path: %w", err)
	}
	canonical, err = filepath.Abs(canonical)
	if err != nil {
		return "", nil, fmt.Errorf("canonicalize GGUF model path: %w", err)
	}
	roots := canonicalRoots(configuredRoots)
	if len(roots) == 0 {
		roots = []string{filepath.Dir(canonical)}
	}
	return canonical, roots, nil
}

func canonicalRoots(roots []string) []string {
	seen := make(map[string]struct{})
	result := make([]string, 0, len(roots))
	for _, raw := range roots {
		root := strings.TrimSpace(raw)
		if root == "" {
			continue
		}
		canonical, err := filepath.EvalSymlinks(root)
		if err != nil {
			continue
		}
		canonical, err = filepath.Abs(canonical)
		if err != nil {
			continue
		}
		if _, exists := seen[canonical]; !exists {
			seen[canonical] = struct{}{}
			result = append(result, canonical)
		}
	}
	sort.Strings(result)
	return result
}

func rootsFromPaths(paths []string) []string {
	roots := make([]string, 0, len(paths))
	for _, raw := range paths {
		path := strings.TrimSpace(raw)
		if path == "" {
			continue
		}
		canonical, err := filepath.EvalSymlinks(path)
		if err != nil {
			continue
		}
		canonical, err = filepath.Abs(canonical)
		if err != nil {
			continue
		}
		info, err := os.Stat(canonical)
		if err != nil {
			continue
		}
		if info.IsDir() {
			roots = append(roots, canonical)
		} else {
			roots = append(roots, filepath.Dir(canonical))
		}
	}
	return canonicalRoots(roots)
}

func withinAnyRoot(path string, roots []string) bool {
	for _, root := range roots {
		relative, err := filepath.Rel(root, path)
		if err != nil {
			continue
		}
		if relative == "." || (relative != ".." && !strings.HasPrefix(relative, ".."+string(os.PathSeparator)) && !filepath.IsAbs(relative)) {
			return true
		}
	}
	return false
}

// ReadGGUFModelMetadata reads only the bounded metadata needed for routing.
func ReadGGUFModelMetadata(path string) (ModelFileMetadata, error) {
	file, err := os.Open(path)
	if err != nil {
		return ModelFileMetadata{}, err
	}
	defer file.Close()
	stat, err := file.Stat()
	if err != nil {
		return ModelFileMetadata{}, err
	}
	if !stat.Mode().IsRegular() || stat.Size() < 24 {
		return ModelFileMetadata{}, fmt.Errorf("%s is not a valid regular GGUF file", filepath.Base(path))
	}

	reader := &io.LimitedReader{R: file, N: minInt64(stat.Size(), maxGGUFMetadataBytes)}
	var magic [4]byte
	if _, err := io.ReadFull(reader, magic[:]); err != nil {
		return ModelFileMetadata{}, err
	}
	if string(magic[:]) != "GGUF" {
		return ModelFileMetadata{}, fmt.Errorf("%s is not a GGUF file", filepath.Base(path))
	}
	var version uint32
	var tensorCount, kvCount uint64
	if err := binary.Read(reader, binary.LittleEndian, &version); err != nil {
		return ModelFileMetadata{}, err
	}
	if version < 2 || version > 3 {
		return ModelFileMetadata{}, fmt.Errorf("unsupported GGUF version %d", version)
	}
	if err := binary.Read(reader, binary.LittleEndian, &tensorCount); err != nil {
		return ModelFileMetadata{}, err
	}
	if err := binary.Read(reader, binary.LittleEndian, &kvCount); err != nil {
		return ModelFileMetadata{}, err
	}
	if kvCount > maxGGUFMetadataKeys {
		return ModelFileMetadata{}, fmt.Errorf("GGUF metadata key count %d exceeds safety limit", kvCount)
	}

	metadata := make(map[string]interface{}, int(kvCount)+3)
	for index := range kvCount {
		key, err := readGGUFString(reader)
		if err != nil {
			return ModelFileMetadata{}, fmt.Errorf("read GGUF key %d: %w", index, err)
		}
		var valueType uint32
		if err := binary.Read(reader, binary.LittleEndian, &valueType); err != nil {
			return ModelFileMetadata{}, err
		}
		value, err := readGGUFValue(reader, valueType)
		if err != nil {
			return ModelFileMetadata{}, fmt.Errorf("read GGUF metadata %s: %w", key, err)
		}
		metadata[key] = value
	}

	name := stringFromMetadata(metadata, "general.name")
	if name == "" {
		name = strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	}
	architecture := stringFromMetadata(metadata, "general.architecture")
	contextLength := intFromMetadata(metadata, architecture+".context_length", "llama.context_length", "context_length")
	quantization := quantizationFromMetadata(metadata)
	fileSize := uint64(stat.Size())
	estimate := estimateModelMemory(fileSize, contextLength)
	metadata["gguf.version"] = uint64(version)
	metadata["gguf.tensor_count"] = tensorCount
	metadata["gguf.metadata_kv_count"] = kvCount

	canonical, _ := filepath.Abs(path)
	return ModelFileMetadata{
		Path:                 canonical,
		ModelID:              scrubbedModelID(name, canonical),
		Name:                 name,
		Architecture:         architecture,
		FileSizeBytes:        fileSize,
		ContextLength:        contextLength,
		Quantization:         quantization,
		EstimatedMemoryBytes: estimate,
		Metadata:             metadata,
	}, nil
}

func readGGUFString(reader io.Reader) (string, error) {
	var length uint64
	if err := binary.Read(reader, binary.LittleEndian, &length); err != nil {
		return "", err
	}
	if length > maxGGUFStringBytes {
		return "", fmt.Errorf("GGUF string length %d exceeds safety limit", length)
	}
	buffer := make([]byte, int(length))
	_, err := io.ReadFull(reader, buffer)
	return string(buffer), err
}

func readGGUFValue(reader io.Reader, valueType uint32) (interface{}, error) {
	switch valueType {
	case 0:
		var value uint8
		return value, binary.Read(reader, binary.LittleEndian, &value)
	case 1:
		var value int8
		return value, binary.Read(reader, binary.LittleEndian, &value)
	case 2:
		var value uint16
		return value, binary.Read(reader, binary.LittleEndian, &value)
	case 3:
		var value int16
		return value, binary.Read(reader, binary.LittleEndian, &value)
	case 4:
		var value uint32
		return value, binary.Read(reader, binary.LittleEndian, &value)
	case 5:
		var value int32
		return value, binary.Read(reader, binary.LittleEndian, &value)
	case 6:
		var value float32
		return value, binary.Read(reader, binary.LittleEndian, &value)
	case 7:
		var raw uint8
		if err := binary.Read(reader, binary.LittleEndian, &raw); err != nil {
			return false, err
		}
		return raw != 0, nil
	case 8:
		return readGGUFString(reader)
	case 9:
		var elementType uint32
		var length uint64
		if err := binary.Read(reader, binary.LittleEndian, &elementType); err != nil {
			return nil, err
		}
		if err := binary.Read(reader, binary.LittleEndian, &length); err != nil {
			return nil, err
		}
		if length > maxGGUFArrayElements {
			return nil, fmt.Errorf("GGUF array length %d exceeds safety limit", length)
		}
		values := make([]interface{}, 0, int(math.Min(float64(length), 32)))
		for index := range length {
			value, err := readGGUFValue(reader, elementType)
			if err != nil {
				return nil, err
			}
			if index < 32 {
				values = append(values, value)
			}
		}
		return values, nil
	case 10:
		var value uint64
		return value, binary.Read(reader, binary.LittleEndian, &value)
	case 11:
		var value int64
		return value, binary.Read(reader, binary.LittleEndian, &value)
	case 12:
		var value float64
		return value, binary.Read(reader, binary.LittleEndian, &value)
	default:
		return nil, fmt.Errorf("unsupported GGUF metadata type %d", valueType)
	}
}

func scrubbedModelID(name, path string) string {
	hash := sha256.Sum256([]byte(path))
	safeName := strings.Trim(strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
			return r
		}
		return '-'
	}, name), "-")
	if safeName == "" {
		safeName = "gguf"
	}
	return safeName + "-" + hex.EncodeToString(hash[:6])
}

func memoryTierFromBytes(bytes uint64) MemoryTier {
	switch {
	case bytes >= 32*gib:
		return MemoryStress
	case bytes >= 8*gib:
		return MemoryStandard
	default:
		return MemoryConstrained
	}
}

func modelMemoryTier(bytes uint64) MemoryTier {
	switch {
	case bytes >= 16*gib:
		return MemoryStress
	case bytes >= 4*gib:
		return MemoryStandard
	default:
		return MemoryConstrained
	}
}

func bytesToGiB(bytes uint64) float64 {
	return float64(bytes) / float64(gib)
}

func estimateModelMemory(fileSize uint64, contextLength int) uint64 {
	estimate := uint64(float64(fileSize) * 1.20)
	if contextLength > 0 {
		estimate += uint64(contextLength) * 512 * 1024
	}
	if estimate < fileSize {
		return fileSize
	}
	return estimate
}

func stringFromMetadata(metadata map[string]interface{}, keys ...string) string {
	for _, key := range keys {
		if value, ok := metadata[key].(string); ok {
			return value
		}
	}
	return ""
}

func intFromMetadata(metadata map[string]interface{}, keys ...string) int {
	for _, key := range keys {
		if key == ".context_length" {
			continue
		}
		value, exists := metadata[key]
		if !exists {
			continue
		}
		switch typed := value.(type) {
		case uint8:
			return int(typed)
		case int8:
			return int(typed)
		case uint16:
			return int(typed)
		case int16:
			return int(typed)
		case uint32:
			return int(typed)
		case int32:
			return int(typed)
		case uint64:
			if typed <= uint64(^uint(0)>>1) {
				return int(typed)
			}
		case int64:
			if typed >= 0 && uint64(typed) <= uint64(^uint(0)>>1) {
				return int(typed)
			}
		}
	}
	return 0
}

func quantizationFromMetadata(metadata map[string]interface{}) string {
	fileType := intFromMetadata(metadata, "general.file_type")
	if label, ok := ggufFileTypes[fileType]; ok {
		return label
	}
	if fileType != 0 {
		return fmt.Sprintf("file_type_%d", fileType)
	}
	return "unknown"
}

func minInt64(left, right int64) int64 {
	if left < right {
		return left
	}
	return right
}

var ggufFileTypes = map[int]string{
	0: "F32", 1: "F16", 2: "Q4_0", 3: "Q4_1", 6: "Q5_0", 7: "Q5_1",
	8: "Q8_0", 9: "Q8_1", 10: "Q2_K", 11: "Q3_K_S", 12: "Q3_K_M",
	13: "Q3_K_L", 14: "Q4_K_S", 15: "Q4_K_M", 16: "Q5_K_S", 17: "Q5_K_M",
	18: "Q6_K", 19: "IQ2_XXS", 20: "IQ2_XS", 21: "Q2_K_S", 22: "IQ3_XS",
	23: "IQ3_XXS", 24: "IQ1_S", 25: "IQ4_NL", 26: "IQ3_S", 27: "IQ3_M",
	28: "IQ2_S", 29: "IQ2_M", 30: "IQ4_XS", 31: "IQ1_M",
}

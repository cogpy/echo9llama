package cognitivecore

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

type sourceSnapshotManifest struct {
	lineageManifest
	SourceRepository string `json:"source_repository"`
	SnapshotRole     string `json:"snapshot_role"`
	ActivationPolicy string `json:"activation_policy"`
	GoPathsSHA256    string `json:"go_paths_sha256"`
	FilesSHA256      string `json:"files_sha256"`
}

func digestBytes(content []byte) string {
	digest := sha256.Sum256(content)
	return hex.EncodeToString(digest[:])
}

func TestCurrentSourceSnapshotVerifiesCompletePathAndHashInventory(t *testing.T) {
	root := filepath.Join("..", "..", "cognitive-core", "ecco9", "upstream")
	manifestPath := filepath.Join(root, "SOURCE_SNAPSHOT_MANIFEST.json")
	content, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatal(err)
	}
	if digestBytes(content) != SourceSnapshotManifestSHA256 {
		t.Fatal("raw source manifest differs from independent adapter pin")
	}
	var manifest sourceSnapshotManifest
	if err := json.Unmarshal(content, &manifest); err != nil {
		t.Fatal(err)
	}
	if manifest.SchemaVersion != 1 || manifest.SourceCommit != SourceCommitSHA || manifest.SourceRepository != SourceRepositoryURL {
		t.Fatal("raw source identity differs from current source pin")
	}
	if manifest.SourceCommit == RetainedRuntimeSourceCommitSHA || manifest.SnapshotRole != "immutable_raw_source_not_runtime" || manifest.ActivationPolicy != "no_auto_activation" {
		t.Fatal("raw source must not be presented as retained or automatically activated runtime")
	}
	if manifest.LocalAdditionCount != 0 || manifest.LocalPatchCount != 0 {
		t.Fatal("raw source must contain no local overlays")
	}
	paths := make(map[string]bool)
	var goPaths, fileDigests []string
	counts := make(map[string]int)
	for _, record := range manifest.Files {
		if !filepath.IsLocal(record.Path) || filepath.ToSlash(filepath.Clean(record.Path)) != record.Path || paths[record.Path] {
			t.Fatalf("unsafe or duplicate source path: %q", record.Path)
		}
		paths[record.Path] = true
		content, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(record.Path)))
		if err != nil {
			t.Fatalf("read raw source %s: %v", record.Path, err)
		}
		if int64(len(content)) != record.Bytes || digestBytes(content) != record.SHA256 {
			t.Fatalf("raw source size/hash mismatch: %s", record.Path)
		}
		if record.LocalPatch != "" || record.UpstreamSHA256 != "" {
			t.Fatalf("raw source record cannot carry an overlay: %s", record.Path)
		}
		counts[record.Kind]++
		switch record.Kind {
		case "go":
			if !strings.HasSuffix(record.Path, ".go") {
				t.Fatalf("Go record lacks Go suffix: %s", record.Path)
			}
			goPaths = append(goPaths, record.Path)
		case "embed_asset", "license", "upstream_module", "module_boundary":
		default:
			t.Fatalf("unknown source record kind: %s", record.Kind)
		}
		fileDigests = append(fileDigests, record.Path+"\x00"+record.SHA256+"\n")
	}
	if counts["go"] != PreservedGoFileCount || counts["go"] != manifest.GoFileCount || counts["embed_asset"] != 45 || counts["embed_asset"] != manifest.EmbedAssetCount || counts["license"] != 3 || counts["license"] != manifest.LicenseFileCount || counts["module_boundary"] != 1 || counts["upstream_module"] != 2 {
		t.Fatalf("unexpected exact raw counts: %v", counts)
	}
	sort.Strings(goPaths)
	sort.Strings(fileDigests)
	goPathDigest := digestBytes([]byte(strings.Join(goPaths, "\n") + "\n"))
	if goPathDigest != manifest.GoPathsSHA256 || goPathDigest != SourceSnapshotGoPathsSHA256 {
		t.Fatal("complete tracked Go path set differs from independent path-set pin")
	}
	filesDigest := digestBytes([]byte(strings.Join(fileDigests, "")))
	if filesDigest != manifest.FilesSHA256 || filesDigest != SourceSnapshotFilesSHA256 {
		t.Fatal("raw path/hash inventory differs from independent content pin")
	}
	paths["SOURCE_SNAPSHOT_MANIFEST.json"] = true
	visited := make(map[string]bool)
	err = filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.Type()&os.ModeSymlink != 0 {
			t.Fatalf("raw source must not contain symlinks: %s", path)
		}
		if entry.IsDir() {
			return nil
		}
		relative, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		relative = filepath.ToSlash(relative)
		if !paths[relative] {
			t.Fatalf("unmanifested raw source file: %s", relative)
		}
		visited[relative] = true
		return nil
	})
	if err != nil || len(visited) != len(paths) {
		t.Fatalf("raw source path completeness failed: files=%d records=%d err=%v", len(visited), len(paths), err)
	}
	boundary, err := os.ReadFile(filepath.Join(root, "go.mod"))
	if err != nil {
		t.Fatal(err)
	}
	if string(boundary) != "module github.com/EchoCog/echollama\n\ngo 1.26.0\n\ntoolchain go1.26.11\n" {
		t.Fatal("raw source must remain a nested minimal compatibility module")
	}
	originalModule, err := os.ReadFile(filepath.Join(root, "go.mod.upstream"))
	if err != nil || len(originalModule) == 0 || string(originalModule) == string(boundary) {
		t.Fatal("different original upstream module graph must be captured separately")
	}
}

func TestCurrentSourceNativeAllowlistMatchesRetainedRuntimeExactly(t *testing.T) {
	root := filepath.Join("..", "..", "cognitive-core", "ecco9")
	var runtime lineageManifest
	content, err := os.ReadFile(filepath.Join(root, "LINEAGE_MANIFEST.json"))
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(content, &runtime); err != nil {
		t.Fatal(err)
	}
	if runtime.SourceCommit != RetainedRuntimeSourceCommitSHA || runtime.SourceCommit == SourceCommitSHA || len(runtime.Files) != 615 {
		t.Fatal("runtime overlay provenance and 615-record inventory must remain independent")
	}
	runtimeGo := 0
	runtimeRecords := make(map[string]string)
	for _, record := range runtime.Files {
		runtimeRecords[record.Path] = record.SHA256
		if record.Kind == "go" || record.Kind == "local_go" {
			runtimeGo++
		}
	}
	if runtimeGo != 567 || runtime.LocalPatchCount != 118 || runtime.LocalAdditionCount != 14 {
		t.Fatalf("retained runtime counts unexpectedly changed: Go=%d patches=%d additions=%d", runtimeGo, runtime.LocalPatchCount, runtime.LocalAdditionCount)
	}
	if len(activeSourcePaths) != 6 {
		t.Fatal("raw preservation must not expand the active native allowlist")
	}
	for _, path := range activeSourcePaths {
		working, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(path)))
		if err != nil {
			t.Fatal(err)
		}
		raw, err := os.ReadFile(filepath.Join(root, "upstream", filepath.FromSlash(path)))
		if err != nil {
			t.Fatal(err)
		}
		if string(working) != string(raw) || digestBytes(raw) != runtimeRecords[path] {
			t.Fatalf("native file was changed or falsely relabeled current: %s", path)
		}
	}
	provenance := DefaultProvenance()
	if provenance.SourceCommit != SourceCommitSHA || provenance.ManifestSHA256 != ManifestSHA256 || provenance.CatalogueSHA256 != CatalogueSHA256 {
		t.Fatal("native provenance must retain executed runtime manifest/catalogue hashes while advancing only its byte-identical subset's source commit")
	}
	if provenance.AdapterVersion != AdapterVersion || provenance.TrustGrade != TrustGrade || len(provenance.SourcePaths) != 6 {
		t.Fatal("snapshot refresh must not alter native adapter semantics or authority")
	}
}

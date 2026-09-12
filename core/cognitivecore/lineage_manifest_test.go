package cognitivecore

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

type lineageManifest struct {
	SchemaVersion    int    `json:"schema_version"`
	SourceCommit     string `json:"source_commit"`
	GoFileCount      int    `json:"go_file_count"`
	EmbedAssetCount  int    `json:"embed_asset_count"`
	LicenseFileCount int    `json:"license_file_count"`
	Files            []struct {
		Path   string `json:"path"`
		Kind   string `json:"kind"`
		Bytes  int64  `json:"bytes"`
		SHA256 string `json:"sha256"`
	} `json:"files"`
}

type integrationCatalogue struct {
	Files []struct {
		Path        string `json:"path"`
		Disposition string `json:"disposition"`
	} `json:"files"`
}

func TestImportedLineageManifestVerifiesEveryGoFile(t *testing.T) {
	root := filepath.Clean(filepath.Join("..", "..", "cognitive-core", "ecco9"))
	manifestBytes, err := os.ReadFile(filepath.Join(root, "LINEAGE_MANIFEST.json"))
	if err != nil {
		t.Fatalf("read lineage manifest: %v", err)
	}
	manifestDigest := sha256.Sum256(manifestBytes)
	if encoded := hex.EncodeToString(manifestDigest[:]); encoded != ManifestSHA256 {
		t.Fatalf("manifest SHA-256 = %s, adapter pins %s", encoded, ManifestSHA256)
	}
	var manifest lineageManifest
	if err := json.Unmarshal(manifestBytes, &manifest); err != nil {
		t.Fatalf("decode lineage manifest: %v", err)
	}
	if manifest.SchemaVersion != 1 || manifest.SourceCommit != SourceCommitSHA {
		t.Fatalf("unexpected source identity: version=%d commit=%s", manifest.SchemaVersion, manifest.SourceCommit)
	}
	if manifest.GoFileCount != PreservedGoFileCount {
		t.Fatalf("go_file_count = %d, want %d", manifest.GoFileCount, PreservedGoFileCount)
	}

	verifiedGo := 0
	verifiedAssets := 0
	verifiedLicenses := 0
	for _, record := range manifest.Files {
		content, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(record.Path)))
		if err != nil {
			t.Fatalf("read preserved %s: %v", record.Path, err)
		}
		if int64(len(content)) != record.Bytes {
			t.Fatalf("size mismatch for %s: got %d want %d", record.Path, len(content), record.Bytes)
		}
		digest := sha256.Sum256(content)
		if encoded := hex.EncodeToString(digest[:]); encoded != record.SHA256 {
			t.Fatalf("SHA-256 mismatch for %s", record.Path)
		}
		switch record.Kind {
		case "go":
			verifiedGo++
		case "embed_asset":
			verifiedAssets++
		case "license":
			verifiedLicenses++
		default:
			t.Fatalf("unknown lineage record kind %q", record.Kind)
		}
	}
	if verifiedGo != manifest.GoFileCount || verifiedAssets != manifest.EmbedAssetCount || verifiedLicenses != manifest.LicenseFileCount {
		t.Fatalf("verified counts go=%d assets=%d licenses=%d, manifest go=%d assets=%d licenses=%d", verifiedGo, verifiedAssets, verifiedLicenses, manifest.GoFileCount, manifest.EmbedAssetCount, manifest.LicenseFileCount)
	}

	catalogueBytes, err := os.ReadFile(filepath.Join(root, "INTEGRATION_CATALOGUE.json"))
	if err != nil {
		t.Fatalf("read integration catalogue: %v", err)
	}
	catalogueDigest := sha256.Sum256(catalogueBytes)
	if hex.EncodeToString(catalogueDigest[:]) != CatalogueSHA256 {
		t.Fatalf("integration catalogue digest does not match active adapter binding")
	}
	var catalogue integrationCatalogue
	if err := json.Unmarshal(catalogueBytes, &catalogue); err != nil {
		t.Fatalf("decode integration catalogue: %v", err)
	}
	if len(catalogue.Files) != manifest.GoFileCount {
		t.Fatalf("catalogue files = %d, want %d", len(catalogue.Files), manifest.GoFileCount)
	}
	active := make(map[string]bool)
	for _, file := range catalogue.Files {
		if file.Disposition == "active_adapter_source" {
			active[file.Path] = true
		}
	}
	if len(active) != len(activeSourcePaths) {
		t.Fatalf("active source files = %d, adapter binds %d", len(active), len(activeSourcePaths))
	}
	for _, path := range activeSourcePaths {
		if !active[path] {
			t.Fatalf("adapter source %q is not active in catalogue", path)
		}
	}
}

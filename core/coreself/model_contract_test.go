package coreself

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"
)

const (
	cog253CorpusDigest  = "6dbf445e98e3306e407f25e74d238a3df6393115037e94c0816b65974a51b0d6"
	cog253LicenseDigest = "8486a10c4393cee1c25392769ddd3b2d6c242d6ec7928e1414efff7dfb2f07ef"
)

func artifactPath(parts ...string) string {
	return filepath.Join(append([]string{"..", "..", "core-self"}, parts...)...)
}

func readArtifactJSON(t *testing.T, target string, value any) {
	t.Helper()
	data, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(data, value); err != nil {
		t.Fatalf("decode %s: %v", target, err)
	}
}

func fileSHA256(t *testing.T, target string) string {
	t.Helper()
	data, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func TestCog253SemanticSourceContract(t *testing.T) {
	if got := fileSHA256(t, artifactPath("sources", "cog253", "archetypal_patterns.json")); got != cog253CorpusDigest {
		t.Fatalf("Cog253 corpus digest = %s", got)
	}
	if got := fileSHA256(t, artifactPath("sources", "cog253", "LICENSE")); got != cog253LicenseDigest {
		t.Fatalf("Cog253 license digest = %s", got)
	}
	var mapping struct {
		SourceCorpus map[string]any `json:"source_corpus"`
		Patterns     []struct {
			ID                   string `json:"id"`
			Name                 string `json:"name"`
			SourcePatternID      string `json:"source_pattern_id"`
			SourceName           string `json:"source_name"`
			SourceConceptualText string `json:"source_conceptual_content"`
			AcceptanceCriterion  string `json:"acceptance_criterion"`
		} `json:"patterns"`
	}
	readArtifactJSON(t, artifactPath("data", "cog253-map.json"), &mapping)
	if len(mapping.Patterns) != 253 || mapping.SourceCorpus["content_sha256"] != cog253CorpusDigest || mapping.SourceCorpus["license"] != "MIT" {
		t.Fatalf("invalid Cog253 source contract: patterns=%d source=%v", len(mapping.Patterns), mapping.SourceCorpus)
	}
	for index, pattern := range mapping.Patterns {
		wantID := fmt.Sprintf("cog%03d", index+1)
		if pattern.ID != wantID || pattern.SourcePatternID == "" || pattern.SourceName == "" || pattern.SourceConceptualText == "" || !strings.Contains(pattern.Name, pattern.SourceName) || !strings.Contains(pattern.AcceptanceCriterion, pattern.SourcePatternID) {
			t.Fatalf("non-semantic pattern %d: %+v", index+1, pattern)
		}
	}
}

func TestKSMDefinitionsAndImplementationEvidenceContract(t *testing.T) {
	var definitions struct {
		Definitions []struct {
			ID        int    `json:"id"`
			Name      string `json:"name"`
			Invariant string `json:"invariant"`
		} `json:"definitions"`
	}
	readArtifactJSON(t, artifactPath("data", "core-self-61.json"), &definitions)
	if len(definitions.Definitions) != 61 {
		t.Fatalf("definition count = %d", len(definitions.Definitions))
	}
	for index, definition := range definitions.Definitions {
		if definition.ID != index+1 || definition.Name == "" || definition.Invariant == "" {
			t.Fatalf("invalid definition %d: %+v", index+1, definition)
		}
	}
	var evidence map[string]struct {
		Status   string   `json:"status"`
		Evidence []string `json:"evidence"`
	}
	readArtifactJSON(t, artifactPath("data", "implementation-evidence.json"), &evidence)
	want := []string{"cog001", "cog003", "cog005", "cog008", "cog014", "cog015", "cog017", "cog020", "cog024", "cog028", "cog031", "cog032", "cog037", "cog043", "cog113", "cog122", "cog208", "cog212", "cog249"}
	got := make([]string, 0, len(evidence))
	for patternID, record := range evidence {
		got = append(got, patternID)
		if record.Status != "verified" || len(record.Evidence) < 2 {
			t.Fatalf("weak implementation evidence %s: %+v", patternID, record)
		}
	}
	sort.Strings(got)
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("evidence IDs = %v, want %v", got, want)
	}
}

func TestExternalManifestsRemainNonCanonical(t *testing.T) {
	entries, err := os.ReadDir(artifactPath("manifests"))
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 5 {
		t.Fatalf("manifest count = %d", len(entries))
	}
	for _, entry := range entries {
		var manifest map[string]any
		readArtifactJSON(t, artifactPath("manifests", entry.Name()), &manifest)
		if manifest["canonical_subject"] != RootSubjectID || manifest["canonical_authority"] != "none" {
			t.Fatalf("manifest acquired canonical authority: %s %+v", entry.Name(), manifest)
		}
		if entry.Name() != "ecco9_cognitive_core.json" && (manifest["execution_permitted"] != false || manifest["runtime_registered"] != false) {
			t.Fatalf("proposed manifest became executable: %s %+v", entry.Name(), manifest)
		}
	}
	var neon map[string]any
	readArtifactJSON(t, artifactPath("manifests", "neon_angel_mesh.json"), &neon)
	if neon["content_digest"] != "656ac8d426370996adc3d8a4aa68b85167bdd8cb84b1d88713963038f9fe146b" || neon["trust_grade"] != "unverified_license" {
		t.Fatalf("Neon Angel evidence contract changed: %+v", neon)
	}
}

func TestPolicyOverlayAndBootstrapCapsuleContract(t *testing.T) {
	var policy map[string]any
	readArtifactJSON(t, artifactPath("policy", "external-authority.json"), &policy)
	if policy["default"] != "deny" || policy["proposal_is_identity"] != false {
		t.Fatalf("external authority policy weakened: %+v", policy)
	}
	var persona map[string]any
	readArtifactJSON(t, artifactPath("persona", "superhotgirl-overlay.json"), &persona)
	if persona["canonical_authority"] != "none" || persona["runtime_registered"] != false || persona["dynamic"] != true {
		t.Fatalf("persona overlay crossed authority boundary: %+v", persona)
	}
	var capsule Capsule
	readArtifactJSON(t, artifactPath("ledger", "bootstrap-capsule.json"), &capsule)
	if err := VerifyCapsule(capsule); err != nil {
		t.Fatalf("bootstrap capsule: %v", err)
	}
	if len(capsule.Ledger) != 1 || capsule.Snapshot.State.EventCount != 1 {
		t.Fatalf("bootstrap capsule contains non-genesis authority: %+v", capsule.Snapshot.State)
	}
}

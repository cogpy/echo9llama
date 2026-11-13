package cognitivecore

const (
	SourceRepositoryURL = "https://github.com/o9nn/ecco9.git"
	// SourceCommitSHA binds only the six reviewed, byte-identical native files.
	SourceCommitSHA = "7d32d715c633a92c684900fbaf25b0111c708612"
	// Retained runtime overlays and their catalogue keep their original identity.
	RetainedRuntimeSourceCommitSHA = "1b22401ee8842fd1aa2769b688d588cd9f1f9ce9"
	ManifestSHA256                 = "6781879f5ddf50598694f06730b1028d07c8ee93a1b179d2e69281d469c40660"
	CatalogueSHA256                = "bd864497ddb035716666fc021c4d575f27d4123f3a59de10ea254d7a60db9f88"
	SourceSnapshotManifestSHA256   = "1ec1b76965056f79b3d21bbd5ad737a36b44add8f6a731c0fad71c570ea1a385"
	SourceSnapshotFilesSHA256      = "212ca92e96d0bf651597d8788a399cd5ed8fd41b0d0f6649230c6e7f5a3f6539"
	SourceSnapshotGoPathsSHA256    = "ec09cc18eed639934503c7f9d4b6947181b13cf04f96b655f757432322f6678e"
	// PreservedGoFileCount counts current raw upstream Go files, not the retained
	// runtime's 615 records (553 imported Go + 14 local Go + 48 assets/licenses).
	PreservedGoFileCount = 553
)

var activeSourcePaths = []string{
	"core/ecco9/platform.go",
	"core/ecco9/types.go",
	"core/ecco9/drivers/reservoir_driver.go",
	"core/ecco9/drivers/memory_driver.go",
	"core/ecco9/drivers/emotion_driver.go",
	"core/ecco9/drivers/consciousness_driver.go",
}

// DefaultProvenance advances the source commit only for the reviewed active
// subset, proven byte-identical to current raw source. The manifest/catalogue
// pins still bind the executed retained runtime, including its original overlay
// provenance. Their six active records have exactly the raw snapshot's hashes;
// this does not relabel any other old runtime overlay as current upstream.
// SourceSnapshotManifestSHA256 independently pins the complete raw snapshot.
// Callers receive their own source-path slice.
func DefaultProvenance() Provenance {
	return Provenance{
		SourceRepository: SourceRepositoryURL,
		SourceCommit:     SourceCommitSHA,
		ManifestSHA256:   ManifestSHA256,
		CatalogueSHA256:  CatalogueSHA256,
		SourcePaths:      append([]string(nil), activeSourcePaths...),
		AdapterVersion:   AdapterVersion,
		TrustGrade:       TrustGrade,
	}
}

// NewDefaultBridge constructs the unchanged reviewed native core. Dormant raw
// source stays behind its own module boundary and receives no runtime authority.
func NewDefaultBridge() (*Bridge, error) {
	return NewBridge(DefaultProvenance(), PreservedGoFileCount)
}

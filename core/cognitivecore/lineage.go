package cognitivecore

const (
	SourceRepositoryURL  = "https://github.com/o9nn/ecco9.git"
	SourceCommitSHA      = "1b22401ee8842fd1aa2769b688d588cd9f1f9ce9"
	ManifestSHA256       = "d1ce16f6a68f138e93c7c2a9a95d29ece1011f432574bedf94187bd50af06e14"
	CatalogueSHA256      = "53a7c8349889af8e7c27aefb1d9efa28fff991ef2aa5299c3a0135da72b053dc"
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

// DefaultProvenance returns the immutable source binding for the reviewed
// active subset. Callers receive their own source-path slice.
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

// NewDefaultBridge constructs the reviewed active core while the rest of the
// imported Go lineage remains preserved behind the nested-module boundary.
func NewDefaultBridge() (*Bridge, error) {
	return NewBridge(DefaultProvenance(), PreservedGoFileCount)
}

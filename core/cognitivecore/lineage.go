package cognitivecore

const (
	SourceRepositoryURL  = "https://github.com/o9nn/ecco9.git"
	SourceCommitSHA      = "1b22401ee8842fd1aa2769b688d588cd9f1f9ce9"
	ManifestSHA256       = "275a0460d4c13f94f4909f9f8ac65d95d48b7511280ca0131bbe08b9bf2cc498"
	CatalogueSHA256      = "c50f920be992ed73057ded062ce783e4c196eb5794831a5b9d5433ea127e9826"
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

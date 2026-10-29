package cognitivecore

const (
	SourceRepositoryURL  = "https://github.com/o9nn/ecco9.git"
	SourceCommitSHA      = "1b22401ee8842fd1aa2769b688d588cd9f1f9ce9"
	ManifestSHA256       = "4457602637e32fa7c1058a7e40d7ecdd7b08e398443ba8f50f6cba4ea415b6ee"
	CatalogueSHA256      = "b388efb9e3ade5d26818701a90ea71d63843122ea0a5ac2679e9c0e69b599278"
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

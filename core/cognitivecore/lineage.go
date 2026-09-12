package cognitivecore

const (
	SourceRepositoryURL  = "https://github.com/o9nn/ecco9.git"
	SourceCommitSHA      = "1b22401ee8842fd1aa2769b688d588cd9f1f9ce9"
	ManifestSHA256       = "b382e99916a2364eb3945770835d3f907d42bbf2f548a7879a2e57c32590a5b7"
	CatalogueSHA256      = "3456bd87a074478b0e60485ec891a1d8d91a443f1aca504840f678dfc8b69b63"
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

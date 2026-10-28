package cognitivecore

const (
	SourceRepositoryURL  = "https://github.com/o9nn/ecco9.git"
	SourceCommitSHA      = "1b22401ee8842fd1aa2769b688d588cd9f1f9ce9"
	ManifestSHA256       = "861e931f4c28be92c87cfbd08c313f32926ab0ff4907dabe4e0f87a2d4d911c2"
	CatalogueSHA256      = "d3ab0722e27fcb900e4f61a49e19f06b100dd0ce6bc96831ba90890dccb5d1b8"
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

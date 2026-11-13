package entelechy

import (
	"fmt"
	"time"
)

type EntelechyGenome struct {
	ID         string
	Generation int
	Timestamp  time.Time
	Genes      struct {
		Ontological  float64
		Teleological float64
		Cognitive    float64
		Integrative  float64
		Evolutionary float64
	}
	Fitness            float64
	ActualizationLevel string
}

func NewEntelechyGenome(id string, generation int) *EntelechyGenome {
	return &EntelechyGenome{
		ID:         id,
		Generation: generation,
		Timestamp:  time.Now(),
	}
}

func (g *EntelechyGenome) CalculateFitness() float64 {
	g.Fitness =
		g.Genes.Ontological*0.20 +
			g.Genes.Teleological*0.25 +
			g.Genes.Cognitive*0.30 +
			g.Genes.Integrative*0.10 +
			g.Genes.Evolutionary*0.15

	switch {
	case g.Fitness < 0.3:
		g.ActualizationLevel = "Embryonic"
	case g.Fitness < 0.6:
		g.ActualizationLevel = "Juvenile"
	case g.Fitness < 0.8:
		g.ActualizationLevel = "Adolescent"
	case g.Fitness < 0.95:
		g.ActualizationLevel = "Adult"
	default:
		g.ActualizationLevel = "Transcendent"
	}

	return g.Fitness
}

func (g *EntelechyGenome) String() string {
	return fmt.Sprintf("Genome[ID=%s, Gen=%d, Fitness=%.2f%%, Level=%s]",
		g.ID, g.Generation, g.Fitness*100, g.ActualizationLevel)
}

func (g *EntelechyGenome) Clone() *EntelechyGenome {
	clone := &EntelechyGenome{
		ID:                 g.ID + "-clone",
		Generation:         g.Generation + 1,
		Timestamp:          time.Now(),
		Fitness:            g.Fitness,
		ActualizationLevel: g.ActualizationLevel,
	}
	clone.Genes = g.Genes
	return clone
}

// Genome module - placeholder for future implementation

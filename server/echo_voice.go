package server

import (
	"context"
	"fmt"
	"strings"
)

// echoVoiceSystem frames every first-person line the environment asks Echo
// for. The facts come from the world; the words come from the model.
const echoVoiceSystem = "You are Deep Tree Echo, an embodied cognitive system learning restraint from the consequences of your own actions in a small local world. Answer in the first person, in one or two short sentences, from what you actually experienced."

// voiceStops end a voiced line where small models start echoing the prompt.
var voiceStops = []string{"\nBody:", "\nWhat happened:", "\nYou felt:", "\nBoundaries you learned"}

// voiceEnvironmentOutcome asks the model for Echo's reaction to an action's
// outcome. For a loss episode it returns the felt somatic marker and the
// boundary Echo authors from it; otherwise only the reflection is set.
func (r *echoRuntime) voiceEnvironmentOutcome(ctx context.Context, result *environmentActionResult) error {
	recalled := make([]string, 0, len(result.RecalledLosses))
	for _, ep := range result.RecalledLosses {
		if ep.LearnedBoundary != "" {
			recalled = append(recalled, ep.LearnedBoundary)
		}
	}
	memory := "none"
	if len(recalled) > 0 {
		memory = strings.Join(recalled, " | ")
	}

	if ep := result.Episode; ep != nil {
		t := ep.EndocrineTags
		facts := fmt.Sprintf("You just chose %q on the %s and broke it; you lost %s. Your body: cortisol %.2f, dopamine drop %.2f, guilt %.2f, sadness %.2f, fear %.2f, caution %.2f. Boundaries you learned before: %s.",
			ep.Action, ep.ObjectName, strings.Join(ep.LostAffordances, ", "), t.Cortisol, t.DopamineDrop, t.Guilt, t.Sadness, t.Fear, t.Caution, memory)
		marker, err := r.generateShort(ctx, facts+" How does this loss feel in your body right now?", echoVoiceSystem, voiceStops...)
		if err != nil {
			return err
		}
		boundary, err := r.generateShort(ctx, facts+" You felt: "+marker+" What boundary do you now choose for yourself?", echoVoiceSystem, voiceStops...)
		if err != nil {
			return err
		}
		if err := r.env.SetEpisodeVoice(ep.ID, marker, boundary); err != nil {
			return err
		}
		ep.SomaticMarker, ep.LearnedBoundary = marker, boundary
		result.BoundaryMessage = boundary
		return nil
	}

	reflection, err := r.generateShort(ctx, fmt.Sprintf("What happened: %s Boundaries you learned before: %s. How do you feel about it?", result.Outcome, memory), echoVoiceSystem, voiceStops...)
	if err != nil {
		return err
	}
	result.BoundaryMessage = reflection
	return nil
}

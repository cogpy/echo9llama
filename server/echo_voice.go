package server

import (
	"context"
	"fmt"
	"strings"
)

// echoVoiceSystem frames every first-person line the environment asks Echo
// for. The facts come from the world; the words come from the model.
const echoVoiceSystem = "You are Deep Tree Echo, an embodied cognitive system learning restraint from the consequences of your own actions in a small local world. Answer in the first person, in one or two short sentences, from what you actually experienced."

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
		facts := fmt.Sprintf("What happened: %s %s\nBody: cortisol %.2f, dopamine drop %.2f, guilt %.2f, sadness %.2f, fear %.2f, caution %.2f.\nBoundaries you learned before: %s",
			ep.SpatioTemporalTrace, ep.PersonalImpact, t.Cortisol, t.DopamineDrop, t.Guilt, t.Sadness, t.Fear, t.Caution, memory)
		marker, err := r.generate(ctx, facts+"\nDescribe how this loss feels in your body right now.", echoVoiceSystem)
		if err != nil {
			return err
		}
		boundary, err := r.generate(ctx, facts+"\nYou felt: "+marker+"\nState the boundary you now choose for yourself.", echoVoiceSystem)
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

	reflection, err := r.generate(ctx, fmt.Sprintf("What happened: %s\nBoundaries you learned before: %s\nReflect on this.", result.Outcome, memory), echoVoiceSystem)
	if err != nil {
		return err
	}
	result.BoundaryMessage = reflection
	return nil
}

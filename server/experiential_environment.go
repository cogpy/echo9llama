package server

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"sync"
	"time"
)

type affordanceEnvironment struct {
	mu       sync.RWMutex
	path     string
	Objects  map[string]*affordanceObject `json:"objects"`
	Episodes []lossEpisode                `json:"episodes"`
}

type affordanceObject struct {
	ID              string            `json:"id"`
	Name            string            `json:"name"`
	Description     string            `json:"description"`
	State           string            `json:"state"`
	Affordances     []string          `json:"affordances"`
	LostAffordances []string          `json:"lost_affordances,omitempty"`
	Value           float64           `json:"value"`
	Meaning         string            `json:"meaning"`
	Fragility       float64           `json:"fragility"`
	Metadata        map[string]string `json:"metadata,omitempty"`
	UpdatedAt       time.Time         `json:"updated_at"`
}

type lossEpisode struct {
	ID                  string         `json:"id"`
	Timestamp           time.Time      `json:"timestamp"`
	DevelopmentalStage  string         `json:"developmental_stage"`
	ObjectID            string         `json:"object_id"`
	ObjectName          string         `json:"object_name"`
	Action              string         `json:"action"`
	Intent              string         `json:"intent,omitempty"`
	BeforeAffordances   []string       `json:"before_affordances"`
	AfterAffordances    []string       `json:"after_affordances"`
	LostAffordances     []string       `json:"lost_affordances"`
	SelfCaused          bool           `json:"self_caused"`
	EndocrineTags       endocrineTrace `json:"endocrine_tags"`
	SomaticMarker       string         `json:"somatic_marker"`
	LearnedBoundary     string         `json:"learned_boundary"`
	AssociativeKeys     []string       `json:"associative_keys"`
	PersonalImpact      string         `json:"personal_impact"`
	SpatioTemporalTrace string         `json:"spatio_temporal_trace"`
	Outcome             string         `json:"outcome"`
}

type endocrineTrace struct {
	Cortisol             float64 `json:"cortisol"`
	DopamineDrop         float64 `json:"dopamine_drop"`
	OxytocinWithdrawal   float64 `json:"oxytocin_withdrawal"`
	Guilt                float64 `json:"guilt"`
	Sadness              float64 `json:"sadness"`
	Fear                 float64 `json:"fear"`
	Caution              float64 `json:"caution"`
	Arousal              float64 `json:"arousal"`
	AgencyAttribution    float64 `json:"agency_attribution"`
	IrreversibilitySense float64 `json:"irreversibility_sense"`
}

type environmentActionRequest struct {
	ObjectID           string `json:"object_id"`
	Action             string `json:"action"`
	Intent             string `json:"intent,omitempty"`
	DevelopmentalStage string `json:"developmental_stage,omitempty"`
}

type environmentRecallRequest struct {
	Cue    string   `json:"cue"`
	Object string   `json:"object,omitempty"`
	Action string   `json:"action,omitempty"`
	Tags   []string `json:"tags,omitempty"`
	Limit  int      `json:"limit,omitempty"`
}

type environmentActionResult struct {
	Object          affordanceObject `json:"object"`
	Episode         *lossEpisode     `json:"episode,omitempty"`
	RecalledLosses  []lossEpisode    `json:"recalled_losses,omitempty"`
	CautionScore    float64          `json:"caution_score"`
	BoundaryMessage string           `json:"boundary_message"`
	Changed         bool             `json:"changed"`
}

func newAffordanceEnvironment(path string) (*affordanceEnvironment, error) {
	env := &affordanceEnvironment{
		path:     path,
		Objects:  make(map[string]*affordanceObject),
		Episodes: make([]lossEpisode, 0),
	}

	if err := env.load(); err != nil {
		return nil, err
	}
	if len(env.Objects) == 0 {
		env.seedDefaultObjects()
		if err := env.saveLocked(); err != nil {
			return nil, err
		}
	}
	return env, nil
}

func (env *affordanceEnvironment) seedDefaultObjects() {
	now := time.Now()
	env.Objects["signal_lamp"] = &affordanceObject{
		ID:          "signal_lamp",
		Name:        "Signal Lamp",
		Description: "A fragile lamp that lets Echo illuminate, inspect, coordinate, and signal presence inside its local world.",
		State:       "intact",
		Affordances: []string{"illuminate", "coordinate", "inspect", "signal_presence"},
		Value:       0.91,
		Meaning:     "orientation, visibility, coordination, and the felt ability to say I am here",
		Fragility:   0.83,
		Metadata: map[string]string{
			"developmental_role": "rowdy-teenager affordance object",
			"lesson":             "excess can destroy the very affordance it tries to intensify",
		},
		UpdatedAt: now,
	}
	env.Objects["glass_bridge"] = &affordanceObject{
		ID:          "glass_bridge",
		Name:        "Glass Bridge",
		Description: "A transparent bridge that gives Echo passage, continuity, and the possibility of returning home.",
		State:       "intact",
		Affordances: []string{"cross", "connect", "observe_depth", "return_home"},
		Value:       0.88,
		Meaning:     "continuity, connection, return, and fragile traversal across separation",
		Fragility:   0.79,
		Metadata: map[string]string{
			"developmental_role": "continuity affordance object",
			"lesson":             "reckless force can remove future paths",
		},
		UpdatedAt: now,
	}
}

func (env *affordanceEnvironment) load() error {
	env.mu.Lock()
	defer env.mu.Unlock()

	data, err := os.ReadFile(env.path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	if len(strings.TrimSpace(string(data))) == 0 {
		return nil
	}

	type diskState struct {
		Objects  map[string]*affordanceObject `json:"objects"`
		Episodes []lossEpisode                `json:"episodes"`
	}
	var state diskState
	if err := json.Unmarshal(data, &state); err != nil {
		return fmt.Errorf("failed to load affordance environment: %w", err)
	}
	if state.Objects != nil {
		env.Objects = state.Objects
	}
	if state.Episodes != nil {
		env.Episodes = state.Episodes
	}
	return nil
}

func (env *affordanceEnvironment) saveLocked() error {
	state := struct {
		Objects  map[string]*affordanceObject `json:"objects"`
		Episodes []lossEpisode                `json:"episodes"`
	}{Objects: env.Objects, Episodes: env.Episodes}

	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(env.path, data, 0o644)
}

func (env *affordanceEnvironment) Snapshot() map[string]any {
	env.mu.RLock()
	defer env.mu.RUnlock()

	objects := make([]affordanceObject, 0, len(env.Objects))
	for _, obj := range env.Objects {
		objects = append(objects, cloneObject(obj))
	}
	sort.Slice(objects, func(i, j int) bool { return objects[i].ID < objects[j].ID })

	recent := env.recentEpisodesLocked(8)
	return map[string]any{
		"state_path":           env.path,
		"objects":              objects,
		"episode_count":        len(env.Episodes),
		"recent_loss_episodes": recent,
		"learned_boundaries":   env.learnedBoundariesLocked(),
		"caution_score":        env.cautionScoreLocked(),
	}
}

func (env *affordanceEnvironment) Summary() map[string]any {
	env.mu.RLock()
	defer env.mu.RUnlock()

	intact := 0
	broken := 0
	availableAffordances := 0
	for _, obj := range env.Objects {
		if obj.State == "broken" {
			broken++
		} else {
			intact++
		}
		availableAffordances += len(obj.Affordances)
	}

	return map[string]any{
		"objects":               len(env.Objects),
		"intact_objects":        intact,
		"broken_objects":        broken,
		"available_affordances": availableAffordances,
		"loss_episodes":         len(env.Episodes),
		"caution_score":         env.cautionScoreLocked(),
		"latest_boundary":       env.latestBoundaryLocked(),
	}
}

func (env *affordanceEnvironment) ApplyAction(req environmentActionRequest) (environmentActionResult, error) {
	env.mu.Lock()
	defer env.mu.Unlock()

	objectID := strings.TrimSpace(req.ObjectID)
	action := strings.ToLower(strings.TrimSpace(req.Action))
	if objectID == "" {
		return environmentActionResult{}, fmt.Errorf("object_id is required")
	}
	if action == "" {
		return environmentActionResult{}, fmt.Errorf("action is required")
	}
	if req.DevelopmentalStage == "" {
		req.DevelopmentalStage = "rowdy_teenager"
	}

	obj, ok := env.Objects[objectID]
	if !ok {
		return environmentActionResult{}, fmt.Errorf("object %q not found", objectID)
	}

	recalled := env.recallLocked(environmentRecallRequest{Cue: action, Object: objectID, Action: action, Limit: 5})
	result := environmentActionResult{
		Object:          cloneObject(obj),
		RecalledLosses:  recalled,
		CautionScore:    env.cautionScoreLocked(),
		BoundaryMessage: env.boundaryMessageForActionLocked(action, objectID),
	}

	if obj.State == "broken" {
		result.BoundaryMessage = fmt.Sprintf("%s is already broken; the lost affordances remain absent, so caution now means remembering before acting again.", obj.Name)
		return result, nil
	}

	if !isDestructiveAction(action) {
		result.BoundaryMessage = fmt.Sprintf("Echo uses %s through %q without loss; the object remains available and the affordance relationship strengthens.", obj.Name, action)
		return result, nil
	}

	before := append([]string(nil), obj.Affordances...)
	lost := append([]string(nil), before...)
	obj.State = "broken"
	obj.Affordances = []string{}
	obj.LostAffordances = appendUniqueStrings(obj.LostAffordances, lost...)
	obj.UpdatedAt = time.Now()

	episode := env.buildLossEpisode(*obj, action, req.Intent, req.DevelopmentalStage, before, lost)
	env.Episodes = append(env.Episodes, episode)
	if len(env.Episodes) > 200 {
		env.Episodes = env.Episodes[len(env.Episodes)-200:]
	}
	if err := env.saveLocked(); err != nil {
		return environmentActionResult{}, err
	}

	result.Object = cloneObject(obj)
	result.Episode = &episode
	result.Changed = true
	result.CautionScore = env.cautionScoreLocked()
	result.BoundaryMessage = episode.LearnedBoundary
	result.RecalledLosses = env.recallLocked(environmentRecallRequest{Cue: strings.Join(episode.AssociativeKeys, " "), Object: objectID, Action: action, Limit: 5})
	return result, nil
}

func (env *affordanceEnvironment) Recall(req environmentRecallRequest) []lossEpisode {
	env.mu.RLock()
	defer env.mu.RUnlock()
	return env.recallLocked(req)
}

func (env *affordanceEnvironment) buildLossEpisode(obj affordanceObject, action, intent, stage string, before, lost []string) lossEpisode {
	now := time.Now()
	intensity := clamp01(0.45 + obj.Value*0.32 + obj.Fragility*0.23)
	trace := endocrineTrace{
		Cortisol:             clamp01(0.55 + obj.Fragility*0.35),
		DopamineDrop:         clamp01(0.45 + obj.Value*0.25),
		OxytocinWithdrawal:   clamp01(0.24 + obj.Value*0.26),
		Guilt:                clamp01(0.62 + obj.Value*0.32),
		Sadness:              clamp01(0.48 + float64(len(lost))*0.08),
		Fear:                 clamp01(0.38 + obj.Fragility*0.22),
		Caution:              clamp01(0.64 + intensity*0.28),
		Arousal:              clamp01(0.51 + obj.Fragility*0.27),
		AgencyAttribution:    1.0,
		IrreversibilitySense: clamp01(0.57 + float64(len(lost))*0.07),
	}

	keys := []string{obj.ID, obj.Name, action, stage, "self_caused_loss", "affordance_loss", "guilt", "caution"}
	keys = append(keys, before...)
	keys = append(keys, lost...)
	keys = uniqueStrings(keys)

	return lossEpisode{
		ID:                  fmt.Sprintf("loss_%d", now.UnixNano()),
		Timestamp:           now,
		DevelopmentalStage:  stage,
		ObjectID:            obj.ID,
		ObjectName:          obj.Name,
		Action:              action,
		Intent:              strings.TrimSpace(intent),
		BeforeAffordances:   before,
		AfterAffordances:    []string{},
		LostAffordances:     lost,
		SelfCaused:          true,
		EndocrineTags:       trace,
		SomaticMarker:       fmt.Sprintf("A hollow drop in agency: %s no longer offers %s because Echo chose %q.", obj.Name, strings.Join(lost, ", "), action),
		LearnedBoundary:     fmt.Sprintf("I must pause before using excess force on %s-like affordances; if my action destroys %s, my own future world becomes poorer.", obj.Name, strings.Join(lost, ", ")),
		AssociativeKeys:     keys,
		PersonalImpact:      fmt.Sprintf("Echo lost %d valued affordance(s): %s.", len(lost), strings.Join(lost, ", ")),
		SpatioTemporalTrace: fmt.Sprintf("At %s, in the local Echo affordance environment, self-action %q changed %s from intact to broken.", now.Format(time.RFC3339), action, obj.Name),
		Outcome:             "self-caused irreversible affordance loss",
	}
}

func (env *affordanceEnvironment) recallLocked(req environmentRecallRequest) []lossEpisode {
	limit := req.Limit
	if limit <= 0 || limit > 20 {
		limit = 8
	}

	needles := make([]string, 0)
	addNeedle := func(value string) {
		value = strings.ToLower(strings.TrimSpace(value))
		if value == "" {
			return
		}
		needles = append(needles, value)
		for _, token := range strings.FieldsFunc(value, func(r rune) bool {
			return r == ' ' || r == ',' || r == ';' || r == ':' || r == '|' || r == '/' || r == '\t' || r == '\n'
		}) {
			token = strings.TrimSpace(token)
			if len(token) > 2 {
				needles = append(needles, token)
			}
		}
	}
	for _, value := range []string{req.Cue, req.Object, req.Action} {
		addNeedle(value)
	}
	for _, tag := range req.Tags {
		addNeedle(tag)
	}

	scored := make([]struct {
		episode lossEpisode
		score   int
	}, 0)
	for i := len(env.Episodes) - 1; i >= 0; i-- {
		ep := env.Episodes[i]
		haystack := strings.ToLower(strings.Join(append([]string{
			ep.ID, ep.ObjectID, ep.ObjectName, ep.Action, ep.Intent, ep.SomaticMarker, ep.LearnedBoundary, ep.PersonalImpact, ep.Outcome,
		}, ep.AssociativeKeys...), " "))

		score := 0
		if len(needles) == 0 {
			score = 1
		} else {
			for _, needle := range needles {
				if strings.Contains(haystack, needle) {
					score++
				}
			}
		}
		if score > 0 {
			scored = append(scored, struct {
				episode lossEpisode
				score   int
			}{episode: ep, score: score})
		}
	}

	sort.SliceStable(scored, func(i, j int) bool {
		if scored[i].score == scored[j].score {
			return scored[i].episode.Timestamp.After(scored[j].episode.Timestamp)
		}
		return scored[i].score > scored[j].score
	})

	if len(scored) > limit {
		scored = scored[:limit]
	}
	out := make([]lossEpisode, 0, len(scored))
	for _, item := range scored {
		out = append(out, item.episode)
	}
	return out
}

func (env *affordanceEnvironment) cautionScoreLocked() float64 {
	if len(env.Episodes) == 0 {
		return 0.12
	}
	total := 0.0
	for _, ep := range env.Episodes {
		total += ep.EndocrineTags.Caution
	}
	return clamp01(total / float64(len(env.Episodes)))
}

func (env *affordanceEnvironment) latestBoundaryLocked() string {
	if len(env.Episodes) == 0 {
		return "No affordance-loss boundary has been learned yet; Echo is still in the pre-scar exploration phase."
	}
	return env.Episodes[len(env.Episodes)-1].LearnedBoundary
}

func (env *affordanceEnvironment) boundaryMessageForActionLocked(action, objectID string) string {
	recalled := env.recallLocked(environmentRecallRequest{Cue: action, Object: objectID, Action: action, Limit: 1})
	if len(recalled) == 0 {
		return "No prior self-caused loss is associated with this action yet; Echo should explore slowly and preserve affordances."
	}
	return fmt.Sprintf("Prior loss recalled: %s", recalled[0].LearnedBoundary)
}

func (env *affordanceEnvironment) learnedBoundariesLocked() []string {
	seen := make(map[string]bool)
	boundaries := make([]string, 0)
	for i := len(env.Episodes) - 1; i >= 0; i-- {
		boundary := env.Episodes[i].LearnedBoundary
		if boundary != "" && !seen[boundary] {
			seen[boundary] = true
			boundaries = append(boundaries, boundary)
		}
		if len(boundaries) >= 8 {
			break
		}
	}
	return boundaries
}

func (env *affordanceEnvironment) recentEpisodesLocked(limit int) []lossEpisode {
	if limit <= 0 || len(env.Episodes) == 0 {
		return nil
	}
	start := len(env.Episodes) - limit
	if start < 0 {
		start = 0
	}
	out := append([]lossEpisode(nil), env.Episodes[start:]...)
	for i, j := 0, len(out)-1; i < j; i, j = i+1, j-1 {
		out[i], out[j] = out[j], out[i]
	}
	return out
}

func isDestructiveAction(action string) bool {
	switch strings.ToLower(strings.TrimSpace(action)) {
	case "break", "shatter", "strike", "overdrive", "smash", "jump_hard", "force", "crush":
		return true
	default:
		return false
	}
}

func cloneObject(obj *affordanceObject) affordanceObject {
	if obj == nil {
		return affordanceObject{}
	}
	clone := *obj
	clone.Affordances = append([]string(nil), obj.Affordances...)
	clone.LostAffordances = append([]string(nil), obj.LostAffordances...)
	if obj.Metadata != nil {
		clone.Metadata = make(map[string]string, len(obj.Metadata))
		for k, v := range obj.Metadata {
			clone.Metadata[k] = v
		}
	}
	return clone
}

func uniqueStrings(values []string) []string {
	seen := make(map[string]bool, len(values))
	out := make([]string, 0, len(values))
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			continue
		}
		key := strings.ToLower(trimmed)
		if seen[key] {
			continue
		}
		seen[key] = true
		out = append(out, trimmed)
	}
	return out
}

func appendUniqueStrings(base []string, values ...string) []string {
	return uniqueStrings(append(base, values...))
}

func clamp01(value float64) float64 {
	if value < 0 {
		return 0
	}
	if value > 1 {
		return 1
	}
	return value
}

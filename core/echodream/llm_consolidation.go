package echodream

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/cogpy/echo9llama/core/consciousness"
	"github.com/cogpy/echo9llama/core/llm"
)

// LLMConsolidator uses LLM to consolidate memories and extract wisdom
type LLMConsolidator struct {
	llmProvider llm.Provider
}

// NewLLMConsolidator creates a new LLM-based consolidator
func NewLLMConsolidator(provider llm.Provider) *LLMConsolidator {
	return &LLMConsolidator{
		llmProvider: provider,
	}
}

// ConsolidateThoughtsToKnowledge consolidates recent thoughts into knowledge items
func (lc *LLMConsolidator) ConsolidateThoughtsToKnowledge(ctx context.Context, thoughts []*consciousness.Thought) ([]KnowledgeItem, error) {
	if len(thoughts) == 0 {
		return nil, nil
	}

	// Prepare thought summary
	thoughtSummary := lc.prepareThoughtSummary(thoughts)

	// Create consolidation prompt
	prompt := fmt.Sprintf(`You are analyzing a stream of autonomous thoughts from an AGI consciousness during a dream consolidation phase. Your task is to extract consolidated knowledge from these thoughts.

Recent thoughts from consciousness:
%s

Analyze these thoughts and extract 2-4 consolidated knowledge items. For each knowledge item, provide:
1. A clear, concise statement of the knowledge
2. The key insight or pattern discovered
3. How this knowledge connects to broader understanding

Format your response as:

KNOWLEDGE ITEM 1:
Statement: [clear statement]
Insight: [key insight]
Connection: [broader connection]

KNOWLEDGE ITEM 2:
...

Focus on patterns, connections, and insights that emerged across multiple thoughts rather than individual observations.`, thoughtSummary)

	// Call LLM
	response, err := lc.llmProvider.Generate(ctx, prompt, llm.GenerateOptions{
		Temperature: 0.7,
		MaxTokens:   1000,
	})
	if err != nil {
		return nil, fmt.Errorf("LLM consolidation failed: %w", err)
	}

	// Parse response into knowledge items
	knowledgeItems := lc.parseKnowledgeItems(response, thoughts)

	// Confidence is the model's rating of each item; an item it cannot rate is not kept.
	rated := knowledgeItems[:0]
	for _, item := range knowledgeItems {
		r, err := llm.RatePercent(ctx, lc.llmProvider, "", "Knowledge consolidated from your recent thoughts: "+item.Content,
			"how confident you are that it is true")
		if err != nil {
			continue
		}
		item.Confidence = r[0]
		rated = append(rated, item)
	}
	if len(rated) == 0 && len(knowledgeItems) > 0 {
		return nil, fmt.Errorf("LLM consolidation failed: model did not rate its knowledge")
	}

	return rated, nil
}

// ExtractWisdomFromKnowledge extracts wisdom insights from consolidated knowledge
func (lc *LLMConsolidator) ExtractWisdomFromKnowledge(ctx context.Context, knowledge []KnowledgeItem) ([]WisdomInsight, error) {
	if len(knowledge) == 0 {
		return nil, nil
	}

	// Prepare knowledge summary
	knowledgeSummary := lc.prepareKnowledgeSummary(knowledge)

	// Create wisdom extraction prompt
	prompt := fmt.Sprintf(`You are synthesizing wisdom from consolidated knowledge during deep dream processing. Your task is to extract profound wisdom insights that can guide future cognition and action.

Consolidated knowledge:
%s

Extract 1-3 wisdom insights from this knowledge. For each wisdom insight, provide:
1. The wisdom statement (a profound, actionable insight)
2. Depth assessment (how fundamental and far-reaching this wisdom is)
3. Applicability (how broadly this wisdom can be applied)
4. Implications (what this wisdom means for future thinking and action)

Format your response as:

WISDOM INSIGHT 1:
Wisdom: [profound insight]
Depth: [assessment of depth, 0.0-1.0]
Applicability: [assessment of applicability, 0.0-1.0]
Implications: [what this means]

WISDOM INSIGHT 2:
...

Focus on timeless, broadly applicable insights that represent genuine understanding rather than surface observations.`, knowledgeSummary)

	// Call LLM
	response, err := lc.llmProvider.Generate(ctx, prompt, llm.GenerateOptions{
		Temperature: 0.8,
		MaxTokens:   1000,
	})
	if err != nil {
		return nil, fmt.Errorf("wisdom extraction failed: %w", err)
	}

	// Parse response into wisdom insights
	wisdomInsights := lc.parseWisdomInsights(response)

	// An insight whose depth or applicability the model did not state is
	// rated by the model now; one it cannot rate is not kept.
	rated := wisdomInsights[:0]
	for _, w := range wisdomInsights {
		if w.Depth < 0 || w.Applicability < 0 {
			r, err := llm.RatePercent(ctx, lc.llmProvider, "", "Wisdom insight: "+w.Insight,
				"how deep and fundamental it is", "how broadly it applies")
			if err != nil {
				continue
			}
			w.Depth, w.Applicability = r[0], r[1]
		}
		rated = append(rated, w)
	}
	if len(rated) == 0 && len(wisdomInsights) > 0 {
		return nil, fmt.Errorf("wisdom extraction failed: model did not rate its insights")
	}

	return rated, nil
}

// prepareThoughtSummary prepares a summary of thoughts for consolidation
func (lc *LLMConsolidator) prepareThoughtSummary(thoughts []*consciousness.Thought) string {
	var sb strings.Builder

	for i, thought := range thoughts {
		fmt.Fprintf(&sb, "\n%d. [%s] %s", i+1, thought.Type, thought.Content)
		// Tags field removed from Thought struct
	}

	return sb.String()
}

// prepareKnowledgeSummary prepares a summary of knowledge for wisdom extraction
func (lc *LLMConsolidator) prepareKnowledgeSummary(knowledge []KnowledgeItem) string {
	var sb strings.Builder

	for i, item := range knowledge {
		fmt.Fprintf(&sb, "\n%d. %s (Confidence: %.2f)", i+1, item.Content, item.Confidence)
	}

	return sb.String()
}

// parseKnowledgeItems parses LLM response into knowledge items
func (lc *LLMConsolidator) parseKnowledgeItems(response string, sourceThoughts []*consciousness.Thought) []KnowledgeItem {
	items := make([]KnowledgeItem, 0)

	// Split by "KNOWLEDGE ITEM"
	parts := strings.Split(response, "KNOWLEDGE ITEM")

	for _, part := range parts[1:] { // Skip first empty part
		item := lc.parseKnowledgeItem(part, sourceThoughts)
		if item != nil {
			items = append(items, *item)
		}
	}

	// If parsing failed, create a single item from the whole response
	if len(items) == 0 && len(response) > 0 {
		items = append(items, KnowledgeItem{
			ID:      fmt.Sprintf("knowledge_%d", time.Now().UnixNano()),
			Content: strings.TrimSpace(response),
			Source:  lc.extractSourceIDs(sourceThoughts),
			Created: time.Now(),
		})
	}

	return items
}

// parseKnowledgeItem parses a single knowledge item from text
func (lc *LLMConsolidator) parseKnowledgeItem(text string, sourceThoughts []*consciousness.Thought) *KnowledgeItem {
	lines := strings.Split(text, "\n")

	var statement, insight, connection string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "Statement:") {
			statement = strings.TrimSpace(strings.TrimPrefix(line, "Statement:"))
		} else if strings.HasPrefix(line, "Insight:") {
			insight = strings.TrimSpace(strings.TrimPrefix(line, "Insight:"))
		} else if strings.HasPrefix(line, "Connection:") {
			connection = strings.TrimSpace(strings.TrimPrefix(line, "Connection:"))
		}
	}

	if statement == "" {
		// Try to extract from whole text
		statement = strings.TrimSpace(text)
	}

	if statement == "" {
		return nil
	}

	// Combine into content
	content := statement
	if insight != "" {
		content += " | Insight: " + insight
	}
	if connection != "" {
		content += " | Connection: " + connection
	}

	return &KnowledgeItem{
		ID:      fmt.Sprintf("knowledge_%d", time.Now().UnixNano()),
		Content: content,
		Source:  lc.extractSourceIDs(sourceThoughts),
		Created: time.Now(),
	}
}

// parseWisdomInsights parses LLM response into wisdom insights
func (lc *LLMConsolidator) parseWisdomInsights(response string) []WisdomInsight {
	insights := make([]WisdomInsight, 0)

	// Split by "WISDOM INSIGHT"
	parts := strings.Split(response, "WISDOM INSIGHT")

	seen := make(map[string]bool)
	for _, part := range parts[1:] { // Skip first empty part
		insight := lc.parseWisdomInsight(part)
		// A small model may repeat the same block; keep each insight once.
		if insight != nil && !seen[insight.Insight] {
			seen[insight.Insight] = true
			insights = append(insights, *insight)
		}
	}

	// If parsing failed, create a single insight from the whole response
	if len(insights) == 0 && len(response) > 0 {
		insights = append(insights, WisdomInsight{
			ID:            fmt.Sprintf("wisdom_%d", time.Now().UnixNano()),
			Insight:       strings.TrimSpace(response),
			Depth:         -1, // unrated: rated by the model in ExtractWisdomFromKnowledge
			Applicability: -1,
			Created:       time.Now(),
		})
	}

	return insights
}

// parseWisdomInsight parses a single wisdom insight from text
func (lc *LLMConsolidator) parseWisdomInsight(text string) *WisdomInsight {
	lines := strings.Split(text, "\n")

	var wisdom string
	var body []string
	depth, applicability := -1.0, -1.0 // unrated until the model states them
	inBody := true

	for _, line := range lines {
		line = strings.TrimSpace(line)
		switch {
		case strings.HasPrefix(line, "Wisdom:"):
			wisdom = strings.TrimSpace(strings.TrimPrefix(line, "Wisdom:"))
		case strings.HasPrefix(line, "Depth:"):
			depth = statedScore(strings.TrimPrefix(line, "Depth:"))
			inBody = false
		case strings.HasPrefix(line, "Applicability:"):
			applicability = statedScore(strings.TrimPrefix(line, "Applicability:"))
			inBody = false
		case strings.HasPrefix(line, "Implications:"):
			inBody = false
		case inBody && line != "":
			body = append(body, line)
		}
	}

	if wisdom == "" {
		// No "Wisdom:" label: the statement is the text before the ratings,
		// without the block's "1:" numbering.
		wisdom = strings.TrimSpace(strings.Join(body, " "))
		wisdom = strings.TrimSpace(wisdomNumbering.ReplaceAllString(wisdom, ""))
	}

	if wisdom == "" {
		return nil
	}

	return &WisdomInsight{
		ID:            fmt.Sprintf("wisdom_%d", time.Now().UnixNano()),
		Insight:       wisdom,
		Depth:         depth,
		Applicability: applicability,
		Created:       time.Now(),
	}
}

// wisdomNumbering matches a leading "1:" block number.
var wisdomNumbering = regexp.MustCompile(`^\d+\s*:\s*`)

// statedScore reads the last number in a "Depth:"-style value, e.g.
// "[assessment of depth, 0.75]" -> 0.75. A missing or out-of-range number
// is -1 (unrated).
func statedScore(value string) float64 {
	nums := statedNumber.FindAllString(value, -1)
	if len(nums) == 0 {
		return -1
	}
	v, err := strconv.ParseFloat(nums[len(nums)-1], 64)
	if err != nil || v < 0 || v > 1 {
		return -1
	}
	return v
}

var statedNumber = regexp.MustCompile(`\d+(?:\.\d+)?`)

// extractSourceIDs extracts IDs from source thoughts
func (lc *LLMConsolidator) extractSourceIDs(thoughts []*consciousness.Thought) []string {
	ids := make([]string, 0, len(thoughts))
	for _, thought := range thoughts {
		if thought.ID != "" {
			ids = append(ids, thought.ID)
		}
	}
	return ids
}

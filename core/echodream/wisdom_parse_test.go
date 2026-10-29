package echodream

import "testing"

// The reply below is what a locally served Qwen2.5-0.5B returned: no
// "Wisdom:" label, bracketed scores, and the same block repeated.
func TestParseWisdomInsightsSmallModelReply(t *testing.T) {
	block := "1:\nRest consolidates learning, so memory needs quiet.\n\nDepth: [assessment of depth, 0.75]\nApplicability: [assessment of applicability, 0.8]\nImplications: [rest matters]\n"
	reply := "WISDOM INSIGHT " + block + "WISDOM INSIGHT " + block

	got := (&LLMConsolidator{}).parseWisdomInsights(reply)
	if len(got) != 1 {
		t.Fatalf("insights = %d, want the repeated block once: %+v", len(got), got)
	}
	w := got[0]
	if w.Insight != "Rest consolidates learning, so memory needs quiet." || w.Depth != 0.75 || w.Applicability != 0.8 {
		t.Fatalf("parsed %+v", w)
	}
}

func TestParseWisdomInsightUnstatedScoresStayUnrated(t *testing.T) {
	w := (&LLMConsolidator{}).parseWisdomInsight(" 1:\nWisdom: Listen first.\nDepth: very deep\n")
	if w == nil || w.Insight != "Listen first." || w.Depth != -1 || w.Applicability != -1 {
		t.Fatalf("parsed %+v", w)
	}
}

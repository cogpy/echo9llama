package thinking

import (
	"testing"
)

func TestExtractThinking(t *testing.T) {
	tests := []struct {
		in, wantContent, wantThink string
	}{
		{
			in:          "<think> internal </think> world",
			wantThink:   "internal ",
			wantContent: "world",
		},
		{
			in:          "<think>a</think><think>b</think>c",
			wantThink:   "a",
			wantContent: "<think>b</think>c",
		},
		{
			in:          "no think",
			wantThink:   "",
			wantContent: "no think",
		},
	}
	for i, tt := range tests {
		parser := Parser{
			OpeningTag: "<think>",
			ClosingTag: "</think>",
		}
		gotThinking, gotContent := parser.AddContent(tt.in)
		if gotContent != tt.wantContent || gotThinking != tt.wantThink {
			t.Errorf("case %d: got (%q,%q), want (%q,%q)", i, gotThinking, gotContent, tt.wantThink, tt.wantContent)
		}
	}
}

func TestThinkingStreaming(t *testing.T) {
	type step struct {
		input          string
		wantThinking   string
		wantContent    string
		wantStateAfter thinkingState
	}

	cases := []struct {
		desc  string
		skip  bool
		steps []step
	}{
		{
			desc: "content without a thinking tag",
			steps: []step{
				{
					input:          "  abc",
					wantThinking:   "",
					wantContent:    "  abc",
					wantStateAfter: thinkingStateThinkingDone,
				},
			},
		},
		{
			desc: "content before a thinking tag nerfs the thinking tag",
			steps: []step{
				{
					input:          "  abc <think>def</think> ghi",
					wantThinking:   "",
					wantContent:    "  abc <think>def</think> ghi",
					wantStateAfter: thinkingStateThinkingDone,
				},
			},
		},
		{
			desc: "building up a thinking tag partially",
			steps: []step{
				{
					input:          "  <th",
					wantThinking:   "",
					wantContent:    "",
					wantStateAfter: thinkingStateLookingForOpening,
				},
				{
					input:          "in",
					wantThinking:   "",
					wantContent:    "",
					wantStateAfter: thinkingStateLookingForOpening,
				},
				{
					input:          "k>a",
					wantThinking:   "a",
					wantContent:    "",
					wantStateAfter: thinkingStateThinking,
				},
			},
		},
		{
			desc: "partial closing tag",
			steps: []step{
				{
					input:          "<think>abc</th",
					wantThinking:   "abc",
					wantContent:    "",
					wantStateAfter: thinkingStateThinking,
				},
				{
					input:          "ink>def",
					wantThinking:   "",
					wantContent:    "def",
					wantStateAfter: thinkingStateThinkingDone,
				},
			},
		},
		{
			desc: "partial closing tag fakeout",
			steps: []step{
				{
					input:          "<think>abc</th",
					wantThinking:   "abc",
					wantContent:    "",
					wantStateAfter: thinkingStateThinking,
				},
				{
					input:          "ing>def",
					wantThinking:   "</thing>def",
					wantContent:    "",
					wantStateAfter: thinkingStateThinking,
				},
				{
					input:          "ghi</thi",
					wantThinking:   "ghi",
					wantContent:    "",
					wantStateAfter: thinkingStateThinking,
				},
				{
					input:          "nk>jkl",
					wantThinking:   "",
					wantContent:    "jkl",
					wantStateAfter: thinkingStateThinkingDone,
				},
			},
		},
		{
			desc: "whitespace after thinking tag",
			steps: []step{
				{
					input:          "  <think>abc</think>\n\ndef",
					wantThinking:   "abc",
					wantContent:    "def",
					wantStateAfter: thinkingStateThinkingDone,
				},
			},
		},
		{
			desc: "whitespace after thinking tag (incremental)",
			steps: []step{
				{
					input:          "  <think>abc</think>",
					wantThinking:   "abc",
					wantContent:    "",
					wantStateAfter: thinkingStateThinkingDoneEatingWhitespace,
				},
				{
					input:          "\n\ndef",
					wantThinking:   "",
					wantContent:    "def",
					wantStateAfter: thinkingStateThinkingDone,
				},
			},
		},
		{
			desc: "whitespace after thinking tag with content and more whitespace",
			steps: []step{
				{
					input:          "  <think>abc</think>\n\ndef ",
					wantThinking:   "abc",
					wantContent:    "def ",
					wantStateAfter: thinkingStateThinkingDone,
				},
				{
					input:          " ghi",
					wantThinking:   "",
					wantContent:    " ghi",
					wantStateAfter: thinkingStateThinkingDone,
				},
			},
		},
		{
			desc: "token by token",
			steps: []step{
				{
					input:          "<think>",
					wantThinking:   "",
					wantContent:    "",
					wantStateAfter: thinkingStateThinkingStartedEatingWhitespace,
				},
				{
					input:          "\n",
					wantThinking:   "",
					wantContent:    "",
					wantStateAfter: thinkingStateThinkingStartedEatingWhitespace,
				},
				{
					input:          "</think>",
					wantThinking:   "",
					wantContent:    "",
					wantStateAfter: thinkingStateThinkingDoneEatingWhitespace,
				},
				{
					input:          "\n\n",
					wantThinking:   "",
					wantContent:    "",
					wantStateAfter: thinkingStateThinkingDoneEatingWhitespace,
				},
				{
					input:          "Hi",
					wantThinking:   "",
					wantContent:    "Hi",
					wantStateAfter: thinkingStateThinkingDone,
				},
				{
					input:          " there",
					wantThinking:   "",
					wantContent:    " there",
					wantStateAfter: thinkingStateThinkingDone,
				},
			},
		},
		{
			desc: "leading thinking whitespace",
			steps: []step{
				{
					input:          "  <think>   \t ",
					wantThinking:   "",
					wantContent:    "",
					wantStateAfter: thinkingStateThinkingStartedEatingWhitespace,
				},
				{
					input:          "  these are some ",
					wantThinking:   "these are some ",
					wantContent:    "",
					wantStateAfter: thinkingStateThinking,
				},
				{
					input:          "thoughts </think>  ",
					wantThinking:   "thoughts ",
					wantContent:    "",
					wantStateAfter: thinkingStateThinkingDoneEatingWhitespace,
				},
				{
					input:          "  more content",
					wantThinking:   "",
					wantContent:    "more content",
					wantStateAfter: thinkingStateThinkingDone,
				},
			},
		},
	}

	for _, c := range cases {
		parser := Parser{
			OpeningTag: "<think>",
			ClosingTag: "</think>",
		}
		if c.skip {
			continue
		}
		for i, step := range c.steps {
			thinking, content := parser.AddContent(step.input)
			if content != step.wantContent || thinking != step.wantThinking {
				t.Errorf("case %q (step %d): got (%q,%q), want (%q,%q)", c.desc, i, content, thinking, step.wantContent, step.wantThinking)
			}
			if parser.state != step.wantStateAfter {
				t.Errorf("case %q (step %d): got state %s, want %s", c.desc, i, parser.state, step.wantStateAfter)
			}
		}
	}
}

package metacognition

import (
	"strings"
	"testing"
)

func TestValidateJSONRejectsAmbiguousOrUnboundedPayload(t *testing.T) {
	for _, raw := range []string{`{"passed":false,"passed":true}`, `{"route":{"model":"a","model":"b"}}`, `{} {}`, `[1,2`, strings.Repeat("[", 34) + "0" + strings.Repeat("]", 34), strings.Repeat(" ", 512*1024+1)} {
		if err := ValidateJSON(raw); err == nil {
			t.Fatalf("invalid JSON accepted: %.80s", raw)
		}
	}
	for _, raw := range []string{`{"passed":false,"score":0}`, `{"a":[{},null,true,1,"x"]}`, `[]`} {
		if err := ValidateJSON(raw); err != nil {
			t.Fatalf("valid JSON rejected: %s: %v", raw, err)
		}
	}
}

package claudecli

import (
	"encoding/json"
	"fmt"
)

// ResultEnvelope is the decoded shape of `claude -p --output-format json`
// stdout. See specs/002-claude-cli-backend/data-model.md
// §`claudecli.ResultEnvelope`.
type ResultEnvelope struct {
	Type    string `json:"type"`
	Subtype string `json:"subtype"`
	Result  string `json:"result"`
	IsError bool   `json:"is_error"`
	// Model is the resolved model id pulled from `modelUsage` (the key set
	// of which lists every model the call billed against). The key from
	// the most-billed model is used; for the single-call identify flow
	// modelUsage has exactly one entry.
	Model string `json:"-"`
	Error string `json:"error"`
	// StructuredOutput is the schema-validated structured payload claude
	// emits at the envelope level when --json-schema is supplied. When
	// present this should be preferred over JSON-decoding the textual
	// `result` separately.
	StructuredOutput json.RawMessage `json:"structured_output"`

	// modelUsage is decoded internally to populate Model.
	ModelUsage map[string]json.RawMessage `json:"modelUsage"`
}

// DecodeEnvelope parses one envelope from the supplied stdout bytes.
// Tolerates additional envelope-level keys (forward compat with newer
// `claude` CLI versions).
func DecodeEnvelope(stdout []byte) (ResultEnvelope, error) {
	if len(stdout) == 0 {
		return ResultEnvelope{}, fmt.Errorf("empty envelope")
	}
	var env ResultEnvelope
	if err := json.Unmarshal(stdout, &env); err != nil {
		return ResultEnvelope{}, fmt.Errorf("decode envelope: %w", err)
	}
	env.Model = pickModelFromUsage(env.ModelUsage)
	return env, nil
}

// pickModelFromUsage returns the first non-empty key in modelUsage. The
// identify flow makes a single subprocess call against a single model,
// so the map has at most one entry; this helper just pulls that key.
func pickModelFromUsage(usage map[string]json.RawMessage) string {
	for k := range usage {
		if k != "" {
			return k
		}
	}
	return ""
}

package prompt

import _ "embed"

// IdentifySchemaVersion mirrors IdentifyGameVersion so the prompt body
// and the response schema bump together; see
// specs/002-claude-cli-backend/contracts/claude-prompt.md.
const IdentifySchemaVersion = IdentifyGameVersion

//go:embed identify_schema_v1.json
var identifySchemaV1 string

// IdentifySchemaV1 returns the canonical JSON Schema string passed to
// `claude --json-schema` to constrain the model's identify response.
func IdentifySchemaV1() string { return identifySchemaV1 }

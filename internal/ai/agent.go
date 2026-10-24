package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/jscyril/tailorcv/internal/domain"
)

const (
	maxAgentSteps       = 6
	submitProposalsTool = "submit_proposals"
)

type Tool struct {
	Name        string
	Description string
	Parameters  json.RawMessage
}

// ToolHandler runs one tool call. Returning done ends the agent loop.
type ToolHandler func(name string, args json.RawMessage) (result map[string]any, done bool)

// AgentProvider is a Provider that can call tools in a loop.
type AgentProvider interface {
	Provider
	RunAgent(ctx context.Context, model, prompt string, tools []Tool, handle ToolHandler) error
}

var submitProposalsParameters = json.RawMessage(`{
  "type": "object",
  "properties": {"proposals": ` + proposalsSchema + `},
  "required": ["proposals"]
}`)

// RunProposalAgent lets the model submit proposals, learn which ones fail
// evidence validation, and resubmit. It returns the last submission as a
// Response so callers validate it exactly like a single-shot Generate.
func RunProposalAgent(ctx context.Context, provider AgentProvider, model string, request Request) ([]byte, error) {
	prompt, err := Prompt(request)
	if err != nil {
		return nil, err
	}
	prompt += "\n\nDo not reply with text. Call " + submitProposalsTool + " with the proposals. It validates them against the supplied facts; if it reports errors, fix every one and call it again."
	var last *Response
	handle := func(name string, args json.RawMessage) (map[string]any, bool) {
		if name != submitProposalsTool {
			return map[string]any{"error": "unknown tool " + name}, false
		}
		var input struct {
			Proposals []domain.AIProposal `json:"proposals"`
		}
		if err := json.Unmarshal(args, &input); err != nil {
			return map[string]any{"accepted": false, "errors": []string{"invalid arguments: " + err.Error()}}, false
		}
		response := Response{SchemaVersion: SchemaVersion, Proposals: input.Proposals}
		last = &response
		if errors := ValidateProposals(request, response); len(errors) > 0 {
			return map[string]any{"accepted": false, "errors": errors}, false
		}
		return map[string]any{"accepted": true}, true
	}
	tool := Tool{Name: submitProposalsTool, Description: "Submit rewritten proposals for validation.", Parameters: submitProposalsParameters}
	if err := provider.RunAgent(ctx, model, prompt, []Tool{tool}, handle); err != nil {
		return nil, err
	}
	if last == nil {
		return nil, fmt.Errorf("agent finished without submitting proposals")
	}
	return json.Marshal(last)
}

func (provider *Gemini) RunAgent(ctx context.Context, model, prompt string, tools []Tool, handle ToolHandler) error {
	model = strings.TrimSpace(strings.TrimPrefix(model, "models/"))
	if model == "" {
		return fmt.Errorf("select a Gemini model")
	}
	declarations := make([]map[string]any, len(tools))
	for index, tool := range tools {
		declarations[index] = map[string]any{"name": tool.Name, "description": tool.Description, "parametersJsonSchema": tool.Parameters}
	}
	contents := []any{map[string]any{"role": "user", "parts": []any{map[string]string{"text": prompt}}}}
	for step := 0; step < maxAgentSteps; step++ {
		payload, err := json.Marshal(map[string]any{
			"contents":         contents,
			"tools":            []any{map[string]any{"functionDeclarations": declarations}},
			"toolConfig":       map[string]any{"functionCallingConfig": map[string]string{"mode": "ANY"}},
			"generationConfig": map[string]any{"temperature": 0},
		})
		if err != nil {
			return fmt.Errorf("encode Gemini agent request: %w", err)
		}
		data, err := provider.generateContent(ctx, model, payload)
		if err != nil {
			return err
		}
		var result struct {
			Candidates []struct {
				Content json.RawMessage `json:"content"`
			} `json:"candidates"`
		}
		if err := json.Unmarshal(data, &result); err != nil || len(result.Candidates) == 0 {
			return fmt.Errorf("Gemini returned no agent response")
		}
		var content struct {
			Parts []struct {
				FunctionCall *struct {
					Name string          `json:"name"`
					Args json.RawMessage `json:"args"`
				} `json:"functionCall"`
			} `json:"parts"`
		}
		if err := json.Unmarshal(result.Candidates[0].Content, &content); err != nil {
			return fmt.Errorf("decode Gemini agent response: %w", err)
		}
		// The model turn is echoed back verbatim so any thought signatures survive.
		contents = append(contents, result.Candidates[0].Content)
		responses := make([]any, 0, len(content.Parts))
		finished := false
		for _, part := range content.Parts {
			if part.FunctionCall == nil {
				continue
			}
			output, done := handle(part.FunctionCall.Name, part.FunctionCall.Args)
			finished = finished || done
			responses = append(responses, map[string]any{"functionResponse": map[string]any{"name": part.FunctionCall.Name, "response": output}})
		}
		if len(responses) == 0 {
			return fmt.Errorf("Gemini replied without calling a tool")
		}
		if finished {
			return nil
		}
		contents = append(contents, map[string]any{"role": "user", "parts": responses})
	}
	return nil
}

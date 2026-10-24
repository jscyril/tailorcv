package ai

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
)

// scriptedAgent replays tool calls in order and records each tool result.
type scriptedAgent struct {
	calls   []string
	results []map[string]any
}

func (*scriptedAgent) Name() string                                              { return "scripted" }
func (*scriptedAgent) Models(context.Context) ([]string, error)                  { return nil, nil }
func (*scriptedAgent) Generate(context.Context, string, Request) ([]byte, error) { return nil, nil }
func (agent *scriptedAgent) RunAgent(_ context.Context, _, _ string, tools []Tool, handle ToolHandler) error {
	for _, args := range agent.calls {
		result, done := handle(tools[0].Name, json.RawMessage(args))
		agent.results = append(agent.results, result)
		if done {
			return nil
		}
	}
	return nil
}

func TestLatexFixAgentReturnsFirstSourceThatCompiles(t *testing.T) {
	original := "\\begin{document}\nLine one\n\\textbf{Line two\n\\end{document}"
	broken, _ := json.Marshal(map[string]string{"source": original + "}"})
	fixed, _ := json.Marshal(map[string]string{"source": strings.Replace(original, "two", "two}", 1)})
	agent := &scriptedAgent{calls: []string{string(broken), string(fixed)}}
	compile := func(_ context.Context, source string) (bool, string, error) {
		return strings.Contains(source, "two}"), "Runaway argument", nil
	}
	got, err := RunLatexFixAgent(context.Background(), agent, "m", original, "Runaway argument", compile)
	if err != nil || !strings.Contains(got, "two}") || len(agent.results) != 2 {
		t.Fatalf("RunLatexFixAgent() = %q, %v after %d calls", got, err, len(agent.results))
	}
	if removedLines(original, "\\begin{document}\\end{document}") <= 1 {
		t.Fatal("removedLines() missed deleted content")
	}
	// A large deletion is rejected before compiling.
	lines := strings.Repeat("content line\n", 20)
	agent = &scriptedAgent{calls: []string{`{"source":"\\end{document}"}`}}
	if _, err := RunLatexFixAgent(context.Background(), agent, "m", lines, "", func(context.Context, string) (bool, string, error) { return true, "", nil }); err == nil {
		t.Fatal("accepted a fix that deleted the document")
	}
}

func TestEvidenceSelectionAgentRejectsUnknownFacts(t *testing.T) {
	candidates := []Fact{{ID: "a", Text: "Built Go APIs"}, {ID: "b", Text: "Led migrations"}}
	agent := &scriptedAgent{calls: []string{`{"factIds":["a","zzz"]}`, `{"factIds":["b","a"]}`}}
	got, err := RunEvidenceSelectionAgent(context.Background(), agent, "m", JobRequirements{Role: "Engineer"}, candidates, 8)
	if err != nil || strings.Join(got, ",") != "b,a" {
		t.Fatalf("RunEvidenceSelectionAgent() = %v, %v", got, err)
	}
	if errs := agent.results[0]["errors"].([]string); !strings.Contains(errs[0], "zzz") {
		t.Fatalf("first result = %#v", agent.results[0])
	}
}

func TestFollowUpAgentRejectsInventedNumbers(t *testing.T) {
	followUp := FollowUp{Company: "Acme", Role: "Platform Engineer", AppliedOn: "2026-10-01", Sender: "Ada"}
	facts := []Fact{{ID: "f", Text: "Cut deploy time by 40%"}}
	base := "Hello, thank you again for considering my application for the Platform Engineer role at Acme. I remain very interested and would welcome the chance to talk about how I cut deploy time "
	invented, _ := json.Marshal(map[string]string{"email": base + "by 75%. Best, Ada"})
	supported, _ := json.Marshal(map[string]string{"email": base + "by 40%. Best, Ada"})
	agent := &scriptedAgent{calls: []string{string(invented), string(supported)}}
	email, err := RunFollowUpAgent(context.Background(), agent, "m", followUp, facts)
	if err != nil || !strings.Contains(email, "40%") {
		t.Fatalf("RunFollowUpAgent() = %q, %v", email, err)
	}
	if errs := agent.results[0]["errors"].([]string); !strings.Contains(strings.Join(errs, ";"), "75%") {
		t.Fatalf("first result = %#v", agent.results[0])
	}
}

func TestResumeImportAgentRequiresVerbatimText(t *testing.T) {
	plain := "Ada Lovelace\nExperience\nAnalytical Engines Ltd — Software Engineer, Jan 2024 – Present\nBuilt the first published algorithm for the engine."
	reworded := `{"profile":{"name":"Ada Lovelace"},"experiences":[{"company":"Analytical Engines Ltd","title":"Software Engineer","startDate":"2024-01","current":true,"bullets":["Pioneered algorithms for the engine."]}],"projects":[],"educations":[]}`
	exact := strings.Replace(reworded, "Pioneered algorithms for the engine.", "Built the first published algorithm for the engine.", 1)
	agent := &scriptedAgent{calls: []string{reworded, exact}}
	data, err := RunResumeImportAgent(context.Background(), agent, "m", "", plain)
	if err != nil || len(data.Experiences) != 1 || data.Experiences[0].Bullets[0].Text != "Built the first published algorithm for the engine." {
		t.Fatalf("RunResumeImportAgent() = %#v, %v", data, err)
	}
	if errs := agent.results[0]["errors"].([]string); !strings.Contains(errs[0], "word for word") {
		t.Fatalf("first result = %#v", agent.results[0])
	}
}

// @vitest-environment jsdom

import { cleanup, render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach, describe, expect, it, vi } from "vitest";
import { useState } from "react";
import { AIWorkspace, type AIRun, type AIWorkspaceProps, type CredentialStatus } from "./AIWorkspace";

const job = { id: "job-1", company: "Example", role: "Platform Engineer", description: "Build reliable releases", createdAt: "", updatedAt: "" };
const analysis = {
  job, score: 90, matchedSkills: ["Go"], unmentionedSkills: [], detectedSkills: ["Go"], requiredSkills: ["Go"], preferredSkills: [], responsibilities: ["Build releases"], keywords: ["release"], searchTerms: ["release"], explanation: "Strong match",
  rankedEvidence: [{ factId: "fact-1", sourceId: "experience-1", sourceType: "experience" as const, sourceLabel: "Example · Engineer", text: "Reduced deployment time by 40% with an audited Go release pipeline", score: 95, matchedSkills: ["Go"], reasons: ["Required skill"], verified: true, selectable: true }],
};
const proposal = { targetFactId: "fact-1", supportingFactIds: ["fact-1"], text: "Reduced deployment time by 40% through an audited Go release pipeline." };

function run(overrides: Partial<AIRun> = {}): AIRun {
  return { id: "run-1", jobId: job.id, provider: "ollama", model: "gemma4:12b", promptVersion: "prompt-v1", schemaVersion: "schema-v1", selectedFactIds: ["fact-1"], validationPassed: true, failureCategory: "", validationErrors: [], proposals: [proposal], resumeVersionId: "", createdAt: "2026-08-06T12:00:00Z", acceptedAt: "", ...overrides };
}

function props(overrides: Partial<AIWorkspaceProps> = {}): AIWorkspaceProps {
  return {
    provider: "ollama", endpoint: "http://127.0.0.1:11434", status: { provider: "ollama", endpoint: "http://127.0.0.1:11434", available: true, models: ["gemma4:12b"], message: "Ollama is available." }, model: "gemma4:12b",
    credential: { configured: false, message: "Gemini API key is not configured." }, apiKey: "", run: null, runs: [], proposals: [], acceptedProposalIDs: [], job, analysis, selectedFactIDs: ["fact-1"], templateName: "Jake style", busy: "",
    onProviderChange: vi.fn(), onEndpointChange: vi.fn(), onModelChange: vi.fn(), onAPIKeyChange: vi.fn(), onSaveCredential: vi.fn(), onDeleteCredential: vi.fn(), onCheck: vi.fn(), onGenerate: vi.fn(), onCancel: vi.fn(), onProposalChange: vi.fn(), onToggleProposal: vi.fn(), onAccept: vi.fn(), onOpenJob: vi.fn(), ...overrides,
  };
}

afterEach(cleanup);

describe("AIWorkspace provider setup", () => {
  it("switches to Gemini and keeps the credential field write-only after save", async () => {
    const user = userEvent.setup();
    function Harness() {
      const [provider, setProvider] = useState<"ollama" | "gemini">("ollama");
      const [apiKey, setAPIKey] = useState("");
      const [credential, setCredential] = useState<CredentialStatus>({ configured: false, message: "Gemini API key is not configured." });
      return <AIWorkspace {...props({ provider, apiKey, credential, status: null, model: "", onProviderChange: setProvider, onAPIKeyChange: setAPIKey, onSaveCredential: () => { setCredential({ configured: true, message: "Gemini API key is stored in the OS keyring." }); setAPIKey(""); }, onDeleteCredential: () => setCredential({ configured: false, message: "Gemini API key is not configured." }) })} />;
    }
    render(<Harness />);
    expect(screen.getByRole("heading", { name: "Ollama review" })).toBeTruthy();
    await user.click(screen.getByRole("button", { name: "Gemini" }));
    const keyInput = screen.getByLabelText("Gemini API key") as HTMLInputElement;
    expect(keyInput.type).toBe("password");
    await user.type(keyInput, "test-api-key");
    await user.click(screen.getByRole("button", { name: "Save to keyring" }));
    expect((screen.getByLabelText("Gemini API key") as HTMLInputElement).value).toBe("");
    expect(screen.getByText("Gemini API key is stored in the OS keyring.")).toBeTruthy();
    await user.click(screen.getByRole("button", { name: "Remove key" }));
    expect(screen.getByText("Gemini API key is not configured.")).toBeTruthy();
  });

  it("reports connection and model-selection interactions", async () => {
    const user = userEvent.setup();
    const onCheck = vi.fn();
    const onModelChange = vi.fn();
    render(<AIWorkspace {...props({ model: "", onCheck, onModelChange })} />);
    await user.click(screen.getByRole("button", { name: "Check connection" }));
    await user.selectOptions(screen.getByLabelText("Local model"), "gemma4:12b");
    expect(onCheck).toHaveBeenCalledOnce();
    expect(onModelChange).toHaveBeenCalledWith("gemma4:12b");
  });

  it.each([
    ["ollama" as const, "gemma4:12b"],
    ["gemini" as const, "gemini-test"],
  ])("enables %s generation only after connection, model, job, and evidence are ready", async (provider, model) => {
    const user = userEvent.setup();
    const onGenerate = vi.fn();
    render(<AIWorkspace {...props({ provider, model, credential: { configured: provider === "gemini", message: "Configured" }, status: { provider, endpoint: provider === "ollama" ? "http://127.0.0.1:11434" : "", available: true, models: [model], message: "Available" }, onGenerate })} />);
    await user.click(screen.getByRole("button", { name: "Generate proposals" }));
    expect(onGenerate).toHaveBeenCalledOnce();
    expect(screen.getByText(`Only normalized requirements and these selected facts will be sent to ${provider === "ollama" ? "Ollama" : "Gemini"}.`)).toBeTruthy();
  });
});

describe("AIWorkspace generation and review", () => {
  it("exposes cancellation while generation is active", async () => {
    const user = userEvent.setup();
    const onCancel = vi.fn();
    render(<AIWorkspace {...props({ busy: "generate", onCancel })} />);
    expect((screen.getByRole("button", { name: "Generating…" }) as HTMLButtonElement).disabled).toBe(true);
    await user.click(screen.getByRole("button", { name: "Cancel" }));
    expect(onCancel).toHaveBeenCalledOnce();
  });

  it("shows a blocked run without an acceptance action", () => {
    render(<AIWorkspace {...props({ run: run({ validationPassed: false, failureCategory: "validation", validationErrors: ["proposal introduces unsupported metrics: 75%"], proposals: [] }) })} />);
    expect(screen.getByText("Run blocked by validation")).toBeTruthy();
    expect(screen.getByText("proposal introduces unsupported metrics: 75%")).toBeTruthy();
    expect(screen.queryByRole("button", { name: "Accept new resume version" })).toBeNull();
  });

  it("edits, excludes, and accepts reviewed proposals", async () => {
    const user = userEvent.setup();
    const onAccept = vi.fn();
    function Harness() {
      const [proposals, setProposals] = useState([proposal]);
      const [accepted, setAccepted] = useState([proposal.targetFactId]);
      return <AIWorkspace {...props({ run: run(), runs: [run()], proposals, acceptedProposalIDs: accepted, onProposalChange: (factID, text) => setProposals((current) => current.map((item) => item.targetFactId === factID ? { ...item, text } : item)), onToggleProposal: (factID) => setAccepted((current) => current.includes(factID) ? [] : [factID]), onAccept })} />;
    }
    render(<Harness />);
    const editor = screen.getByLabelText("Proposed wording for fact-1") as HTMLTextAreaElement;
    await user.clear(editor);
    await user.type(editor, "Improved an audited Go release pipeline while preserving the supported evidence.");
    expect(editor.value).toContain("preserving the supported evidence");
    await user.click(screen.getByRole("checkbox"));
    expect(editor.disabled).toBe(true);
    expect((screen.getByRole("button", { name: "Accept new resume version" }) as HTMLButtonElement).disabled).toBe(true);
    await user.click(screen.getByRole("checkbox"));
    await user.click(screen.getByRole("button", { name: "Accept new resume version" }));
    expect(onAccept).toHaveBeenCalledOnce();
    expect(screen.getByText("Validated")).toBeTruthy();
  });
});

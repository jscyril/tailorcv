// @vitest-environment jsdom

import { cleanup, render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import App from "./App";

const bindings = vi.hoisted(() => ({
  AcceptAITailoring: vi.fn(), AnalyzeJobDescription: vi.fn(), CancelAITailoring: vi.fn(), CheckGemini: vi.fn(), CheckOllama: vi.fn(),
  CompileLatex: vi.fn(), CompileResumeVersion: vi.fn(), CreateResumeVersion: vi.fn(), DeleteAchievement: vi.fn(), DeleteCertification: vi.fn(),
  DeleteEducation: vi.fn(), DeleteExperience: vi.fn(), DeleteGeminiAPIKey: vi.fn(), DeleteJob: vi.fn(), DeleteProject: vi.fn(), DeleteResumeTemplate: vi.fn(),
  ExportCompiledPDF: vi.fn(), ExportLatexSource: vi.fn(), ExportProfileBackup: vi.fn(), GenerateAITailoring: vi.fn(), GetAISettings: vi.fn(),
  GetGeminiCredentialStatus: vi.fn(), GetProfile: vi.fn(), GetSelectedResumeTemplateID: vi.fn(), ImportGitHubProjects: vi.fn(), ImportProfileBackup: vi.fn(),
  ImportResumeTemplate: vi.fn(), ListAchievements: vi.fn(), ListAIRuns: vi.fn(), ListApplications: vi.fn(), ListCertifications: vi.fn(), ListEducations: vi.fn(),
  ListExperiences: vi.fn(), ListJobs: vi.fn(), ListProjects: vi.fn(), ListResumeTemplates: vi.fn(), RenderResumeTemplate: vi.fn(), SaveAchievement: vi.fn(),
  SaveAISettings: vi.fn(), SaveCertification: vi.fn(), SaveEducation: vi.fn(), SaveExperience: vi.fn(), SaveGeminiAPIKey: vi.fn(), SaveProfile: vi.fn(),
  SaveProject: vi.fn(), SaveResumeTemplate: vi.fn(), SaveResumeVersionEdit: vi.fn(), SelectResumeTemplate: vi.fn(), UpdateApplicationStatus: vi.fn(),
}));

vi.mock("../wailsjs/go/main/App", () => bindings);

const fact = {
  id: "fact-1", text: "Built reliable Go deployment services for production workloads", provenance: "manual", sourceUrl: "",
  verification: "verified", importance: "essential", position: 0, createdAt: "2026-01-01T00:00:00Z", updatedAt: "2026-01-01T00:00:00Z",
};
const experience = {
  id: "experience-1", company: "Example Systems", title: "Platform Engineer", location: "", startDate: "2024-01", endDate: "", current: true,
  position: 0, bullets: [fact], createdAt: "2026-01-01T00:00:00Z", updatedAt: "2026-01-01T00:00:00Z",
};
const profile = {
  name: "Ada Lovelace", headline: "Platform Engineer", email: "ada@example.com", phone: "", location: "", website: "",
  githubUsername: "", linkedInUrl: "", summary: "Builds reliable systems.", skills: ["Go"], contactLinks: [], updatedAt: "2026-01-01T00:00:00Z",
};
const template = { id: "builtin-jake-style", name: "Jake style", description: "Built in", source: "resume", builtIn: true, createdAt: "", updatedAt: "" };
const job = {
  id: "job-1", company: "Fictional Cloud", role: "Senior Platform Engineer",
  description: "Build reliable Go deployment services for secure production systems and improve release operations.",
  createdAt: "2026-01-01T00:00:00Z", updatedAt: "2026-01-01T00:00:00Z",
};
const evidence = {
  factId: fact.id, sourceId: experience.id, sourceType: "experience", sourceLabel: "Platform Engineer · Example Systems", text: fact.text,
  score: 92, matchedSkills: ["Go"], reasons: ["Matches Go", "Marked essential by you", "Role is current"], verified: true, selectable: true,
};
const analysis = {
  job, score: 100, matchedSkills: ["Go"], unmentionedSkills: [], detectedSkills: ["Go"], requiredSkills: ["Go"], preferredSkills: [],
  responsibilities: ["Build reliable Go deployment services"], keywords: ["deployment"], searchTerms: ["Go", "deployment"], rankedEvidence: [evidence], explanation: "Strong match",
};
const version = {
  id: "version-1", applicationId: "application-1", versionNumber: 1, jobDescriptionSnapshot: job.description, selectedFactIds: [fact.id],
  latexSource: "\\documentclass{article}\\begin{document}Built reliable Go deployment services\\end{document}", templateId: template.id,
  rankingExplanations: [evidence], contentHash: "hash", compileSuccess: false, compileEngine: "", compileDurationMs: 0,
  compileDiagnostics: [], compiledAt: "", pdfAvailable: false, createdAt: "2026-01-01T00:00:00Z",
};

beforeEach(() => {
  for (const mock of Object.values(bindings)) mock.mockReset();
  bindings.GetProfile.mockResolvedValue(profile);
  bindings.ListExperiences.mockResolvedValue([experience]);
  bindings.ListProjects.mockResolvedValue([]);
  bindings.ListEducations.mockResolvedValue([]);
  bindings.ListCertifications.mockResolvedValue([]);
  bindings.ListAchievements.mockResolvedValue([]);
  bindings.ListResumeTemplates.mockResolvedValue([template]);
  bindings.GetSelectedResumeTemplateID.mockResolvedValue(template.id);
  bindings.ListJobs.mockResolvedValue([]);
  bindings.ListApplications.mockResolvedValue([]);
  bindings.ListAIRuns.mockResolvedValue([]);
  bindings.GetAISettings.mockResolvedValue({ provider: "ollama", ollamaEndpoint: "http://127.0.0.1:11434", ollamaModel: "", geminiModel: "" });
  bindings.GetGeminiCredentialStatus.mockResolvedValue({ configured: false, message: "Gemini API key is not configured." });
  bindings.RenderResumeTemplate.mockResolvedValue("\\documentclass{article}\\begin{document}Preview\\end{document}");
  bindings.SaveAISettings.mockImplementation(async (settings) => settings);
  bindings.AnalyzeJobDescription.mockResolvedValue(analysis);
  bindings.CreateResumeVersion.mockResolvedValue({ application: { id: "application-1", jobId: job.id, status: "draft", selectedFactIds: [fact.id], versions: [version], createdAt: "", updatedAt: "" }, version });
});

afterEach(cleanup);

async function analyzeJobThroughUI() {
  const user = userEvent.setup();
  render(<App />);
  await user.click(await screen.findByRole("button", { name: "Job match" }));
  await user.type(screen.getByLabelText("Company"), job.company);
  await user.type(screen.getByLabelText("Role"), job.role);
  await user.type(screen.getByPlaceholderText("Paste the complete role description here…"), job.description);
  await user.click(screen.getByRole("button", { name: "Save & rank evidence" }));
  await screen.findByText("Ranked career evidence");
  return user;
}

describe("App Wails workflow integration", () => {
  it("analyzes a job and creates a deterministic resume version through bindings", async () => {
    const user = await analyzeJobThroughUI();
    expect(bindings.AnalyzeJobDescription).toHaveBeenCalledWith({ id: "", company: job.company, role: job.role, description: job.description });
    expect((screen.getByRole("checkbox") as HTMLInputElement).checked).toBe(true);

    await user.click(screen.getByRole("button", { name: "Save resume version" }));
    await waitFor(() => expect(bindings.CreateResumeVersion).toHaveBeenCalledWith({ jobId: job.id, selectedFactIds: [fact.id], templateId: template.id }));
    expect(await screen.findByText("Resume version 1 saved from 1 selected facts.")).toBeTruthy();
  });

  it("runs and accepts a recorded Ollama proposal through bindings", async () => {
    const proposal = { targetFactId: fact.id, supportingFactIds: [fact.id], text: "Built reliable Go deployment services for secure production workloads." };
    const run = {
      id: "run-1", jobId: job.id, provider: "ollama", model: "recorded", promptVersion: "prompt-v1", schemaVersion: "schema-v1",
      selectedFactIds: [fact.id], validationPassed: true, failureCategory: "", validationErrors: [], proposals: [proposal], resumeVersionId: "",
      createdAt: "2026-01-01T00:00:00Z", acceptedAt: "",
    };
    bindings.CheckOllama.mockResolvedValue({ provider: "ollama", endpoint: "http://127.0.0.1:11434", available: true, models: ["recorded"], message: "Ollama is available." });
    bindings.GenerateAITailoring.mockResolvedValue(run);
    bindings.AcceptAITailoring.mockResolvedValue({ application: { id: "application-1", jobId: job.id, status: "draft", selectedFactIds: [fact.id], versions: [version], createdAt: "", updatedAt: "" }, version });

    const user = await analyzeJobThroughUI();
    await user.click(screen.getByRole("button", { name: /^AI assistant/ }));
    await user.click(screen.getByRole("button", { name: "Check connection" }));
    await waitFor(() => expect(bindings.CheckOllama).toHaveBeenCalledWith("http://127.0.0.1:11434"));
    await user.click(screen.getByRole("button", { name: "Generate proposals" }));
    await screen.findByText("Original evidence and proposed wording");

    expect(bindings.GenerateAITailoring).toHaveBeenCalledWith({ jobId: job.id, selectedFactIds: [fact.id], provider: "ollama", model: "recorded", endpoint: "http://127.0.0.1:11434" });
    await user.click(screen.getByRole("button", { name: "Accept new resume version" }));
    await waitFor(() => expect(bindings.AcceptAITailoring).toHaveBeenCalledOnce());
    expect(bindings.AcceptAITailoring.mock.calls[0][0]).toMatchObject({ runId: run.id, templateId: template.id, proposals: [proposal] });
  });
});

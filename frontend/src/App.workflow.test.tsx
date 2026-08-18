// @vitest-environment jsdom

import { cleanup, render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import App from "./App";

const bindings = vi.hoisted(() => ({
  AcceptAITailoring: vi.fn(), AnalyzeJobDescription: vi.fn(), CancelAITailoring: vi.fn(), CheckGemini: vi.fn(), CheckOllama: vi.fn(),
  CompileLatex: vi.fn(), CompileResumeVersion: vi.fn(), CreateResumeVersion: vi.fn(), DeleteAchievement: vi.fn(), DeleteCertification: vi.fn(),
  DeleteEducation: vi.fn(), DeleteExperience: vi.fn(), DeleteGeminiAPIKey: vi.fn(), DeleteJob: vi.fn(), DeleteProject: vi.fn(), DeleteResumeTemplate: vi.fn(),
  ExportCompiledPDF: vi.fn(), ExportLatexSource: vi.fn(), ExportProfileBackup: vi.fn(), GenerateAITailoring: vi.fn(), GenerateProjectReadmeBullets: vi.fn(), GetAISettings: vi.fn(),
  GetGeminiCredentialStatus: vi.fn(), GetProfile: vi.fn(), GetSelectedResumeTemplateID: vi.fn(), ImportGitHubProjects: vi.fn(), ImportProfileBackup: vi.fn(),
  ImportResumeTemplate: vi.fn(), ListAchievements: vi.fn(), ListAIRuns: vi.fn(), ListApplications: vi.fn(), ListCertifications: vi.fn(), ListEducations: vi.fn(),
  ListExperiences: vi.fn(), ListJobs: vi.fn(), ListProjects: vi.fn(), ListResumeTemplates: vi.fn(), OpenResumeVersion: vi.fn(), RenderResumeTemplate: vi.fn(), SaveAchievement: vi.fn(),
  SaveAISettings: vi.fn(), SaveCertification: vi.fn(), SaveEducation: vi.fn(), SaveExperience: vi.fn(), SaveGeminiAPIKey: vi.fn(), SaveProfile: vi.fn(),
  SaveProject: vi.fn(), SaveResumeTemplate: vi.fn(), SaveResumeVersionEdit: vi.fn(), SelectResumeTemplate: vi.fn(), UpdateApplicationStatus: vi.fn(),
}));

vi.mock("../wailsjs/go/main/App", () => bindings);

if (!Range.prototype.getClientRects) {
  Range.prototype.getClientRects = () => [] as unknown as DOMRectList;
}

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

afterEach(() => {
  cleanup();
  vi.restoreAllMocks();
});

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

  it("adds editable, unsaved project bullets from a GitHub README", async () => {
    const project = {
      id: "project-1", name: "TailorCV", role: "Creator", description: "", url: "", repositoryUrl: "https://github.com/example/tailorcv", repositoryId: 1,
      repositoryReadme: "TailorCV is a Go and React desktop app for evidence-backed resumes.", repositoryVisibility: "public", repositoryUpdatedAt: "2026-01-01T00:00:00Z",
      startDate: "", endDate: "", ongoing: false, provenance: "github", verification: "verified", resumeEligible: true, skills: ["Go", "React"], detectedLanguages: [], bullets: [], createdAt: "", updatedAt: "",
    };
    bindings.ListProjects.mockResolvedValue([project]);
    bindings.CheckOllama.mockResolvedValue({ provider: "ollama", endpoint: "http://127.0.0.1:11434", available: true, models: ["recorded"], message: "Ollama is available." });
    bindings.GenerateProjectReadmeBullets.mockResolvedValue({ bullets: ["Built a Go and React desktop app for evidence-backed resumes."] });

    const user = userEvent.setup();
    render(<App />);
    await user.click(await screen.findByRole("button", { name: /^AI assistant/ }));
    await user.click(screen.getByRole("button", { name: "Check connection" }));
    await user.click(screen.getByRole("button", { name: "Projects" }));
    await user.click(screen.getByRole("button", { name: "Manage evidence" }));
    await user.click(await screen.findByRole("button", { name: "Generate editable bullets from README" }));

    await waitFor(() => expect(bindings.GenerateProjectReadmeBullets).toHaveBeenCalledWith({ projectId: project.id, provider: "ollama", model: "recorded", endpoint: "http://127.0.0.1:11434" }));
    expect(await screen.findByText("1 editable README bullet added. Review and save the project to keep them.")).toBeTruthy();
    expect(screen.getByDisplayValue("Built a Go and React desktop app for evidence-backed resumes.")).toBeTruthy();
  });

  it("compiles, reopens, and exports a saved resume artifact through bindings", async () => {
    const application = {
      id: "application-1", jobId: job.id, status: "draft", selectedFactIds: [fact.id], versions: [version],
      createdAt: "2026-01-01T00:00:00Z", updatedAt: "2026-01-01T00:00:00Z",
    };
    const compiledVersion = {
      ...version, compileSuccess: true, compileEngine: "Recorded Tectonic", compileDurationMs: 12,
      compiledAt: "2026-01-01T00:01:00Z", pdfAvailable: true,
    };
    const compileResult = {
      success: true, pdfBase64: "JVBERi0xLjcKJSVFT0YK", engine: "Recorded Tectonic", durationMs: 12, log: "", diagnostics: [],
    };
    bindings.ListJobs.mockResolvedValue([job]);
    bindings.ListApplications.mockResolvedValue([application]);
    bindings.OpenResumeVersion
      .mockResolvedValueOnce({ version, compileResult: { success: false, pdfBase64: "", engine: "", durationMs: 0, log: "", diagnostics: [] } })
      .mockResolvedValueOnce({ version: compiledVersion, compileResult });
    bindings.CompileResumeVersion.mockResolvedValue(compileResult);
    bindings.ExportCompiledPDF.mockResolvedValue({ path: "/tmp/staff-platform-resume.pdf", cancelled: false });

    const user = userEvent.setup();
    render(<App />);
    await user.click(await screen.findByRole("button", { name: "Job match" }));
    await user.click(await screen.findByRole("button", { name: /Senior Platform Engineer Fictional Cloud/ }));
    await user.click(screen.getByRole("button", { name: "Open source" }));
    await waitFor(() => expect(bindings.OpenResumeVersion).toHaveBeenCalledTimes(1));

    await user.click(await screen.findByRole("button", { name: "Compile PDF" }));
    await waitFor(() => expect(bindings.CompileResumeVersion).toHaveBeenCalledWith(version.id));

    await user.click(screen.getByRole("button", { name: /^Job match/ }));
    await user.click(screen.getByRole("button", { name: "Open source" }));
    await waitFor(() => expect(bindings.OpenResumeVersion).toHaveBeenCalledTimes(2));
    expect(await screen.findByText("Opened immutable resume version 1 with its saved PDF.")).toBeTruthy();

    const exportButton = screen.getByRole("button", { name: "Export PDF" });
    expect((exportButton as HTMLButtonElement).disabled).toBe(false);
    await user.click(exportButton);
    await waitFor(() => expect(bindings.ExportCompiledPDF).toHaveBeenCalledOnce());
    expect(await screen.findByText("PDF exported to /tmp/staff-platform-resume.pdf")).toBeTruthy();
  });

  it("exports and restores a backup through the native-dialog bindings", async () => {
    const backupResult = {
      path: "/tmp/tailorcv-profile-backup.json", cancelled: false, experienceCount: 1, projectCount: 0, educationCount: 0,
      certificationCount: 0, achievementCount: 0, jobCount: 0, applicationCount: 0, resumeVersionCount: 0, templateCount: 0, aiRunCount: 0,
    };
    bindings.ExportProfileBackup.mockResolvedValue(backupResult);
    bindings.ImportProfileBackup.mockResolvedValue(backupResult);
    vi.spyOn(window, "confirm").mockReturnValue(true);

    const user = userEvent.setup();
    render(<App />);
    await user.click(await screen.findByRole("button", { name: "Backup & restore" }));
    await user.click(screen.getByRole("button", { name: "Choose destination" }));
    await waitFor(() => expect(bindings.ExportProfileBackup).toHaveBeenCalledOnce());
    expect(await screen.findByText("Backup exported to /tmp/tailorcv-profile-backup.json")).toBeTruthy();

    await user.click(screen.getByRole("button", { name: "Choose backup" }));
    await waitFor(() => expect(bindings.ImportProfileBackup).toHaveBeenCalledOnce());
    expect(await screen.findByText("Backup restored from /tmp/tailorcv-profile-backup.json")).toBeTruthy();
    expect(bindings.GetProfile).toHaveBeenCalledTimes(2);
  });

  it("creates the first local profile through onboarding", async () => {
    const empty = {
      name: "", headline: "", email: "", phone: "", location: "", website: "", githubUsername: "", linkedInUrl: "",
      summary: "", skills: [], contactLinks: [], updatedAt: "",
    };
    bindings.GetProfile.mockResolvedValue(empty);
    bindings.ListExperiences.mockResolvedValue([]);
    bindings.SaveProfile.mockImplementation(async (input) => ({ ...input, updatedAt: "2026-08-08T12:00:00Z" }));

    const user = userEvent.setup();
    render(<App />);
    expect(await screen.findByRole("heading", { name: "Start with facts you control." })).toBeTruthy();
    await user.type(screen.getByLabelText("Full name"), "Ada Lovelace");
    await user.type(screen.getByLabelText("Email"), "ada@example.com");
    await user.type(screen.getByLabelText("Professional headline"), "Platform Engineer");
    await user.type(screen.getByLabelText(/^Core skills/), "Go, TypeScript, Go");
    await user.click(screen.getByRole("button", { name: "Create local profile" }));

    await waitFor(() => expect(bindings.SaveProfile).toHaveBeenCalledOnce());
    expect(bindings.SaveProfile.mock.calls[0][0]).toMatchObject({
      name: "Ada Lovelace", email: "ada@example.com", headline: "Platform Engineer", skills: ["Go", "TypeScript"],
    });
    expect(await screen.findByText("Profile saved locally.")).toBeTruthy();
    expect(screen.queryByRole("heading", { name: "Start with facts you control." })).toBeNull();
  });

  it("keeps imported GitHub projects locked until review and save", async () => {
    const githubProject = {
      id: "project-42", name: "release-console", role: "", description: "Fictional release automation", url: "",
      repositoryUrl: "https://github.com/ada-example/release-console", repositoryId: 42, repositoryReadme: "# Release Console",
      repositoryVisibility: "public", repositoryUpdatedAt: "2026-08-01T12:00:00Z", startDate: "", endDate: "", ongoing: false,
      provenance: "github", verification: "unverified", resumeEligible: false, position: 0, skills: ["Go"],
      detectedLanguages: [{ name: "Go", bytes: 900 }, { name: "Shell", bytes: 100 }], bullets: [],
      createdAt: "2026-08-08T12:00:00Z", updatedAt: "2026-08-08T12:00:00Z",
    };
    const approvedProject = { ...githubProject, role: "Creator", verification: "verified", resumeEligible: true };
    bindings.GetProfile.mockResolvedValue({ ...profile, githubUsername: "ada-example" });
    bindings.ListProjects.mockResolvedValueOnce([]).mockResolvedValue([githubProject]);
    bindings.ImportGitHubProjects.mockResolvedValue({ fetched: 1, imported: 1, updated: 0, skipped: 0, languageFallbacks: 0, readmeFallbacks: 0 });
    bindings.SaveProject.mockImplementation(async (input) => ({ ...approvedProject, ...input, updatedAt: "2026-08-08T12:01:00Z" }));

    const user = userEvent.setup();
    render(<App />);
    await user.click(await screen.findByRole("button", { name: "Projects" }));
    await user.click(screen.getByRole("button", { name: "Sync GitHub" }));
    await waitFor(() => expect(bindings.ImportGitHubProjects).toHaveBeenCalledOnce());
    expect(await screen.findByText("Review required")).toBeTruthy();
    expect(screen.getByRole("button", { name: "Review release-console before selection" })).toBeTruthy();

    await user.click(screen.getByRole("button", { name: "Manage evidence" }));
    await user.click(screen.getByRole("button", { name: "Approve for resume" }));
    await user.type(screen.getByLabelText("Your role"), "Creator");
    await user.click(screen.getByRole("button", { name: "Save project" }));
    await waitFor(() => expect(bindings.SaveProject).toHaveBeenCalledOnce());
    expect(bindings.SaveProject.mock.calls[0][0]).toMatchObject({
      id: githubProject.id, repositoryId: 42, provenance: "github", verification: "verified", resumeEligible: true,
    });

    await user.click(screen.getByRole("button", { name: /^Select for resume/ }));
    const addButton = screen.getByRole("button", { name: "Add release-console to resume" });
    await user.click(addButton);
    expect(await screen.findByText("1 projects selected")).toBeTruthy();
  });

  it("supports keyboard navigation and exposes current-view, form, and evidence semantics", async () => {
    const user = await analyzeJobThroughUI();
    expect(screen.getByRole("button", { name: /^Job match/ }).getAttribute("aria-current")).toBe("page");
    expect(screen.getByLabelText("Job description")).toBeTruthy();

    const evidenceCheckbox = screen.getByRole("checkbox", { name: /Remove Platform Engineer · Example Systems evidence from resume/ });
    evidenceCheckbox.focus();
    await user.keyboard(" ");
    expect((evidenceCheckbox as HTMLInputElement).checked).toBe(false);
    expect(screen.getByRole("checkbox", { name: /Add Platform Engineer · Example Systems evidence to resume/ })).toBeTruthy();

    const projects = screen.getByRole("button", { name: "Projects" });
    projects.focus();
    await user.keyboard("{Enter}");
    expect(projects.getAttribute("aria-current")).toBe("page");
    expect(screen.getByRole("button", { name: /Select for resume/ }).getAttribute("aria-pressed")).toBe("true");
  });

  it("announces compiler diagnostics and moves keyboard focus to the source line", async () => {
    bindings.CompileLatex.mockResolvedValue({
      success: false, pdfBase64: "", engine: "Recorded Tectonic", durationMs: 8, log: "", diagnostics: [
        { line: 1, severity: "error", message: "Undefined control sequence" },
      ],
    });
    const user = userEvent.setup();
    render(<App />);
    const latexNavigation = await screen.findByRole("button", { name: "LaTeX source" });
    latexNavigation.focus();
    await user.keyboard("{Enter}");
    expect(screen.getByRole("region", { name: "LaTeX source" })).toBeTruthy();
    const sourceEditor = screen.getByRole("textbox", { name: "LaTeX source editor" });
    expect(screen.getByRole("complementary", { name: "Resume preview" })).toBeTruthy();

    await user.click(screen.getByRole("button", { name: "Compile PDF" }));
    const diagnostics = await screen.findByRole("region", { name: "Compiler diagnostics" });
    expect(diagnostics.getAttribute("aria-live")).toBe("polite");
    const diagnostic = screen.getByRole("button", { name: /error Line 1 Undefined control sequence/ });
    diagnostic.focus();
    await user.keyboard("{Enter}");
    await waitFor(() => expect(document.activeElement).toBe(sourceEditor));
  });

  it("operates PDF preview controls from the keyboard", async () => {
    bindings.CompileLatex.mockResolvedValue({
      success: true, pdfBase64: "JVBERi0xLjcKJSVFT0YK", engine: "Recorded Tectonic", durationMs: 8, log: "", diagnostics: [],
    });
    const user = userEvent.setup();
    render(<App />);
    await user.click(await screen.findByRole("button", { name: "LaTeX source" }));
    await user.click(screen.getByRole("button", { name: "Compile PDF" }));
    const controls = await screen.findByRole("group", { name: "PDF preview controls" });
    expect(controls.textContent).toContain("100%");
    const zoomIn = screen.getByRole("button", { name: "Zoom in" });
    zoomIn.focus();
    await user.keyboard("{Enter}");
    expect(controls.textContent).toContain("115%");
  });

  it("keeps focus on a native-dialog trigger when selection is cancelled", async () => {
    bindings.ExportProfileBackup.mockResolvedValue({ cancelled: true });
    const user = userEvent.setup();
    render(<App />);
    await user.click(await screen.findByRole("button", { name: "Backup & restore" }));
    const trigger = screen.getByRole("button", { name: "Choose destination" });
    trigger.focus();
    await user.keyboard("{Enter}");
    await waitFor(() => expect(bindings.ExportProfileBackup).toHaveBeenCalledOnce());
    expect(document.activeElement).toBe(trigger);
  });
});

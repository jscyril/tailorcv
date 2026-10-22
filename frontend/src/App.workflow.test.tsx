// @vitest-environment jsdom

import { cleanup, render, screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import App from "./App";

const bindings = vi.hoisted(() => ({
  AcceptAITailoring: vi.fn(), AnalyzeJobDescription: vi.fn(), CancelAITailoring: vi.fn(), CheckGemini: vi.fn(), CheckOllama: vi.fn(),
  CompileLatex: vi.fn(), CompileResumeVersion: vi.fn(), CreateResume: vi.fn(), CreateResumeVersion: vi.fn(), DeleteResume: vi.fn(), DuplicateResume: vi.fn(), RenameResume: vi.fn(), DeleteAchievement: vi.fn(), DeleteCertification: vi.fn(),
  DeleteEducation: vi.fn(), DeleteExperience: vi.fn(), DeleteGeminiAPIKey: vi.fn(), DeleteJob: vi.fn(), DeleteProject: vi.fn(), DeleteResumeTemplate: vi.fn(),
  ExportCompiledPDF: vi.fn(), ExportLatexSource: vi.fn(), ExportProfileBackup: vi.fn(), GenerateAITailoring: vi.fn(), GenerateProjectReadmeBullets: vi.fn(), GetAISettings: vi.fn(),
  GetGeminiCredentialStatus: vi.fn(), GetProfile: vi.fn(), GetSelectedResumeTemplateID: vi.fn(), ImportGitHubProjects: vi.fn(), ImportProfileBackup: vi.fn(),
  ImportResumeTemplate: vi.fn(), ListAchievements: vi.fn(), ListAIRuns: vi.fn(), ListApplications: vi.fn(), ListCertifications: vi.fn(), ListEducations: vi.fn(),
  ListExperiences: vi.fn(), ListJobs: vi.fn(), ListProjects: vi.fn(), ListResumeTemplates: vi.fn(), OpenResumeVersion: vi.fn(), RenderResumeTemplate: vi.fn(), SaveAchievement: vi.fn(),
  SaveAISettings: vi.fn(), SaveCertification: vi.fn(), SaveEducation: vi.fn(), SaveExperience: vi.fn(), SaveGeminiAPIKey: vi.fn(), SaveProfile: vi.fn(),
  SaveProject: vi.fn(), SaveResumeTemplate: vi.fn(), SaveResumeVersionEdit: vi.fn(), SelectResumeTemplate: vi.fn(), UpdateApplicationStatus: vi.fn(), UpdateApplicationDetails: vi.fn(), KeywordCoverage: vi.fn(), DraftCoverLetter: vi.fn(), GenerateCoverLetter: vi.fn(), SaveCoverLetter: vi.fn(), ExportCoverLetterPDF: vi.fn(),
  ListAutomaticBackups: vi.fn(), RestoreAutomaticBackup: vi.fn(), GetLatexDraft: vi.fn(), SaveLatexDraft: vi.fn(), ClearLatexDraft: vi.fn(),
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
  bindings.ListAutomaticBackups.mockResolvedValue([]);
  bindings.KeywordCoverage.mockResolvedValue({ required: [], preferred: [] });
  bindings.GetLatexDraft.mockResolvedValue({ source: "", versionId: "", savedAt: "" });
  bindings.SaveLatexDraft.mockResolvedValue(undefined);
  bindings.ClearLatexDraft.mockResolvedValue(undefined);
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
    await user.click(screen.getByRole("button", { name: /^Projects/ }));
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
    await waitFor(() => expect(bindings.ExportCompiledPDF).toHaveBeenCalledWith(`Ada Lovelace - ${job.company} - ${job.role}`));
    expect(await screen.findByText("PDF exported to /tmp/staff-platform-resume.pdf")).toBeTruthy();
  });

  it("lists resumes on Home with versions and opens, starts, renames, duplicates, and deletes them", async () => {
    const versionTwo = { ...version, id: "version-2", versionNumber: 2, createdAt: "2026-01-02T00:00:00Z" };
    const tailored = { id: "application-1", jobId: job.id, name: "", status: "submitted", selectedFactIds: [fact.id], versions: [versionTwo, version], createdAt: "2026-01-01T00:00:00Z", updatedAt: "2026-01-02T00:00:00Z" };
    const general = { id: "application-2", jobId: "", name: "General resume", status: "draft", selectedFactIds: [], versions: [{ ...version, id: "version-g1", applicationId: "application-2" }], createdAt: "2026-01-01T00:00:00Z", updatedAt: "2026-01-01T00:00:00Z" };
    const copy = { ...general, id: "application-3", name: "Copy of Senior Platform Engineer at Fictional Cloud", updatedAt: "2026-01-03T00:00:00Z" };
    bindings.ListJobs.mockResolvedValue([job]);
    bindings.ListApplications.mockResolvedValue([tailored, general]);
    bindings.OpenResumeVersion.mockImplementation(async (id: string) => ({ version: { ...version, id }, compileResult: { success: false, pdfBase64: "", engine: "", durationMs: 0, log: "", diagnostics: [] } }));
    bindings.RenameResume.mockResolvedValue({ ...general, name: "Backend resume" });
    bindings.DuplicateResume.mockResolvedValue(copy);
    bindings.DeleteResume.mockResolvedValue(undefined);
    const confirm = vi.spyOn(window, "confirm").mockReturnValue(true);

    const user = userEvent.setup();
    render(<App />);
    expect(screen.getByRole("button", { name: "Home" }).getAttribute("aria-current")).toBe("page");
    expect(await screen.findByText("Senior Platform Engineer at Fictional Cloud")).toBeTruthy();
    expect(screen.getByText(/General resume · version 1/)).toBeTruthy();

    // Open an older version from the version history.
    await user.click(screen.getByText("2 versions"));
    await user.click(screen.getByRole("button", { name: "Open version 1 of Senior Platform Engineer at Fictional Cloud" }));
    await waitFor(() => expect(bindings.OpenResumeVersion).toHaveBeenCalledWith(version.id));
    expect(screen.getByRole("button", { name: "LaTeX source" }).getAttribute("aria-current")).toBe("page");

    await user.click(screen.getByRole("button", { name: "Home" }));
    await user.click(screen.getByRole("button", { name: "Open latest version of Senior Platform Engineer at Fictional Cloud" }));
    await waitFor(() => expect(bindings.OpenResumeVersion).toHaveBeenLastCalledWith(versionTwo.id));

    // Rename inline.
    await user.click(screen.getByRole("button", { name: "Home" }));
    await user.click(screen.getByRole("button", { name: "Rename General resume" }));
    const nameInput = screen.getByLabelText("Resume name");
    await user.clear(nameInput);
    await user.type(nameInput, "Backend resume{Enter}");
    await waitFor(() => expect(bindings.RenameResume).toHaveBeenCalledWith({ applicationId: general.id, name: "Backend resume" }));
    expect(await screen.findByText("Backend resume")).toBeTruthy();

    // Duplicate refreshes the list.
    bindings.ListApplications.mockResolvedValue([copy, tailored, { ...general, name: "Backend resume" }]);
    await user.click(screen.getByRole("button", { name: "Duplicate Senior Platform Engineer at Fictional Cloud" }));
    await waitFor(() => expect(bindings.DuplicateResume).toHaveBeenCalledWith(tailored.id));
    expect(await screen.findByText(copy.name)).toBeTruthy();

    // Delete asks first and removes the row.
    await user.click(screen.getByRole("button", { name: `Delete ${copy.name}` }));
    expect(confirm).toHaveBeenCalledWith(expect.stringContaining(copy.name));
    await waitFor(() => expect(bindings.DeleteResume).toHaveBeenCalledWith(copy.id));
    await waitFor(() => expect(screen.queryByText(copy.name)).toBeNull());

    // Starting a tailored resume opens an empty job.
    await user.click(screen.getByRole("button", { name: "New tailored resume" }));
    expect(screen.getByRole("button", { name: /^Job match/ }).getAttribute("aria-current")).toBe("page");
    expect((screen.getByLabelText("Job description") as HTMLTextAreaElement).value).toBe("");
    confirm.mockRestore();
  });

  it("tracks applications on Home and filters them by search and status", async () => {
    const tailored = { id: "application-1", jobId: job.id, name: "", status: "submitted", followUpOn: "2020-01-01", selectedFactIds: [fact.id], versions: [version], createdAt: "2026-01-01T00:00:00Z", updatedAt: "2026-01-02T00:00:00Z" };
    const general = { id: "application-2", jobId: "", name: "General resume", status: "draft", selectedFactIds: [], versions: [{ ...version, id: "version-g1", applicationId: "application-2" }], createdAt: "2026-01-01T00:00:00Z", updatedAt: "2026-01-01T00:00:00Z" };
    bindings.ListJobs.mockResolvedValue([job]);
    bindings.ListApplications.mockResolvedValue([tailored, general]);
    // The backend fills in the applied date; saving details afterwards must keep it.
    bindings.UpdateApplicationStatus.mockImplementation(async ({ status }) => ({ ...tailored, status, appliedOn: "2026-10-21" }));
    bindings.UpdateApplicationDetails.mockImplementation(async (input) => ({ ...tailored, ...input }));

    const user = userEvent.setup();
    render(<App />);
    const title = "Senior Platform Engineer at Fictional Cloud";
    expect(await screen.findByText(title)).toBeTruthy();
    // A past follow-up date on an open application is flagged.
    expect(screen.getByText("Follow-up due")).toBeTruthy();

    // Search matches company names; the status filter narrows further.
    await user.type(screen.getByLabelText("Search resumes"), "fictional");
    expect(screen.queryByText("General resume")).toBeNull();
    await user.selectOptions(screen.getByLabelText("Filter by status"), "draft");
    expect(screen.getByText("No resumes match.", { exact: false })).toBeTruthy();
    await user.click(screen.getByRole("button", { name: "Clear filters" }));
    expect(screen.getByText("General resume")).toBeTruthy();

    // Move the application forward and save its tracking details.
    const tracker = screen.getByRole("form", { name: `Tracking for ${title}` });
    await user.click(screen.getAllByText("Track application")[0]);
    await user.click(within(tracker).getByRole("button", { name: "Interviewing" }));
    await waitFor(() => expect(bindings.UpdateApplicationStatus).toHaveBeenCalledWith({ applicationId: tailored.id, status: "interviewing" }));
    const save = within(tracker).getByRole("button", { name: "Save tracking" }) as HTMLButtonElement;
    expect(save.disabled).toBe(true);
    await user.type(within(tracker).getByLabelText("Job posting link"), "https://jobs.example.com/1");
    await user.type(within(tracker).getByLabelText("Notes"), "Recruiter: Sam");
    await user.click(save);
    await waitFor(() => expect(bindings.UpdateApplicationDetails).toHaveBeenCalledWith({ applicationId: tailored.id, appliedOn: "2026-10-21", followUpOn: "2020-01-01", postingUrl: "https://jobs.example.com/1", notes: "Recruiter: Sam" }));
    expect(await screen.findByText("Application tracking saved.")).toBeTruthy();
  });

  it("shows which job skills an opened tailored resume covers", async () => {
    const tailored = { id: "application-1", jobId: job.id, name: "", status: "draft", selectedFactIds: [fact.id], versions: [version], createdAt: "2026-01-01T00:00:00Z", updatedAt: "2026-01-02T00:00:00Z" };
    bindings.ListJobs.mockResolvedValue([job]);
    bindings.ListApplications.mockResolvedValue([tailored]);
    bindings.OpenResumeVersion.mockResolvedValue({ version, compileResult: { success: false, pdfBase64: "", engine: "", durationMs: 0, log: "", diagnostics: [] } });
    bindings.KeywordCoverage.mockResolvedValue({ required: [{ skill: "Go", present: true }, { skill: "Kubernetes", present: false }], preferred: [] });

    const user = userEvent.setup();
    render(<App />);
    await user.click(await screen.findByRole("button", { name: "Open latest version of Senior Platform Engineer at Fictional Cloud" }));
    const coverage = await screen.findByRole("region", { name: "Job keyword coverage" });
    expect(within(coverage).getByText("1 of 2 job skills in this resume")).toBeTruthy();
    expect(within(coverage).getByText("(missing)", { exact: false }).textContent).toContain("missing");
    expect(bindings.KeywordCoverage).toHaveBeenCalledWith(job.id, version.latexSource);
  });

  it("drafts, edits, saves, and exports a cover letter from Home", async () => {
    const tailored = { id: "application-1", jobId: job.id, name: "", status: "draft", selectedFactIds: [fact.id], versions: [version], createdAt: "2026-01-01T00:00:00Z", updatedAt: "2026-01-02T00:00:00Z" };
    bindings.ListJobs.mockResolvedValue([job]);
    bindings.ListApplications.mockResolvedValue([tailored]);
    bindings.DraftCoverLetter.mockResolvedValue("Dear Fictional Cloud hiring team,\n\n- Built reliable Go deployment services");
    bindings.SaveCoverLetter.mockImplementation(async ({ text }) => ({ ...tailored, coverLetter: text }));
    bindings.ExportCoverLetterPDF.mockResolvedValue({ path: "/tmp/letter.pdf", cancelled: false });
    const confirm = vi.spyOn(window, "confirm").mockReturnValue(true);

    const user = userEvent.setup();
    render(<App />);
    await user.click(await screen.findByText("Cover letter (none yet)"));
    const panel = screen.getByRole("region", { name: "Cover letter for Senior Platform Engineer at Fictional Cloud" });
    // AI drafting needs a model first.
    expect((within(panel).getByRole("button", { name: "Draft with AI" }) as HTMLButtonElement).disabled).toBe(true);

    await user.click(within(panel).getByRole("button", { name: "Draft from evidence" }));
    const letter = within(panel).getByLabelText("Cover letter text") as HTMLTextAreaElement;
    await waitFor(() => expect(letter.value).toContain("Built reliable Go deployment services"));
    expect(confirm).not.toHaveBeenCalled();
    await user.type(letter, " Thanks!");
    await user.click(within(panel).getByRole("button", { name: "Save letter" }));
    await waitFor(() => expect(bindings.SaveCoverLetter).toHaveBeenCalledWith({ applicationId: tailored.id, text: expect.stringContaining("Thanks!") }));
    expect(await screen.findByText("Cover letter saved.")).toBeTruthy();

    // Drafting again asks before replacing the saved letter.
    confirm.mockReturnValue(false);
    await user.click(within(panel).getByRole("button", { name: "Draft from evidence" }));
    expect(confirm).toHaveBeenCalledWith("Replace the current cover letter with a new draft?");
    expect(bindings.DraftCoverLetter).toHaveBeenCalledTimes(1);

    await user.click(within(panel).getByRole("button", { name: "Export PDF" }));
    await waitFor(() => expect(bindings.ExportCoverLetterPDF).toHaveBeenCalledWith(expect.stringContaining("Thanks!"), "Ada Lovelace - Fictional Cloud - Cover letter"));
    expect(await screen.findByText("Cover letter exported to /tmp/letter.pdf.")).toBeTruthy();
    confirm.mockRestore();
  });

  it("compares two saved versions of a resume on Home", async () => {
    const one = { ...version, latexSource: "\\section{Skills}\nGo\n\\end{document}" };
    const two = { ...version, id: "version-2", versionNumber: 2, latexSource: "\\section{Skills}\nGo, Rust\n\\end{document}" };
    bindings.ListJobs.mockResolvedValue([job]);
    bindings.ListApplications.mockResolvedValue([{ id: "application-1", jobId: job.id, name: "Platform", status: "draft", selectedFactIds: [], versions: [two, one], createdAt: "", updatedAt: "" }]);

    const user = userEvent.setup();
    render(<App />);
    await user.click(await screen.findByText("2 versions"));
    await user.click(screen.getByRole("button", { name: "Compare versions of Platform" }));
    const compare = screen.getByRole("region", { name: "Compare versions of Platform" });
    expect(compare.textContent).toContain("+1 −1 lines");
    expect(compare.querySelector(".version-diff .removed")?.textContent).toContain("Go");
    expect(compare.querySelector(".version-diff .added")?.textContent).toContain("Go, Rust");

    await user.selectOptions(screen.getByLabelText("From"), "version-2");
    expect(compare.textContent).toContain("No changes");
  });

  it("creates a general resume from the selected template and saves template edits as a new resume", async () => {
    const created = { id: "application-9", jobId: "", name: "General resume", status: "draft", selectedFactIds: [], versions: [], createdAt: "2026-01-01T00:00:00Z", updatedAt: "2026-01-01T00:00:00Z" };
    const createdVersion = { ...version, id: "version-9", applicationId: created.id, latexSource: "\\documentclass{article}\\begin{document}Preview\\end{document}" };
    bindings.CreateResume.mockResolvedValue({ application: created, version: createdVersion });
    bindings.ListApplications.mockResolvedValueOnce([]).mockResolvedValue([{ ...created, versions: [createdVersion] }]);
    bindings.OpenResumeVersion.mockResolvedValue({ version: createdVersion, compileResult: { success: false, pdfBase64: "", engine: "", durationMs: 0, log: "", diagnostics: [] } });

    const user = userEvent.setup();
    render(<App />);
    expect(await screen.findByText(/No resumes yet/)).toBeTruthy();
    await user.click(screen.getAllByRole("button", { name: "New general resume" })[0]);
    await waitFor(() => expect(bindings.CreateResume).toHaveBeenCalledWith({ name: "General resume", templateId: template.id, latexSource: "\\documentclass{article}\\begin{document}Preview\\end{document}" }));
    await waitFor(() => expect(bindings.OpenResumeVersion).toHaveBeenCalledWith(createdVersion.id));
    expect(await screen.findByText('Saved "General resume" as version 1.')).toBeTruthy();
    expect(screen.getByRole("button", { name: "Version saved" })).toBeTruthy();

    // A template loaded without an open version offers "Save as new resume".
    bindings.SelectResumeTemplate.mockResolvedValue(template);
    await user.click(screen.getByRole("button", { name: "Reload data" }));
    const saveAsNew = await screen.findByRole("button", { name: "Save as new resume" });
    await user.click(saveAsNew);
    await waitFor(() => expect(bindings.CreateResume).toHaveBeenLastCalledWith(expect.objectContaining({ name: "Jake style resume", templateId: template.id })));
  });

  it("recovers a saved LaTeX draft, keeps saving it, and asks before discarding it", async () => {
    const draftSource = "\\documentclass{article}\\begin{document}Recovered draft\\end{document}";
    bindings.GetLatexDraft.mockResolvedValue({ source: draftSource, versionId: "", savedAt: "2026-10-21T10:00:00Z" });
    bindings.ListApplications.mockResolvedValue([{ id: "application-1", jobId: "", name: "Backend", status: "draft", selectedFactIds: [], versions: [version], createdAt: "", updatedAt: "" }]);
    bindings.CompileLatex.mockResolvedValue({ success: false, pdfBase64: "", engine: "", durationMs: 0, log: "", diagnostics: [] });
    const confirm = vi.spyOn(window, "confirm").mockReturnValue(false);

    const user = userEvent.setup();
    render(<App />);
    expect(await screen.findByText(/Unsaved LaTeX edits from/)).toBeTruthy();
    await new Promise((resolve) => setTimeout(resolve, 1700));
    expect(bindings.SaveLatexDraft).not.toHaveBeenCalled();
    expect(bindings.ClearLatexDraft).not.toHaveBeenCalled();

    await user.click(screen.getByRole("button", { name: "Restore edits" }));
    expect(screen.getByRole("textbox", { name: "LaTeX source editor" }).textContent).toContain("Recovered draft");
    await waitFor(() => expect(bindings.SaveLatexDraft).toHaveBeenCalledWith(draftSource, ""), { timeout: 3000 });

    await user.click(screen.getByRole("button", { name: "Home" }));
    await user.click(await screen.findByRole("button", { name: "Open latest version of Backend" }));
    expect(confirm).toHaveBeenCalledWith("You have unsaved LaTeX changes. Discard them?");
    expect(bindings.OpenResumeVersion).not.toHaveBeenCalled();
  });

  it("discards a recovered LaTeX draft", async () => {
    bindings.GetLatexDraft.mockResolvedValue({ source: "old draft", versionId: "", savedAt: "2026-10-21T10:00:00Z" });
    const user = userEvent.setup();
    render(<App />);
    await user.click(await screen.findByRole("button", { name: "Discard" }));
    await waitFor(() => expect(bindings.ClearLatexDraft).toHaveBeenCalled());
    expect(screen.queryByText(/Unsaved LaTeX edits from/)).toBeNull();
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

  it("lists automatic backups and restores one after confirmation", async () => {
    const daily = { name: "tailorcv-daily-2026-10-20.json", kind: "daily", createdAt: "2026-10-20T09:00:00Z", size: 4096 };
    bindings.ListAutomaticBackups.mockResolvedValue([daily]);
    bindings.RestoreAutomaticBackup.mockResolvedValue({ path: "/backups/" + daily.name, cancelled: false, certificationCount: 0, achievementCount: 0, resumeVersionCount: 0 });
    const confirm = vi.spyOn(window, "confirm").mockReturnValueOnce(false).mockReturnValue(true);

    const user = userEvent.setup();
    render(<App />);
    await user.click(await screen.findByRole("button", { name: "Backup & restore" }));
    const restore = await screen.findByRole("button", { name: /^Restore Daily backup/ });
    expect(screen.getByText("4 KB")).toBeTruthy();

    await user.click(restore);
    expect(bindings.RestoreAutomaticBackup).not.toHaveBeenCalled();
    await user.click(restore);
    await waitFor(() => expect(bindings.RestoreAutomaticBackup).toHaveBeenCalledWith(daily.name));
    expect(await screen.findByText(/^Restored the backup from/)).toBeTruthy();
    expect(confirm.mock.calls[1][0]).toContain("safety copy");
    await waitFor(() => expect(bindings.ListAutomaticBackups).toHaveBeenCalledTimes(2));
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

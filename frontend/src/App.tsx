import { FormEvent, useEffect, useMemo, useRef, useState } from "react";
import { AcceptAITailoring, AnalyzeJobDescription, ApplyResumeImport, CancelAITailoring, CheckGemini, CheckOllama, ClearLatexDraft, CompileLatex, CompileResumeVersion, CreateResume, CreateResumeVersion, DeleteAchievement, DeleteCertification, DeleteEducation, DeleteExperience, DeleteGeminiAPIKey, DeleteJob, DeleteProject, DeleteResume, DeleteResumeTemplate, DraftCoverLetter, DuplicateResume, ExportCompiledPDF, ExportCoverLetterPDF, ExportLatexSource, ExportProfileBackup, GenerateAITailoring, GenerateCoverLetter, GenerateProjectReadmeBullets, GetAISettings, GetGeminiCredentialStatus, GetLatexDraft, GetProfile, GetSelectedResumeTemplateID, ImportGitHubProject, ImportGitHubProjects, ImportProfileBackup, ImportResumeTemplate, KeywordCoverage, ListAIRuns, ListAchievements, ListApplications, ListCertifications, ListEducations, ListExperiences, ListJobs, ListProjects, ListResumeTemplates, OpenResumeVersion, PreviewResumeImport, RenameResume, RenderResumeTemplate, RestoreAutomaticBackup, SaveAISettings, SaveAchievement, SaveCertification, SaveCoverLetter, SaveEducation, SaveExperience, SaveGeminiAPIKey, SaveLatexDraft, SaveProfile, SaveProject, SaveResumeTemplate, SaveResumeVersionEdit, SelectResumeTemplate, UpdateApplicationDetails, UpdateApplicationStatus } from "../wailsjs/go/main/App";
import { domain } from "../wailsjs/go/models";
import { Achievement, AchievementDraft, Certification, CertificationDraft, achievementToDraft, certificationToDraft, toAchievementInput, toCertificationInput } from "./lib/credentials";
import { Education, EducationDraft, educationToDraft, newEducationDraft, toEducationInput } from "./lib/education";
import { errorMessage } from "./lib/errors";
import { Experience, ExperienceDraft, experienceToDraft, newEvidenceBullet, newExperienceDraft, toExperienceInput } from "./lib/experience";
import { Profile, emptyProfile, parseSkills, profileCompletion } from "./lib/profile";
import { Project, ProjectDraft, isProjectSelectable, newProjectDraft, projectToDraft, reconcileSelectedProjectKeys, removeSelectedProjectKey, toProjectInput } from "./lib/project";
import { type Application, type ResumeVersion, type View } from "./lib/resume";
import { NavButton, PanelHeader, TopToolbar } from "./components/ui";
import { type AIProposal, type AIProvider, type AIProviderStatus, type AIRun, AIWorkspace, type CredentialStatus, type Job, type JobAnalysis } from "./features/ai/AIWorkspace";
import { type ApplicationStatus } from "./features/applications/ApplicationStatusControl";
import { DataWorkspace } from "./features/data/DataWorkspace";
import { LatexWorkspace } from "./features/editor/LatexWorkspace";
import { ExperienceSection } from "./features/experience/ExperienceSection";
import { WorkspaceOverview, type CoverLetterHandlers } from "./features/home/Home";
import { type ApplicationDetails } from "./features/applications/ApplicationTracker";
import { JobTailor } from "./features/jobs/JobTailor";
import { Onboarding } from "./features/onboarding/Onboarding";
import { ResumeImportDialog } from "./features/onboarding/ResumeImportDialog";
import { ResumePreview } from "./features/preview/ResumePreview";
import { CredentialsWorkspace, EducationWorkspace, ProfileWorkspace, SkillsWorkspace } from "./features/profile/ProfileWorkspaces";
import { ProjectWorkspace } from "./features/projects/ProjectWorkspace";
import { TemplatesWorkspace } from "./features/templates/TemplatesWorkspace";

const EMPTY_JOB: Job = { id: "", company: "", role: "", description: "", createdAt: "", updatedAt: "" };

const DEFAULT_LATEX = String.raw`\documentclass[10pt]{article}
\usepackage[margin=0.55in]{geometry}
\usepackage[hidelinks]{hyperref}
\setlength{\parindent}{0pt}

\begin{document}
\begin{center}
  {\LARGE \textbf{Your Name}}\\
  Backend Engineer $\cdot$ your@email.com $\cdot$ github.com/you
\end{center}

\section*{Experience}
% TailorCV inserts approved, evidence-backed bullets here.

\section*{Projects}
% Selected projects are rendered here.

\section*{Education}
% Saved education records are rendered here.

\section*{Skills}
% Skills from your career profile are rendered here.
\end{document}`;

function profileHasData(profile: Profile): boolean {
  return Boolean(profile.name || profile.email || profile.headline || profile.phone || profile.location || profile.website || profile.githubUsername || profile.linkedInUrl || profile.summary || profile.skills.length);
}

export default function App() {
  const [view, setView] = useState<View>("overview");
  const [profile, setProfile] = useState<Profile>(emptyProfile);
  const [experiences, setExperiences] = useState<ExperienceDraft[]>([]);
  const [projects, setProjects] = useState<ProjectDraft[]>([]);
  const [educations, setEducations] = useState<EducationDraft[]>([]);
  const [certifications,setCertifications]=useState<CertificationDraft[]>([]);
  const [achievements,setAchievements]=useState<AchievementDraft[]>([]);
  const [skillsText, setSkillsText] = useState("");
  const [jobDraft, setJobDraft] = useState<Job>(EMPTY_JOB);
  const [jobs, setJobs] = useState<Job[]>([]);
  const [applications, setApplications] = useState<Application[]>([]);
  const [applicationStatusBusy, setApplicationStatusBusy] = useState(false);
  const [selectedFactIDs, setSelectedFactIDs] = useState<string[]>([]);
  const [analysis, setAnalysis] = useState<JobAnalysis | null>(null);
  const [busy, setBusy] = useState(true);
  const [experienceBusyKey, setExperienceBusyKey] = useState("");
  const [projectBusyKey, setProjectBusyKey] = useState("");
  const [projectReadmeBusyKey, setProjectReadmeBusyKey] = useState("");
  const [educationBusyKey, setEducationBusyKey] = useState("");
  const [credentialBusyKey,setCredentialBusyKey]=useState("");
  const [backupBusy, setBackupBusy] = useState<"export" | "import" | "">("");
  const [githubBusy, setGitHubBusy] = useState(false);
  const [lastBackupResult, setLastBackupResult] = useState<domain.BackupResult | null>(null);
  const [onboardingPending, setOnboardingPending] = useState(false);
  const [message, setMessage] = useState("");
  const [error, setError] = useState("");
  const [selectedProjectKeys, setSelectedProjectKeys] = useState<string[]>([]);
  const [latexSource, setLatexSource] = useState(DEFAULT_LATEX);
  const [templates, setTemplates] = useState<domain.ResumeTemplate[]>([]);
  const [selectedTemplateID, setSelectedTemplateID] = useState("");
  const [templateBusyKey, setTemplateBusyKey] = useState("");
  const [compileBusy, setCompileBusy] = useState(false);
  const [versionBusy, setVersionBusy] = useState(false);
  const [openVersionID, setOpenVersionID] = useState("");
  // The editor text as last loaded or saved; anything else is unsaved.
  const [latexBaseline, setLatexBaseline] = useState("");
  const [recoveredDraft, setRecoveredDraft] = useState<domain.LatexDraft | null>(null);
  const [compileResult, setCompileResult] = useState<domain.CompileResult | null>(null);
  const compileRequestRef = useRef(0);
  const compileTimerRef = useRef<number | undefined>(undefined);
  const skipNextAutomaticCompileRef = useRef(false);
  const [ollamaEndpoint, setOllamaEndpoint] = useState("http://127.0.0.1:11434");
	const [aiProvider, setAIProvider] = useState<AIProvider>("ollama");
  const [aiProviderStatus, setAIProviderStatus] = useState<AIProviderStatus | null>(null);
	const [ollamaModel, setOllamaModel] = useState("");
	const [geminiModel, setGeminiModel] = useState("");
	const [geminiAPIKey, setGeminiAPIKey] = useState("");
	const [geminiCredential, setGeminiCredential] = useState<CredentialStatus>({ configured: false, message: "Gemini API key is not configured." });
  const [aiRun, setAIRun] = useState<AIRun | null>(null);
  const [aiRuns, setAIRuns] = useState<AIRun[]>([]);
  const [reviewedProposals, setReviewedProposals] = useState<AIProposal[]>([]);
  const [acceptedProposalIDs, setAcceptedProposalIDs] = useState<string[]>([]);
  const [aiBusy, setAIBusy] = useState<"check" | "credential" | "generate" | "accept" | "">("");

  const loadWorkspaceData = async () => {
    const [result, savedExperiences, savedProjects, savedEducations,savedCertifications,savedAchievements, savedTemplates, savedTemplateID, savedJobs, savedApplications, savedAIRuns, savedAISettings, credentialStatus] = await Promise.all([GetProfile(), ListExperiences(), ListProjects(), ListEducations(),ListCertifications(),ListAchievements(), ListResumeTemplates(), GetSelectedResumeTemplateID(), ListJobs(), ListApplications(), ListAIRuns(), GetAISettings(), GetGeminiCredentialStatus().catch((reason) => ({ configured: false, message: errorMessage(reason) }))]);
    const loaded = { ...emptyProfile, ...result } as Profile;
    loaded.skills ??= [];
    loaded.contactLinks ??= [];
    setProfile(loaded);
    setSkillsText(loaded.skills.join(", "));
    setExperiences((savedExperiences as unknown as Experience[]).map(experienceToDraft));
    const projectDrafts = (savedProjects as unknown as Project[]).map(projectToDraft);
    setProjects(projectDrafts);
    const selectedKeys = projectDrafts.filter(isProjectSelectable).slice(0, 3).map((project) => project.key);
    setSelectedProjectKeys(selectedKeys);
    setEducations((savedEducations as unknown as Education[]).map(educationToDraft));
    setCertifications((savedCertifications as unknown as Certification[]).map(certificationToDraft));setAchievements((savedAchievements as unknown as Achievement[]).map(achievementToDraft));
    setTemplates(savedTemplates);
    setSelectedTemplateID(savedTemplateID);
    setJobs(savedJobs as unknown as Job[]);
    setApplications(savedApplications as unknown as Application[]);
    setAIRuns(savedAIRuns as unknown as AIRun[]);
		setOnboardingPending(!profileHasData(loaded) && savedExperiences.length === 0 && savedProjects.length === 0 && savedEducations.length === 0);
		setAIProvider(savedAISettings.provider as AIProvider);
		setOllamaEndpoint(savedAISettings.ollamaEndpoint);
		setOllamaModel(savedAISettings.ollamaModel);
		setGeminiModel(savedAISettings.geminiModel);
		setGeminiCredential(credentialStatus as CredentialStatus);
		setAIProviderStatus(null);
    setJobDraft(EMPTY_JOB);
    setAnalysis(null);
    setSelectedFactIDs([]);
    const rendered = await RenderResumeTemplate(savedTemplateID, selectedKeys.filter((key) => !key.startsWith("new-project-")));
    setLatexSource(rendered);
    setOpenVersionID("");
    setLatexBaseline(rendered);
    setCompileResult(null);
  };

  const latexDirty = latexSource !== latexBaseline;
  const confirmDiscardLatex = () => !latexDirty || window.confirm("You have unsaved LaTeX changes. Discard them?");

  useEffect(() => {
    GetLatexDraft().then((draft) => { if (draft.source) setRecoveredDraft(draft); }).catch((reason) => setError(errorMessage(reason)));
  }, []);

  // Keep unsaved edits in the database so a crash or close does not lose them.
  // Waits for startup and for the user's decision on a recovered draft.
  useEffect(() => {
    if (busy || recoveredDraft) return;
    const timer = window.setTimeout(() => {
      (latexDirty ? SaveLatexDraft(latexSource, openVersionID) : ClearLatexDraft()).catch((reason) => console.error("LaTeX draft:", reason));
    }, 1500);
    return () => window.clearTimeout(timer);
  }, [busy, recoveredDraft, latexDirty, latexSource, openVersionID]);

  useEffect(() => {
    loadWorkspaceData().catch((reason) => setError(errorMessage(reason))).finally(() => setBusy(false));
  }, []);

  const completion = useMemo(() => profileCompletion(profile), [profile]);

  const updateProfile = <K extends keyof Profile>(field: K, value: Profile[K]) => {
    setProfile((current) => ({ ...current, [field]: value }));
    setMessage("");
  };

  const saveProfile = async (event: FormEvent) => {
    event.preventDefault();
    setBusy(true);
    setError("");
    setMessage("");
    try {
      const saved = await SaveProfile(new domain.ProfileInput({
        ...profile,
        skills: parseSkills(skillsText),
        contactLinks: profile.contactLinks,
      }));
      const normalized = { ...emptyProfile, ...saved } as Profile;
      setProfile(normalized);
      setSkillsText(normalized.skills.join(", "));
      setOnboardingPending(false);
      setMessage("Profile saved locally.");
    } catch (reason) {
      setError(errorMessage(reason));
    } finally {
      setBusy(false);
    }
  };

  const analyzeJob = async (event: FormEvent) => {
    event.preventDefault();
    setBusy(true);
    setError("");
    setAnalysis(null);
    try {
      const result = await AnalyzeJobDescription({ id: jobDraft.id, company: jobDraft.company, role: jobDraft.role, description: jobDraft.description });
      const next = result as unknown as JobAnalysis;
      setAnalysis(next);
      setJobDraft(next.job);
      setSelectedFactIDs(next.rankedEvidence.filter((evidence) => evidence.selectable).slice(0, 6).map((evidence) => evidence.factId));
      setAIRun(null);
      setReviewedProposals([]);
      setAcceptedProposalIDs([]);
      setJobs((await ListJobs()) as unknown as Job[]);
      setMessage("Job saved and evidence ranked locally.");
    } catch (reason) {
      setError(errorMessage(reason));
    } finally {
      setBusy(false);
    }
  };

  const updateJobDraft = (field: "company" | "role" | "description", value: string) => {
    setJobDraft((current) => ({ ...current, [field]: value }));
    setAnalysis(null);
    setSelectedFactIDs([]);
    setAIRun(null);
    setMessage("");
  };

  const newJob = () => {
    setJobDraft(EMPTY_JOB);
    setAnalysis(null);
    setSelectedFactIDs([]);
    setAIRun(null);
    setMessage("");
  };

  const openJob = (job: Job) => {
    setJobDraft(job);
    setAnalysis(null);
    setSelectedFactIDs(applications.find((application) => application.jobId === job.id)?.selectedFactIds ?? []);
    setAIRun(null);
    setMessage("");
  };

  const toggleEvidenceFact = (factID: string) => {
    setSelectedFactIDs((current) => current.includes(factID) ? current.filter((id) => id !== factID) : [...current, factID]);
    setAIRun(null);
    setReviewedProposals([]);
    setAcceptedProposalIDs([]);
    setMessage("");
  };

  const createResumeVersion = async () => {
    if (!confirmDiscardLatex()) return;
    setVersionBusy(true);
    setError("");
    setMessage("");
    try {
      const result = await CreateResumeVersion({ jobId: jobDraft.id, selectedFactIds: selectedFactIDs, templateId: selectedTemplateID });
      const version = result.version as unknown as ResumeVersion;
      setApplications((await ListApplications()) as unknown as Application[]);
      setLatexSource(version.latexSource);
      setOpenVersionID(version.id);
      setLatexBaseline(version.latexSource);
      setCompileResult(null);
      setMessage(`Resume version ${version.versionNumber} saved from ${selectedFactIDs.length} selected facts.`);
    } catch (reason) {
      setError(errorMessage(reason));
    } finally {
      setVersionBusy(false);
    }
  };

  const updateApplicationStatus = async (status: ApplicationStatus, target = currentApplication) => {
    if (!target || target.status === status) return;
    setApplicationStatusBusy(true);
    setError("");
    setMessage("");
    try {
      const updated = await UpdateApplicationStatus({ applicationId: target.id, status });
      setApplications((current) => current.map((application) => application.id === target.id ? updated as unknown as Application : application));
      setMessage(`Application marked ${status}.`);
    } catch (reason) {
      setError(errorMessage(reason));
    } finally {
      setApplicationStatusBusy(false);
    }
  };

  const [resumeImport, setResumeImport] = useState<domain.ResumeImport | null>(null);
  const [resumeImportBusy, setResumeImportBusy] = useState(false);
  const previewResumeImport = async () => {
    setError("");
    setMessage("");
    try {
      const preview = await PreviewResumeImport();
      if (preview.fileName) setResumeImport(preview);
    } catch (reason) {
      setError(errorMessage(reason));
    }
  };
  const applyResumeImport = async (selected: domain.ResumeImport) => {
    setResumeImportBusy(true);
    setError("");
    try {
      const result = await ApplyResumeImport(selected);
      await loadWorkspaceData();
      setResumeImport(null);
      setOnboardingPending(false);
      const plural = (count: number, word: string) => `${count} ${word}${count === 1 ? "" : "s"}`;
      setMessage(`Imported ${plural(result.experiences, "role")}, ${plural(result.projects, "project")}, ${result.educations} education ${result.educations === 1 ? "entry" : "entries"}, and ${plural(result.skills, "new skill")}.${result.duplicates ? ` Skipped ${result.duplicates} already in your profile.` : ""} Pick a template to build a resume from it.`);
      setView("experience");
    } catch (reason) {
      setError(errorMessage(reason));
    } finally {
      setResumeImportBusy(false);
    }
  };
  const [coverLetterBusyId, setCoverLetterBusyId] = useState("");
  const aiModel = aiProvider === "ollama" ? ollamaModel : geminiModel;
  // Runs one cover letter action, showing errors; returns null on failure.
  const coverLetterAction = async <T,>(application: Application, action: () => Promise<T>): Promise<T | null> => {
    setCoverLetterBusyId(application.id);
    setError("");
    setMessage("");
    try {
      return await action();
    } catch (reason) {
      setError(errorMessage(reason));
      return null;
    } finally {
      setCoverLetterBusyId("");
    }
  };
  const coverLetters: CoverLetterHandlers = {
    aiModel: aiModel ? `${aiProvider === "ollama" ? "Ollama" : "Gemini"} · ${aiModel}` : "",
    busyId: coverLetterBusyId,
    draft: (application) => coverLetterAction(application, () => DraftCoverLetter(application.id)),
    generate: (application) => coverLetterAction(application, () => GenerateCoverLetter({ applicationId: application.id, provider: aiProvider, model: aiModel, endpoint: aiProvider === "ollama" ? ollamaEndpoint : "" })),
    save: async (application, text) => {
      const updated = await coverLetterAction(application, () => SaveCoverLetter({ applicationId: application.id, text }));
      if (!updated) return false;
      setApplications((current) => current.map((item) => item.id === application.id ? updated as unknown as Application : item));
      setMessage("Cover letter saved.");
      return true;
    },
    export: async (application, text) => {
      const job = jobs.find((item) => item.id === application.jobId);
      const result = await coverLetterAction(application, () => ExportCoverLetterPDF(text, [profile.name, job?.company || job?.role, "Cover letter"].filter(Boolean).join(" - ")));
      if (result && !result.cancelled) setMessage(`Cover letter exported to ${result.path}.`);
    },
  };

  const saveApplicationDetails = async (target: Application, details: ApplicationDetails) => {
    setApplicationStatusBusy(true);
    setError("");
    setMessage("");
    try {
      const updated = await UpdateApplicationDetails({ applicationId: target.id, ...details });
      setApplications((current) => current.map((application) => application.id === target.id ? updated as unknown as Application : application));
      setMessage("Application tracking saved.");
      return true;
    } catch (reason) {
      setError(errorMessage(reason));
      return false;
    } finally {
      setApplicationStatusBusy(false);
    }
  };

  // checkUnsaved is false when the caller just saved the editor text itself.
  const openResumeVersion = async (version: ResumeVersion, checkUnsaved = true) => {
    if (checkUnsaved && !confirmDiscardLatex()) return;
    setError("");
    setCompileResult(null);
    try {
      const workspace = await OpenResumeVersion(version.id);
      const opened = workspace.version as unknown as ResumeVersion;
      setLatexSource(opened.latexSource);
      setOpenVersionID(opened.id);
      setLatexBaseline(opened.latexSource);
      setCompileResult(opened.compiledAt && (!opened.compileSuccess || opened.pdfAvailable) ? workspace.compileResult : null);
      skipNextAutomaticCompileRef.current = true;
      setView("latex");
      setMessage(opened.pdfAvailable
        ? `Opened immutable resume version ${opened.versionNumber} with its saved PDF.`
        : `Opened immutable resume version ${opened.versionNumber}.`);
    } catch (reason) {
      setError(errorMessage(reason));
    }
  };

  const saveResumeVersionEdit = async () => {
    const application = applications.find((item) => item.versions.some((version) => version.id === openVersionID));
    if (!application || !openVersionID) {
      setError("Open a saved resume version before saving an edited snapshot.");
      return;
    }
    setVersionBusy(true);
    setError("");
    try {
      const saved = await SaveResumeVersionEdit({ applicationId: application.id, baseVersionId: openVersionID, latexSource });
      const version = saved as unknown as ResumeVersion;
      setOpenVersionID(version.id);
      setLatexBaseline(version.latexSource);
      setApplications((await ListApplications()) as unknown as Application[]);
      setMessage(`Saved edited source as immutable version ${version.versionNumber}.`);
    } catch (reason) {
      setError(errorMessage(reason));
    } finally {
      setVersionBusy(false);
    }
  };

  const refreshApplications = async () => setApplications((await ListApplications()) as unknown as Application[]);

  const saveAsNewResume = async (name: string, source: string) => {
    setVersionBusy(true);
    setError("");
    try {
      const result = await CreateResume({ name, templateId: selectedTemplateID, latexSource: source });
      await refreshApplications();
      await openResumeVersion(result.version as unknown as ResumeVersion, false);
      setMessage(`Saved "${result.application.name}" as version 1.`);
    } catch (reason) {
      setError(errorMessage(reason));
    } finally {
      setVersionBusy(false);
    }
  };

  const createGeneralResume = async () => {
    if (!confirmDiscardLatex()) return;
    setError("");
    try {
      const rendered = await RenderResumeTemplate(selectedTemplateID, selectedSavedProjectIDs());
      await saveAsNewResume("General resume", rendered);
    } catch (reason) {
      setError(errorMessage(reason));
    }
  };

  const renameResume = async (application: Application, name: string) => {
    setError("");
    try {
      const renamed = await RenameResume({ applicationId: application.id, name });
      setApplications((current) => current.map((item) => item.id === application.id ? renamed as unknown as Application : item));
      setMessage(`Renamed to "${renamed.name}".`);
      return true;
    } catch (reason) {
      setError(errorMessage(reason));
      return false;
    }
  };

  const duplicateResume = async (application: Application) => {
    setError("");
    try {
      const copy = await DuplicateResume(application.id);
      await refreshApplications();
      setMessage(`Created "${copy.name}" from the latest version.`);
    } catch (reason) {
      setError(errorMessage(reason));
    }
  };

  const deleteResume = async (application: Application, title: string) => {
    if (!window.confirm(`Delete "${title}" and all ${application.versions.length} of its versions? This cannot be undone.`)) return;
    setError("");
    try {
      await DeleteResume(application.id);
      setApplications((current) => current.filter((item) => item.id !== application.id));
      if (application.versions.some((version) => version.id === openVersionID)) {
        setOpenVersionID("");
        setLatexBaseline("");
      }
      setMessage(`Deleted "${title}".`);
    } catch (reason) {
      setError(errorMessage(reason));
    }
  };

  const deleteJob = async (job: Job) => {
    if (!window.confirm(`Delete ${job.role || "this saved job"}${job.company ? ` at ${job.company}` : ""}?`)) return;
    setBusy(true);
    setError("");
    try {
      await DeleteJob(job.id);
      setJobs((current) => current.filter((item) => item.id !== job.id));
      setApplications((current) => current.filter((application) => application.jobId !== job.id));
      if (jobDraft.id === job.id) newJob();
      setMessage("Saved job deleted.");
    } catch (reason) {
      setError(errorMessage(reason));
    } finally {
      setBusy(false);
    }
  };

  const addExperience = () => {
    setExperiences((current) => [...current, newExperienceDraft()]);
    setMessage("");
  };

  const updateExperience = (key: string, next: ExperienceDraft) => {
    setExperiences((current) => current.map((experience) => experience.key === key ? next : experience));
    setMessage("");
  };

  const saveExperience = async (event: FormEvent, draft: ExperienceDraft) => {
    event.preventDefault();
    setExperienceBusyKey(draft.key);
    setError("");
    setMessage("");
    try {
      const saved = await SaveExperience(new domain.ExperienceInput(toExperienceInput(draft)));
      const normalized = experienceToDraft(saved as unknown as Experience);
      setExperiences((current) => current.map((experience) => experience.key === draft.key ? normalized : experience));
      setMessage("Experience saved locally.");
    } catch (reason) {
      setError(errorMessage(reason));
    } finally {
      setExperienceBusyKey("");
    }
  };

  const deleteExperience = async (draft: ExperienceDraft) => {
    if (draft.id && !window.confirm(`Delete ${draft.title || "this role"} at ${draft.company || "this company"}?`)) {
      return;
    }
    if (!draft.id) {
      setExperiences((current) => current.filter((experience) => experience.key !== draft.key));
      return;
    }
    setExperienceBusyKey(draft.key);
    setError("");
    try {
      await DeleteExperience(draft.id);
      setExperiences((current) => current.filter((experience) => experience.key !== draft.key));
      setMessage("Experience deleted.");
    } catch (reason) {
      setError(errorMessage(reason));
    } finally {
      setExperienceBusyKey("");
    }
  };

  const addProject = () => {
    setProjects((current) => [...current, newProjectDraft()]);
    setMessage("");
  };

  const updateProject = (key: string, next: ProjectDraft) => {
    setProjects((current) => current.map((project) => project.key === key ? next : project));
    if (!isProjectSelectable(next)) {
      setSelectedProjectKeys((current) => current.filter((item) => item !== key));
    }
    setMessage("");
  };

  const saveProject = async (event: FormEvent, draft: ProjectDraft) => {
    event.preventDefault();
    setProjectBusyKey(draft.key);
    setError("");
    setMessage("");
    try {
      const saved = await SaveProject(new domain.ProjectInput(toProjectInput(draft)));
      const normalized = projectToDraft(saved as unknown as Project);
      setProjects((current) => current.map((project) => project.key === draft.key ? normalized : project));
      setSelectedProjectKeys((current) => reconcileSelectedProjectKeys(current, draft.key, normalized.key, isProjectSelectable(normalized)));
      setMessage("Project saved locally.");
    } catch (reason) {
      setError(errorMessage(reason));
    } finally {
      setProjectBusyKey("");
    }
  };

  const deleteProject = async (draft: ProjectDraft) => {
    if (draft.id && !window.confirm(`Delete ${draft.name || "this project"}?`)) {
      return;
    }
    if (!draft.id) {
      setProjects((current) => current.filter((project) => project.key !== draft.key));
      setSelectedProjectKeys((current) => removeSelectedProjectKey(current, draft.key));
      return;
    }
    setProjectBusyKey(draft.key);
    setError("");
    try {
      await DeleteProject(draft.id);
      setProjects((current) => current.filter((project) => project.key !== draft.key));
      setSelectedProjectKeys((current) => removeSelectedProjectKey(current, draft.key));
      setMessage("Project deleted.");
    } catch (reason) {
      setError(errorMessage(reason));
    } finally {
      setProjectBusyKey("");
    }
  };

  const generateProjectReadmeBullets = async (draft: ProjectDraft) => {
    setProjectReadmeBusyKey(draft.key);
    setError("");
    setMessage("");
    try {
      const result = await GenerateProjectReadmeBullets({ projectId: draft.id, provider: aiProvider, model: aiProvider === "ollama" ? ollamaModel : geminiModel, endpoint: aiProvider === "ollama" ? ollamaEndpoint : "" }) as { bullets: string[] };
      setProjects((current) => current.map((project) => project.key !== draft.key ? project : { ...project, bullets: [...project.bullets, ...result.bullets.map((text) => ({ ...newEvidenceBullet(), text, provenance: "github" as const, sourceUrl: project.repositoryUrl }))] }));
      setMessage(`${result.bullets.length} editable README bullet${result.bullets.length === 1 ? "" : "s"} added. Review and save the project to keep them.`);
    } catch (reason) {
      setError(errorMessage(reason));
    } finally {
      setProjectReadmeBusyKey("");
    }
  };

  const addEducation = () => {
    setEducations((current) => [...current, newEducationDraft()]);
    setMessage("");
  };

  const updateEducation = (key: string, next: EducationDraft) => {
    setEducations((current) => current.map((education) => education.key === key ? next : education));
    setMessage("");
  };

  const saveEducation = async (event: FormEvent, draft: EducationDraft) => {
    event.preventDefault();
    setEducationBusyKey(draft.key);
    setError("");
    setMessage("");
    try {
      const saved = await SaveEducation(new domain.EducationInput(toEducationInput(draft)));
      const normalized = educationToDraft(saved as unknown as Education);
      setEducations((current) => current.map((education) => education.key === draft.key ? normalized : education));
      setMessage("Education saved locally.");
    } catch (reason) {
      setError(errorMessage(reason));
    } finally {
      setEducationBusyKey("");
    }
  };

  const deleteEducation = async (draft: EducationDraft) => {
    if (draft.id && !window.confirm(`Delete ${draft.degree || "this education record"} from ${draft.institution || "this institution"}?`)) {
      return;
    }
    if (!draft.id) {
      setEducations((current) => current.filter((education) => education.key !== draft.key));
      return;
    }
    setEducationBusyKey(draft.key);
    setError("");
    try {
      await DeleteEducation(draft.id);
      setEducations((current) => current.filter((education) => education.key !== draft.key));
      setMessage("Education deleted.");
    } catch (reason) {
      setError(errorMessage(reason));
    } finally {
      setEducationBusyKey("");
    }
  };

  const saveCertification=async(event:FormEvent,draft:CertificationDraft)=>{event.preventDefault();setCredentialBusyKey(draft.key);setError("");try{const saved=await SaveCertification(new domain.CertificationInput(toCertificationInput(draft)));const normalized=certificationToDraft(saved as unknown as Certification);setCertifications(current=>current.map(item=>item.key===draft.key?normalized:item));setMessage("Certification saved locally.")}catch(reason){setError(errorMessage(reason))}finally{setCredentialBusyKey("")}};
  const deleteCertification=async(draft:CertificationDraft)=>{if(!draft.id){setCertifications(current=>current.filter(item=>item.key!==draft.key));return}if(!window.confirm(`Delete ${draft.name||"this certification"}?`))return;setCredentialBusyKey(draft.key);try{await DeleteCertification(draft.id);setCertifications(current=>current.filter(item=>item.key!==draft.key));setMessage("Certification deleted.")}catch(reason){setError(errorMessage(reason))}finally{setCredentialBusyKey("")}};
  const saveAchievement=async(event:FormEvent,draft:AchievementDraft)=>{event.preventDefault();setCredentialBusyKey(draft.key);setError("");try{const saved=await SaveAchievement(new domain.AchievementInput(toAchievementInput(draft)));const normalized=achievementToDraft(saved as unknown as Achievement);setAchievements(current=>current.map(item=>item.key===draft.key?normalized:item));setMessage("Achievement saved locally.")}catch(reason){setError(errorMessage(reason))}finally{setCredentialBusyKey("")}};
  const deleteAchievement=async(draft:AchievementDraft)=>{if(!draft.id){setAchievements(current=>current.filter(item=>item.key!==draft.key));return}if(!window.confirm(`Delete ${draft.title||"this achievement"}?`))return;setCredentialBusyKey(draft.key);try{await DeleteAchievement(draft.id);setAchievements(current=>current.filter(item=>item.key!==draft.key));setMessage("Achievement deleted.")}catch(reason){setError(errorMessage(reason))}finally{setCredentialBusyKey("")}};

  const toggleProject = (key: string) => {
    if (!projects.some((project) => project.key === key && isProjectSelectable(project))) return;
    setSelectedProjectKeys((current) => current.includes(key) ? current.filter((item) => item !== key) : [...current, key]);
  };

  const saveCurrentAISettings = async (provider = aiProvider, nextOllamaModel = ollamaModel, nextGeminiModel = geminiModel) => {
    await SaveAISettings({ provider, ollamaEndpoint, ollamaModel: nextOllamaModel, geminiModel: nextGeminiModel });
  };

  const checkAIProvider = async () => {
    setAIBusy("check");
    setError("");
    setMessage("");
    try {
      await saveCurrentAISettings();
      const status = await (aiProvider === "ollama" ? CheckOllama(ollamaEndpoint) : CheckGemini()) as unknown as AIProviderStatus;
      setAIProviderStatus(status);
      if (status.endpoint) setOllamaEndpoint(status.endpoint);
			const currentModel = aiProvider === "ollama" ? ollamaModel : geminiModel;
			if (status.models.length && !status.models.includes(currentModel)) {
				if (aiProvider === "ollama") setOllamaModel(status.models[0]);
				else setGeminiModel(status.models[0]);
				await saveCurrentAISettings(aiProvider, aiProvider === "ollama" ? status.models[0] : ollamaModel, aiProvider === "gemini" ? status.models[0] : geminiModel);
			}
      if (status.available) setMessage(status.message);
      else setError(status.message);
    } catch (reason) {
      setAIProviderStatus(null);
      setError(errorMessage(reason));
    } finally {
      setAIBusy("");
    }
  };

	const changeAIProvider = async (provider: AIProvider) => {
		setAIProvider(provider);
		setAIProviderStatus(null);
		setAIRun(null);
		try {
			await saveCurrentAISettings(provider);
		} catch (reason) {
			setError(errorMessage(reason));
		}
	};

	const changeAIModel = async (model: string) => {
		if (aiProvider === "ollama") setOllamaModel(model);
		else setGeminiModel(model);
		try {
			await saveCurrentAISettings(aiProvider, aiProvider === "ollama" ? model : ollamaModel, aiProvider === "gemini" ? model : geminiModel);
		} catch (reason) {
			setError(errorMessage(reason));
		}
	};

	const saveGeminiCredential = async () => {
		setAIBusy("credential");
		setError("");
		try {
			const status = await SaveGeminiAPIKey(geminiAPIKey) as unknown as CredentialStatus;
			setGeminiCredential(status);
			setGeminiAPIKey("");
			setAIProviderStatus(null);
			setMessage(status.message);
		} catch (reason) {
			setError(errorMessage(reason));
		} finally {
			setAIBusy("");
		}
	};

	const deleteGeminiCredential = async () => {
		setAIBusy("credential");
		setError("");
		try {
			const status = await DeleteGeminiAPIKey() as unknown as CredentialStatus;
			setGeminiCredential(status);
			setGeminiAPIKey("");
			setAIProviderStatus(null);
			setGeminiModel("");
			setMessage(status.message);
		} catch (reason) {
			setError(errorMessage(reason));
		} finally {
			setAIBusy("");
		}
	};

  const generateAITailoring = async () => {
    setAIBusy("generate");
    setError("");
    setMessage("");
    setAIRun(null);
    try {
      const run = await GenerateAITailoring({
        jobId: jobDraft.id,
        selectedFactIds: selectedFactIDs,
			provider: aiProvider,
			model: aiProvider === "ollama" ? ollamaModel : geminiModel,
			endpoint: aiProvider === "ollama" ? ollamaEndpoint : "",
      }) as unknown as AIRun;
      setAIRun(run);
      setAIRuns((current) => [run, ...current.filter((item) => item.id !== run.id)]);
      setReviewedProposals(run.proposals.map((proposal) => ({ ...proposal, supportingFactIds: [...proposal.supportingFactIds] })));
      setAcceptedProposalIDs(run.proposals.map((proposal) => proposal.targetFactId));
      setMessage(run.validationPassed ? `${run.proposals.length} evidence-constrained proposals are ready for review.` : "The AI run was recorded but did not pass validation.");
    } catch (reason) {
      setError(errorMessage(reason));
    } finally {
      setAIBusy("");
    }
  };

  const cancelAITailoring = async () => {
    await CancelAITailoring();
    setMessage("Cancelling the current AI request…");
  };

  const updateReviewedProposal = (targetFactID: string, text: string) => {
    setReviewedProposals((current) => current.map((proposal) => proposal.targetFactId === targetFactID ? { ...proposal, text } : proposal));
  };

  const toggleAcceptedProposal = (targetFactID: string) => {
    setAcceptedProposalIDs((current) => current.includes(targetFactID) ? current.filter((id) => id !== targetFactID) : [...current, targetFactID]);
  };

  const acceptAITailoring = async () => {
    if (!aiRun || !confirmDiscardLatex()) return;
    setAIBusy("accept");
    setError("");
    setMessage("");
    try {
      const proposals = reviewedProposals.filter((proposal) => acceptedProposalIDs.includes(proposal.targetFactId));
      const result = await AcceptAITailoring(new domain.AcceptAITailoringInput({ runId: aiRun.id, templateId: selectedTemplateID, proposals }));
      const version = result.version as unknown as ResumeVersion;
      setApplications((await ListApplications()) as unknown as Application[]);
      setAIRuns((await ListAIRuns()) as unknown as AIRun[]);
      setLatexSource(version.latexSource);
      setOpenVersionID(version.id);
      setLatexBaseline(version.latexSource);
      setCompileResult(null);
      setAIRun({ ...aiRun, resumeVersionId: version.id, acceptedAt: new Date().toISOString() });
      setView("latex");
      setMessage(`Accepted AI wording into immutable resume version ${version.versionNumber}.`);
    } catch (reason) {
      setError(errorMessage(reason));
    } finally {
      setAIBusy("");
    }
  };

  const exportBackup = async () => {
    setBackupBusy("export");
    setError("");
    setMessage("");
    try {
      const result = await ExportProfileBackup();
      if (!result.cancelled) {
        setLastBackupResult(result);
        setMessage(`Backup exported to ${result.path}`);
      }
    } catch (reason) {
      setError(errorMessage(reason));
    } finally {
      setBackupBusy("");
    }
  };

  const importBackup = async () => {
    if (!window.confirm("Restore a TailorCV backup? Your current data will be replaced after the selected file passes validation. A safety copy of it is saved first, so you can undo this.")) return;
    setBackupBusy("import");
    setError("");
    setMessage("");
    try {
      const result = await ImportProfileBackup();
      if (!result.cancelled) {
        await loadWorkspaceData();
        setLastBackupResult(result);
        setMessage(`Backup restored from ${result.path}`);
      }
    } catch (reason) {
      setError(errorMessage(reason));
    } finally {
      setBackupBusy("");
    }
  };

  const restoreAutomaticBackup = async (backup: domain.AutomaticBackup) => {
    const when = new Date(backup.createdAt).toLocaleString();
    if (!window.confirm(`Restore the backup from ${when}? Your current data will be replaced. A safety copy of it is saved first, so you can undo this.`)) return;
    setBackupBusy("import");
    setError("");
    setMessage("");
    try {
      const result = await RestoreAutomaticBackup(backup.name);
      await loadWorkspaceData();
      setLastBackupResult(result);
      setMessage(`Restored the backup from ${when}.`);
    } catch (reason) {
      setError(errorMessage(reason));
    } finally {
      setBackupBusy("");
    }
  };

  const syncGitHubProjects = async () => {
    setGitHubBusy(true);
    setError("");
    setMessage("");
    try {
      const result = await ImportGitHubProjects();
      const savedProjects = await ListProjects();
      const projectDrafts = (savedProjects as unknown as Project[]).map(projectToDraft);
      setProjects(projectDrafts);
      setSelectedProjectKeys((current) => current.filter((key) => projectDrafts.some((project) => project.key === key && isProjectSelectable(project))));
      const languageNote = result.languageFallbacks ? ` ${result.languageFallbacks} repositories used primary-language fallback because GitHub language details were unavailable.` : "";
      const readmeNote = result.readmeFallbacks ? ` ${result.readmeFallbacks} README snapshots could not be refreshed.` : "";
      setMessage(`GitHub sync complete: ${result.imported} imported, ${result.updated} refreshed, ${result.skipped} skipped.${languageNote}${readmeNote}`);
    } catch (reason) {
      setError(errorMessage(reason));
    } finally {
      setGitHubBusy(false);
    }
  };

  const importGitHubProject = async (repositoryURL: string) => {
    setGitHubBusy(true);
    setError("");
    setMessage("");
    try {
      const project = await ImportGitHubProject(repositoryURL) as unknown as Project;
      const savedProjects = await ListProjects();
      setProjects((savedProjects as unknown as Project[]).map(projectToDraft));
      setMessage(`${project.name} imported from GitHub. Review it before adding it to a resume.`);
    } catch (reason) {
      setError(errorMessage(reason));
    } finally {
      setGitHubBusy(false);
    }
  };

  const selectedSavedProjectIDs = () => selectedProjectKeys
    .map((key) => projects.find((project) => project.key === key)?.id ?? "")
    .filter(Boolean);

  const useResumeTemplate = async (templateID: string, openEditor = false) => {
    if (!confirmDiscardLatex()) return;
    setTemplateBusyKey(templateID);
    setError("");
    setMessage("");
    try {
      const template = await SelectResumeTemplate(templateID);
      const rendered = await RenderResumeTemplate(templateID, selectedSavedProjectIDs());
      setSelectedTemplateID(templateID);
      setLatexSource(rendered);
      setOpenVersionID("");
      setLatexBaseline(rendered);
      setCompileResult(null);
      setMessage(`${template.name} loaded with your saved resume data.`);
      if (openEditor) setView("latex");
    } catch (reason) {
      setError(errorMessage(reason));
    } finally {
      setTemplateBusyKey("");
    }
  };

  const importResumeTemplate = async () => {
    if (!confirmDiscardLatex()) return;
    setTemplateBusyKey("import");
    setError("");
    setMessage("");
    try {
      const imported = await ImportResumeTemplate();
      if (!imported.id) return;
      const refreshed = await ListResumeTemplates();
      setTemplates(refreshed);
      await SelectResumeTemplate(imported.id);
      setSelectedTemplateID(imported.id);
      const rendered = await RenderResumeTemplate(imported.id, selectedSavedProjectIDs());
      setLatexSource(rendered);
      setOpenVersionID("");
      setLatexBaseline(rendered);
      setCompileResult(null);
      setMessage(`${imported.name} imported into your local template library.`);
    } catch (reason) {
      setError(errorMessage(reason));
    } finally {
      setTemplateBusyKey("");
    }
  };

  const duplicateResumeTemplate = async (template: domain.ResumeTemplate) => {
    setTemplateBusyKey(template.id);
    setError("");
    setMessage("");
    try {
      const saved = await SaveResumeTemplate(new domain.ResumeTemplateInput({
        id: "",
        name: `${template.name} Copy`,
        description: `Editable copy of ${template.name}.`,
        source: template.source,
      }));
      setTemplates(await ListResumeTemplates());
      setMessage(`${saved.name} added to your template library.`);
    } catch (reason) {
      setError(errorMessage(reason));
    } finally {
      setTemplateBusyKey("");
    }
  };

  const saveCurrentTemplate = async () => {
    const active = templates.find((template) => template.id === selectedTemplateID);
    if (!active) return;
    setTemplateBusyKey(active.id);
    setError("");
    setMessage("");
    try {
      const saved = await SaveResumeTemplate(new domain.ResumeTemplateInput({
        id: active.builtIn ? "" : active.id,
        name: active.builtIn ? `${active.name} Copy` : active.name,
        description: active.builtIn ? `Editable copy of ${active.name}.` : active.description,
        source: latexSource,
      }));
      await SelectResumeTemplate(saved.id);
      setSelectedTemplateID(saved.id);
      setLatexBaseline(latexSource);
      setTemplates(await ListResumeTemplates());
      setMessage(active.builtIn ? `${saved.name} created and selected.` : `${saved.name} updated.`);
    } catch (reason) {
      setError(errorMessage(reason));
    } finally {
      setTemplateBusyKey("");
    }
  };

  const deleteResumeTemplate = async (template: domain.ResumeTemplate) => {
    if (template.builtIn || !window.confirm(`Delete the template ${template.name}?`)) return;
    if (template.id === selectedTemplateID && !confirmDiscardLatex()) return;
    setTemplateBusyKey(template.id);
    setError("");
    setMessage("");
    try {
      await DeleteResumeTemplate(template.id);
      const refreshed = await ListResumeTemplates();
      setTemplates(refreshed);
      if (selectedTemplateID === template.id) {
        const fallbackID = await GetSelectedResumeTemplateID();
        setSelectedTemplateID(fallbackID);
        const rendered = await RenderResumeTemplate(fallbackID, selectedSavedProjectIDs());
        setLatexSource(rendered);
        setOpenVersionID("");
        setLatexBaseline(rendered);
        setCompileResult(null);
      }
      setMessage(`${template.name} deleted.`);
    } catch (reason) {
      setError(errorMessage(reason));
    } finally {
      setTemplateBusyKey("");
    }
  };

  const restoreLatexDraft = (draft: domain.LatexDraft) => {
    const version = applications.flatMap((application) => application.versions).find((item) => item.id === draft.versionId);
    // Keep the baseline as the version (or current template), so the draft shows as unsaved.
    if (version) {
      setOpenVersionID(version.id);
      setLatexBaseline(version.latexSource);
    }
    setLatexSource(draft.source);
    setCompileResult(null);
    setRecoveredDraft(null);
    setView("latex");
    setMessage("Recovered your unsaved LaTeX edits. Save them to keep them.");
  };

  const discardLatexDraft = () => {
    setRecoveredDraft(null);
    ClearLatexDraft().catch((reason) => setError(errorMessage(reason)));
  };

  const updateLatexSource = (source: string) => {
    compileRequestRef.current += 1;
    setLatexSource(source);
    setCompileResult(null);
    setMessage("");
  };

  const compileLatexSource = async (source: string, automatic: boolean) => {
    const requestID = ++compileRequestRef.current;
    setCompileBusy(true);
    if (!automatic) {
      setError("");
      setMessage("");
    }
    try {
      const persistArtifact = !automatic && openVersionID !== "" && source === latexBaseline;
      const result = persistArtifact ? await CompileResumeVersion(openVersionID) : await CompileLatex(source);
      if (requestID !== compileRequestRef.current) return;
      setCompileResult(result);
      if (!automatic) {
        setMessage(result.success
          ? `Compiled with ${result.engine} in ${result.durationMs} ms.`
          : `Compilation found ${result.diagnostics?.length || 1} issue${result.diagnostics?.length === 1 ? "" : "s"}.`);
      }
      if (persistArtifact) setApplications((await ListApplications()) as unknown as Application[]);
    } catch (reason) {
      if (requestID !== compileRequestRef.current) return;
      setCompileResult(null);
      setError(errorMessage(reason));
    } finally {
      if (requestID === compileRequestRef.current) setCompileBusy(false);
    }
  };

  const compileLatex = () => {
    if (compileTimerRef.current !== undefined) window.clearTimeout(compileTimerRef.current);
    void compileLatexSource(latexSource, false);
  };

  useEffect(() => {
    if (compileTimerRef.current !== undefined) window.clearTimeout(compileTimerRef.current);
    if (view !== "latex" || !latexSource.trim()) {
      compileRequestRef.current += 1;
      setCompileBusy(false);
      return;
    }
    if (skipNextAutomaticCompileRef.current) {
      skipNextAutomaticCompileRef.current = false;
      compileRequestRef.current += 1;
      setCompileBusy(false);
      return;
    }
    compileTimerRef.current = window.setTimeout(() => {
      compileTimerRef.current = undefined;
      void compileLatexSource(latexSource, true);
    }, 900);
    return () => {
      if (compileTimerRef.current !== undefined) window.clearTimeout(compileTimerRef.current);
    };
  }, [view, latexSource]);

  // Default export file name, e.g. "Ada Lovelace - Acme - Staff Engineer".
  const openApplication = applications.find((application) => application.versions.some((version) => version.id === openVersionID));
  const openApplicationJob = jobs.find((job) => job.id === openApplication?.jobId);
  const coverageJobID = openApplicationJob?.id ?? "";
  const [keywordCoverage, setKeywordCoverage] = useState<domain.KeywordCoverage | null>(null);
  useEffect(() => {
    if (!coverageJobID || !latexSource.trim()) {
      setKeywordCoverage(null);
      return;
    }
    let cancelled = false;
    const timer = window.setTimeout(() => {
      KeywordCoverage(coverageJobID, latexSource).then((coverage) => { if (!cancelled) setKeywordCoverage(coverage); }).catch(() => { if (!cancelled) setKeywordCoverage(null); });
    }, 400);
    return () => { cancelled = true; window.clearTimeout(timer); };
  }, [coverageJobID, latexSource]);
  const exportName = [profile.name, ...(openApplicationJob ? [openApplicationJob.company, openApplicationJob.role] : [openApplication?.name])].filter(Boolean).join(" - ");

  const exportCompiledPDF = async () => {
    setError("");
    try {
      const result = await ExportCompiledPDF(exportName);
      if (!result.cancelled) setMessage(`PDF exported to ${result.path}`);
    } catch (reason) {
      setError(errorMessage(reason));
    }
  };

  const exportLatexSource = async () => {
    setError("");
    try {
      const result = await ExportLatexSource(latexSource, exportName);
      if (!result.cancelled) setMessage(`LaTeX source exported to ${result.path}`);
    } catch (reason) {
      setError(errorMessage(reason));
    }
  };

  const selectedProjects = projects.filter((project) => selectedProjectKeys.includes(project.key));
  const activeTemplate = templates.find((template) => template.id === selectedTemplateID);
  // General resumes have no job, so only match once the draft job is saved.
  const currentApplication = jobDraft.id ? applications.find((application) => application.jobId === jobDraft.id) : undefined;
  const hasCareerEvidence = profile.skills.length > 0 || experiences.some((experience) => experience.bullets.length > 0) || projects.some((project) => project.skills.length > 0 || project.bullets.length > 0);
  const showOnboarding = !busy && onboardingPending;

  return (
    <div className="studio-shell">
      {showOnboarding && !resumeImport && <Onboarding profile={profile} skillsText={skillsText} busy={busy} onChange={updateProfile} onSkillsChange={setSkillsText} onSubmit={saveProfile} onSkip={() => setOnboardingPending(false)} onImport={previewResumeImport} />}
      {resumeImport && <ResumeImportDialog data={resumeImport} busy={resumeImportBusy} onImport={applyResumeImport} onCancel={() => setResumeImport(null)} />}
      <TopToolbar profile={profile} templateName={activeTemplate?.name ?? "Resume template"} compiling={compileBusy} canExport={Boolean(compileResult?.success && compileResult.pdfBase64)} onCompile={compileLatex} onExport={exportCompiledPDF} />

      <div className="studio-body">
        <aside className="studio-sidebar">
          <div className="sidebar-intro">
            <span>Workspace</span>
            <p>Build once. Tailor precisely.</p>
          </div>
          <nav aria-label="Resume sections">
            <p className="nav-section-label">Build</p>
            <NavButton active={view === "overview"} label="Home" icon="home" onClick={() => setView("overview")} />
            <NavButton active={view === "profile"} label="Profile" icon="user" onClick={() => setView("profile")} />
            <NavButton active={view === "experience"} label="Experience" icon="briefcase" badge={experiences.length || undefined} onClick={() => setView("experience")} />
            <NavButton active={view === "projects"} label="Projects" icon="folder" badge={selectedProjectKeys.length || undefined} onClick={() => setView("projects")} />
            <NavButton active={view === "education"} label="Education" icon="education" badge={educations.length || undefined} onClick={() => setView("education")} />
            <NavButton active={view === "credentials"} label="Credentials" icon="check" badge={(certifications.length+achievements.length)||undefined} onClick={() => setView("credentials")} />
            <NavButton active={view === "skills"} label="Skills" icon="sparkles" badge={profile.skills.length || undefined} onClick={() => setView("skills")} />
            <p className="nav-section-label nav-section-spaced">Tailor</p>
            <NavButton active={view === "templates"} label="Templates" icon="template" badge={templates.length || undefined} onClick={() => setView("templates")} />
            <NavButton active={view === "latex"} label="LaTeX source" icon="code" onClick={() => setView("latex")} />
            <NavButton active={view === "job"} label="Job match" icon="target" badge={analysis ? `${analysis.score}%` : jobs.length || undefined} onClick={() => setView("job")} />
            <NavButton active={view === "ai"} label="AI assistant" icon="chat" badge="Setup" onClick={() => setView("ai")} />
            <p className="nav-section-label nav-section-spaced">System</p>
            <NavButton active={view === "data"} label="Backup & restore" icon="database" onClick={() => setView("data")} />
          </nav>

          <button className="provider-status" onClick={() => setView("ai")}>
            <span className={`status-dot ${aiProviderStatus?.available ? "" : "muted"}`} />
            <span><strong>AI provider</strong><small>{aiProvider === "ollama" ? "Ollama" : "Gemini"} · {aiProviderStatus?.available ? `${aiProviderStatus.models.length} models` : "not checked"}</small></span>
            <span aria-hidden="true">›</span>
          </button>
        </aside>

        <main className="studio-workspace">
          {error && <div className="notice error" role="alert">{error}<button aria-label="Dismiss error" onClick={() => setError("")}>×</button></div>}
          {recoveredDraft && <div className="notice draft-recovery" role="alert">Unsaved LaTeX edits from {new Date(recoveredDraft.savedAt).toLocaleString()} were found.<span><button className="primary-button" onClick={() => restoreLatexDraft(recoveredDraft)}>Restore edits</button><button className="secondary-button" onClick={discardLatexDraft}>Discard</button></span></div>}
          {message && <div className="notice success" role="status">{message}<button aria-label="Dismiss message" onClick={() => setMessage("")}>×</button></div>}

          <div className="editor-pane">
            {view === "overview" && <WorkspaceOverview profile={profile} completion={completion} experiences={experiences} projects={projects} applications={applications} jobs={jobs} busy={versionBusy} onOpen={setView} onNewResume={() => { newJob(); setView("job"); }} onNewGeneralResume={createGeneralResume} onOpenJob={(job) => { openJob(job); setView("job"); }} onOpenVersion={openResumeVersion} onRename={renameResume} onDuplicate={duplicateResume} onDelete={deleteResume} statusBusy={applicationStatusBusy} onStatusChange={updateApplicationStatus} onSaveDetails={saveApplicationDetails} coverLetters={coverLetters} />}
            {view === "profile" && <ProfileWorkspace profile={profile} busy={busy} message={message} onChange={updateProfile} onSubmit={saveProfile} onImport={previewResumeImport} />}
            {view === "experience" && <section className="workspace-panel scroll-panel"><PanelHeader eyebrow="Career evidence" title="Experience" description="Keep every claim factual, ordered, and reviewable." action={<button className="secondary-button" onClick={addExperience}>Add role</button>} /><ExperienceSection experiences={experiences} busyKey={experienceBusyKey} onAdd={addExperience} onUpdate={updateExperience} onSave={saveExperience} onDelete={deleteExperience} /></section>}
            {view === "projects" && <ProjectWorkspace projects={projects} selectedKeys={selectedProjectKeys} busyKey={projectBusyKey} readmeBusyKey={projectReadmeBusyKey} aiReady={Boolean(aiProviderStatus?.available && (aiProvider === "ollama" ? ollamaModel : geminiModel))} githubUsername={profile.githubUsername} githubBusy={githubBusy} onToggle={toggleProject} onAdd={addProject} onUpdate={updateProject} onSave={saveProject} onDelete={deleteProject} onGenerateReadmeBullets={generateProjectReadmeBullets} onImportGitHubProject={importGitHubProject} onSyncGitHub={syncGitHubProjects} onOpenProfile={() => setView("profile")} />}
            {view === "education" && <EducationWorkspace educations={educations} busyKey={educationBusyKey} onAdd={addEducation} onUpdate={updateEducation} onSave={saveEducation} onDelete={deleteEducation} />}
            {view === "credentials"&&<CredentialsWorkspace certifications={certifications} achievements={achievements} busyKey={credentialBusyKey} onCertifications={setCertifications} onAchievements={setAchievements} onSaveCertification={saveCertification} onDeleteCertification={deleteCertification} onSaveAchievement={saveAchievement} onDeleteAchievement={deleteAchievement}/>}
            {view === "skills" && <SkillsWorkspace skillsText={skillsText} busy={busy} message={message} onChange={setSkillsText} onSubmit={saveProfile} />}
            {view === "templates" && <TemplatesWorkspace templates={templates} selectedID={selectedTemplateID} busyKey={templateBusyKey} onImport={importResumeTemplate} onUse={(id) => useResumeTemplate(id)} onEdit={(id) => useResumeTemplate(id, true)} onDuplicate={duplicateResumeTemplate} onDelete={deleteResumeTemplate} />}
            {view === "latex" && <LatexWorkspace source={latexSource} template={activeTemplate} result={compileResult} busy={templateBusyKey !== ""} compiling={compileBusy} versionBusy={versionBusy} hasOpenVersion={openVersionID !== ""} versionDirty={openVersionID !== "" && latexSource !== latexBaseline} onChange={updateLatexSource} onSave={saveCurrentTemplate} onSaveVersion={saveResumeVersionEdit} onSaveAsResume={() => saveAsNewResume(`${activeTemplate?.name ?? "General"} resume`, latexSource)} onReload={() => useResumeTemplate(selectedTemplateID)} onCompile={compileLatex} onExport={exportLatexSource} coverage={keywordCoverage} />}
            {view === "job" && <JobTailor job={jobDraft} jobs={jobs} application={currentApplication} analysis={analysis} selectedFactIDs={selectedFactIDs} busy={busy} versionBusy={versionBusy} statusBusy={applicationStatusBusy} hasEvidence={hasCareerEvidence} templateName={activeTemplate?.name ?? "Selected template"} onChange={updateJobDraft} onNew={newJob} onOpen={openJob} onDelete={deleteJob} onToggleEvidence={toggleEvidenceFact} onCreateVersion={createResumeVersion} onOpenVersion={openResumeVersion} onStatusChange={updateApplicationStatus} onSubmit={analyzeJob} onProfile={() => setView("skills")} />}
            {view === "ai" && <AIWorkspace provider={aiProvider} endpoint={ollamaEndpoint} status={aiProviderStatus} model={aiProvider === "ollama" ? ollamaModel : geminiModel} credential={geminiCredential} apiKey={geminiAPIKey} run={aiRun} runs={aiRuns} proposals={reviewedProposals} acceptedProposalIDs={acceptedProposalIDs} job={jobDraft} analysis={analysis} selectedFactIDs={selectedFactIDs} templateName={activeTemplate?.name ?? "Selected template"} busy={aiBusy} onProviderChange={changeAIProvider} onEndpointChange={(value) => { setOllamaEndpoint(value); setAIProviderStatus(null); }} onModelChange={changeAIModel} onAPIKeyChange={setGeminiAPIKey} onSaveCredential={saveGeminiCredential} onDeleteCredential={deleteGeminiCredential} onCheck={checkAIProvider} onGenerate={generateAITailoring} onCancel={cancelAITailoring} onProposalChange={updateReviewedProposal} onToggleProposal={toggleAcceptedProposal} onAccept={acceptAITailoring} onOpenJob={() => setView("job")} />}
            {view === "data" && <DataWorkspace profile={profile} experiences={experiences} projects={projects} educations={educations} jobs={jobs} applications={applications} busy={backupBusy} lastResult={lastBackupResult} onExport={exportBackup} onImport={importBackup} onRestoreAutomatic={restoreAutomaticBackup} />}
          </div>

          <ResumePreview profile={profile} experiences={experiences} projects={selectedProjects} educations={educations} certifications={certifications} achievements={achievements} compileResult={compileResult} />
        </main>
      </div>
    </div>
  );
}

import { FormEvent, useEffect, useMemo, useRef, useState } from "react";
import { defaultKeymap, history, historyKeymap, indentWithTab } from "@codemirror/commands";
import { defaultHighlightStyle, syntaxHighlighting, StreamLanguage } from "@codemirror/language";
import { stex } from "@codemirror/legacy-modes/mode/stex";
import { EditorState } from "@codemirror/state";
import { drawSelection, EditorView, highlightActiveLine, highlightSpecialChars, keymap, lineNumbers } from "@codemirror/view";
import pdfWorkerURL from "pdfjs-dist/build/pdf.worker.min.mjs?url";
import {
  AcceptAITailoring,
  AnalyzeJobDescription,
  CancelAITailoring,
  CheckGemini,
  CheckOllama,
  CompileLatex,
  CompileResumeVersion,
  CreateResumeVersion,
  DeleteEducation,
  DeleteCertification,
  DeleteAchievement,
  DeleteExperience,
  DeleteGeminiAPIKey,
  DeleteJob,
  DeleteProject,
  DeleteResumeTemplate,
  ExportCompiledPDF,
  ExportLatexSource,
  ExportProfileBackup,
  GenerateProjectReadmeBullets,
  GetProfile,
  GetAISettings,
  GetGeminiCredentialStatus,
  GetSelectedResumeTemplateID,
  GenerateAITailoring,
  ImportGitHubProjects,
  ImportGitHubProject,
  ImportProfileBackup,
  ImportResumeTemplate,
  ListEducations,
  ListCertifications,
  ListAchievements,
  ListAIRuns,
  ListApplications,
  ListExperiences,
  ListJobs,
  ListProjects,
  ListResumeTemplates,
  OpenResumeVersion,
  RenderResumeTemplate,
  SaveEducation,
  SaveCertification,
  SaveAchievement,
  SaveAISettings,
  SaveExperience,
  SaveGeminiAPIKey,
  SaveProject,
  SaveProfile,
  SaveResumeTemplate,
  SaveResumeVersionEdit,
  SelectResumeTemplate,
  UpdateApplicationStatus,
} from "../wailsjs/go/main/App";
import { domain } from "../wailsjs/go/models";
import {
  Education,
  EducationDraft,
  educationToDraft,
  newEducationDraft,
  toEducationInput,
} from "./lib/education";
import { Achievement, AchievementDraft, Certification, CertificationDraft, achievementToDraft, certificationToDraft, newAchievementDraft, newCertificationDraft, toAchievementInput, toCertificationInput } from "./lib/credentials";
import {
  EvidenceBullet,
  Experience,
  ExperienceDraft,
  experienceToDraft,
  moveItem,
  newEvidenceBullet,
  newExperienceDraft,
  toExperienceInput,
} from "./lib/experience";
import {
  Project,
  ProjectDraft,
  filterProjects,
  isProjectSelectable,
  newProjectDraft,
  projectToDraft,
  reconcileSelectedProjectKeys,
  removeSelectedProjectKey,
  toggleProjectLanguage,
  toProjectInput,
} from "./lib/project";
import {
  emptyProfile,
  parseSkills,
  Profile,
  profileCompletion,
} from "./lib/profile";
import {
  AIWorkspace,
  type AIRun,
  type AIProposal,
  type AIProvider,
  type AIProviderStatus,
  type CredentialStatus,
  type EvidenceMatch,
  type Job,
  type JobAnalysis,
} from "./features/ai/AIWorkspace";
import { ApplicationStatusControl, type ApplicationStatus } from "./features/applications/ApplicationStatusControl";
import { Onboarding } from "./features/onboarding/Onboarding";

type View = "overview" | "profile" | "experience" | "projects" | "education" | "credentials" | "skills" | "templates" | "latex" | "job" | "ai" | "data";

type ResumeVersion = {
  id: string;
  applicationId: string;
  versionNumber: number;
  jobDescriptionSnapshot: string;
  selectedFactIds: string[];
  latexSource: string;
  templateId: string;
  rankingExplanations: EvidenceMatch[];
  contentHash: string;
  compileSuccess: boolean;
  compileEngine: string;
  compileDurationMs: number;
  compileDiagnostics: domain.CompileDiagnostic[];
  compiledAt: string;
  pdfAvailable: boolean;
  createdAt: string;
};

type Application = {
  id: string;
  jobId: string;
  status: ApplicationStatus;
  selectedFactIds: string[];
  versions: ResumeVersion[];
  createdAt: string;
  updatedAt: string;
};

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

function errorMessage(error: unknown): string {
  if (error instanceof Error) return error.message;
  if (typeof error === "string") return error;
  return "Something went wrong. Please try again.";
}

function profileHasData(profile: Profile): boolean {
  return Boolean(profile.name || profile.email || profile.headline || profile.phone || profile.location || profile.website || profile.githubUsername || profile.linkedInUrl || profile.summary || profile.skills.length);
}

export default function App() {
  const [view, setView] = useState<View>("projects");
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
  const [savedVersionSource, setSavedVersionSource] = useState("");
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
    setSavedVersionSource("");
    setCompileResult(null);
  };

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
    setVersionBusy(true);
    setError("");
    setMessage("");
    try {
      const result = await CreateResumeVersion({ jobId: jobDraft.id, selectedFactIds: selectedFactIDs, templateId: selectedTemplateID });
      const version = result.version as unknown as ResumeVersion;
      setApplications((await ListApplications()) as unknown as Application[]);
      setLatexSource(version.latexSource);
      setOpenVersionID(version.id);
      setSavedVersionSource(version.latexSource);
      setCompileResult(null);
      setMessage(`Resume version ${version.versionNumber} saved from ${selectedFactIDs.length} selected facts.`);
    } catch (reason) {
      setError(errorMessage(reason));
    } finally {
      setVersionBusy(false);
    }
  };

  const updateApplicationStatus = async (status: ApplicationStatus) => {
    if (!currentApplication || currentApplication.status === status) return;
    setApplicationStatusBusy(true);
    setError("");
    setMessage("");
    try {
      const updated = await UpdateApplicationStatus({ applicationId: currentApplication.id, status });
      setApplications((current) => current.map((application) => application.id === currentApplication.id ? updated as unknown as Application : application));
      setMessage(`Application marked ${status}.`);
    } catch (reason) {
      setError(errorMessage(reason));
    } finally {
      setApplicationStatusBusy(false);
    }
  };

  const openResumeVersion = async (version: ResumeVersion) => {
    setError("");
    setCompileResult(null);
    try {
      const workspace = await OpenResumeVersion(version.id);
      const opened = workspace.version as unknown as ResumeVersion;
      setLatexSource(opened.latexSource);
      setOpenVersionID(opened.id);
      setSavedVersionSource(opened.latexSource);
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
      setSavedVersionSource(version.latexSource);
      setApplications((await ListApplications()) as unknown as Application[]);
      setMessage(`Saved edited source as immutable version ${version.versionNumber}.`);
    } catch (reason) {
      setError(errorMessage(reason));
    } finally {
      setVersionBusy(false);
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
    if (!aiRun) return;
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
      setSavedVersionSource(version.latexSource);
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
    if (!window.confirm("Restore a TailorCV backup? Your current profile, experience, projects, and education will be replaced after the selected file passes validation.")) return;
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
    setTemplateBusyKey(templateID);
    setError("");
    setMessage("");
    try {
      const template = await SelectResumeTemplate(templateID);
      const rendered = await RenderResumeTemplate(templateID, selectedSavedProjectIDs());
      setSelectedTemplateID(templateID);
      setLatexSource(rendered);
      setOpenVersionID("");
      setSavedVersionSource("");
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
      setLatexSource(await RenderResumeTemplate(imported.id, selectedSavedProjectIDs()));
      setOpenVersionID("");
      setSavedVersionSource("");
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
        setLatexSource(await RenderResumeTemplate(fallbackID, selectedSavedProjectIDs()));
        setCompileResult(null);
      }
      setMessage(`${template.name} deleted.`);
    } catch (reason) {
      setError(errorMessage(reason));
    } finally {
      setTemplateBusyKey("");
    }
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
      const persistArtifact = !automatic && openVersionID !== "" && source === savedVersionSource;
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

  const exportCompiledPDF = async () => {
    setError("");
    try {
      const result = await ExportCompiledPDF();
      if (!result.cancelled) setMessage(`PDF exported to ${result.path}`);
    } catch (reason) {
      setError(errorMessage(reason));
    }
  };

  const exportLatexSource = async () => {
    setError("");
    try {
      const result = await ExportLatexSource(latexSource);
      if (!result.cancelled) setMessage(`LaTeX source exported to ${result.path}`);
    } catch (reason) {
      setError(errorMessage(reason));
    }
  };

  const selectedProjects = projects.filter((project) => selectedProjectKeys.includes(project.key));
  const activeTemplate = templates.find((template) => template.id === selectedTemplateID);
  const currentApplication = applications.find((application) => application.jobId === jobDraft.id);
  const hasCareerEvidence = profile.skills.length > 0 || experiences.some((experience) => experience.bullets.length > 0) || projects.some((project) => project.skills.length > 0 || project.bullets.length > 0);
  const showOnboarding = !busy && onboardingPending;

  return (
    <div className="studio-shell">
      {showOnboarding && <Onboarding profile={profile} skillsText={skillsText} busy={busy} onChange={updateProfile} onSkillsChange={setSkillsText} onSubmit={saveProfile} onSkip={() => setOnboardingPending(false)} />}
      <TopToolbar profile={profile} templateName={activeTemplate?.name ?? "Resume template"} compiling={compileBusy} canExport={Boolean(compileResult?.success && compileResult.pdfBase64)} onCompile={compileLatex} onExport={exportCompiledPDF} />

      <div className="studio-body">
        <aside className="studio-sidebar">
          <div className="sidebar-intro">
            <span>Workspace</span>
            <p>Build once. Tailor precisely.</p>
          </div>
          <nav aria-label="Resume sections">
            <p className="nav-section-label">Build</p>
            <NavButton active={view === "overview"} label="Overview" icon="home" onClick={() => setView("overview")} />
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
          {message && <div className="notice success" role="status">{message}<button aria-label="Dismiss message" onClick={() => setMessage("")}>×</button></div>}

          <div className="editor-pane">
            {view === "overview" && <WorkspaceOverview profile={profile} completion={completion} experiences={experiences} projects={projects} applications={applications} onOpen={setView} />}
            {view === "profile" && <ProfileWorkspace profile={profile} busy={busy} message={message} onChange={updateProfile} onSubmit={saveProfile} />}
            {view === "experience" && <section className="workspace-panel scroll-panel"><PanelHeader eyebrow="Career evidence" title="Experience" description="Keep every claim factual, ordered, and reviewable." action={<button className="secondary-button" onClick={addExperience}>Add role</button>} /><ExperienceSection experiences={experiences} busyKey={experienceBusyKey} onAdd={addExperience} onUpdate={updateExperience} onSave={saveExperience} onDelete={deleteExperience} /></section>}
            {view === "projects" && <ProjectWorkspace projects={projects} selectedKeys={selectedProjectKeys} busyKey={projectBusyKey} readmeBusyKey={projectReadmeBusyKey} aiReady={Boolean(aiProviderStatus?.available && (aiProvider === "ollama" ? ollamaModel : geminiModel))} githubUsername={profile.githubUsername} githubBusy={githubBusy} onToggle={toggleProject} onAdd={addProject} onUpdate={updateProject} onSave={saveProject} onDelete={deleteProject} onGenerateReadmeBullets={generateProjectReadmeBullets} onImportGitHubProject={importGitHubProject} onSyncGitHub={syncGitHubProjects} onOpenProfile={() => setView("profile")} />}
            {view === "education" && <EducationWorkspace educations={educations} busyKey={educationBusyKey} onAdd={addEducation} onUpdate={updateEducation} onSave={saveEducation} onDelete={deleteEducation} />}
            {view === "credentials"&&<CredentialsWorkspace certifications={certifications} achievements={achievements} busyKey={credentialBusyKey} onCertifications={setCertifications} onAchievements={setAchievements} onSaveCertification={saveCertification} onDeleteCertification={deleteCertification} onSaveAchievement={saveAchievement} onDeleteAchievement={deleteAchievement}/>}
            {view === "skills" && <SkillsWorkspace skillsText={skillsText} busy={busy} message={message} onChange={setSkillsText} onSubmit={saveProfile} />}
            {view === "templates" && <TemplatesWorkspace templates={templates} selectedID={selectedTemplateID} busyKey={templateBusyKey} onImport={importResumeTemplate} onUse={(id) => useResumeTemplate(id)} onEdit={(id) => useResumeTemplate(id, true)} onDuplicate={duplicateResumeTemplate} onDelete={deleteResumeTemplate} />}
            {view === "latex" && <LatexWorkspace source={latexSource} template={activeTemplate} result={compileResult} busy={templateBusyKey !== ""} compiling={compileBusy} versionBusy={versionBusy} hasOpenVersion={openVersionID !== ""} versionDirty={openVersionID !== "" && latexSource !== savedVersionSource} onChange={updateLatexSource} onSave={saveCurrentTemplate} onSaveVersion={saveResumeVersionEdit} onReload={() => useResumeTemplate(selectedTemplateID)} onCompile={compileLatex} onExport={exportLatexSource} />}
            {view === "job" && <JobTailor job={jobDraft} jobs={jobs} application={currentApplication} analysis={analysis} selectedFactIDs={selectedFactIDs} busy={busy} versionBusy={versionBusy} statusBusy={applicationStatusBusy} hasEvidence={hasCareerEvidence} templateName={activeTemplate?.name ?? "Selected template"} onChange={updateJobDraft} onNew={newJob} onOpen={openJob} onDelete={deleteJob} onToggleEvidence={toggleEvidenceFact} onCreateVersion={createResumeVersion} onOpenVersion={openResumeVersion} onStatusChange={updateApplicationStatus} onSubmit={analyzeJob} onProfile={() => setView("skills")} />}
            {view === "ai" && <AIWorkspace provider={aiProvider} endpoint={ollamaEndpoint} status={aiProviderStatus} model={aiProvider === "ollama" ? ollamaModel : geminiModel} credential={geminiCredential} apiKey={geminiAPIKey} run={aiRun} runs={aiRuns} proposals={reviewedProposals} acceptedProposalIDs={acceptedProposalIDs} job={jobDraft} analysis={analysis} selectedFactIDs={selectedFactIDs} templateName={activeTemplate?.name ?? "Selected template"} busy={aiBusy} onProviderChange={changeAIProvider} onEndpointChange={(value) => { setOllamaEndpoint(value); setAIProviderStatus(null); }} onModelChange={changeAIModel} onAPIKeyChange={setGeminiAPIKey} onSaveCredential={saveGeminiCredential} onDeleteCredential={deleteGeminiCredential} onCheck={checkAIProvider} onGenerate={generateAITailoring} onCancel={cancelAITailoring} onProposalChange={updateReviewedProposal} onToggleProposal={toggleAcceptedProposal} onAccept={acceptAITailoring} onOpenJob={() => setView("job")} />}
            {view === "data" && <DataWorkspace profile={profile} experiences={experiences} projects={projects} educations={educations} jobs={jobs} applications={applications} busy={backupBusy} lastResult={lastBackupResult} onExport={exportBackup} onImport={importBackup} />}
          </div>

          <ResumePreview profile={profile} experiences={experiences} projects={selectedProjects} educations={educations} certifications={certifications} achievements={achievements} compileResult={compileResult} />
        </main>
      </div>
    </div>
  );
}

function NavButton({ active, label, icon, badge, onClick }: { active: boolean; label: string; icon: IconName; badge?: string | number; onClick: () => void }) {
  return (
    <button className={`nav-button ${active ? "active" : ""}`} aria-current={active ? "page" : undefined} onClick={onClick}>
      <Icon name={icon} />
      <span className="nav-button-label">{label}</span>
      {badge !== undefined && <span className="nav-badge">{badge}</span>}
    </button>
  );
}

type IconName = "home" | "user" | "briefcase" | "folder" | "education" | "sparkles" | "template" | "code" | "target" | "chat" | "download" | "refresh" | "check" | "search" | "database";

const iconPaths: Record<IconName, React.ReactNode> = {
  home: <><path d="m3 10 9-7 9 7" /><path d="M5 9v11h14V9" /><path d="M9 20v-6h6v6" /></>,
  user: <><circle cx="12" cy="8" r="4" /><path d="M4 21c.8-4.2 3.5-6 8-6s7.2 1.8 8 6" /></>,
  briefcase: <><rect x="3" y="7" width="18" height="13" rx="2" /><path d="M8 7V4h8v3M3 12h18M10 12v2h4v-2" /></>,
  folder: <><path d="M3 6h7l2 2h9v11H3z" /><path d="M3 10h18" /></>,
  education: <><path d="m2 9 10-5 10 5-10 5z" /><path d="M6 11v5c3 2 9 2 12 0v-5M22 9v7" /></>,
  sparkles: <><path d="m12 2 1.5 5.5L19 9l-5.5 1.5L12 16l-1.5-5.5L5 9l5.5-1.5z" /><path d="m19 15 .8 2.2L22 18l-2.2.8L19 21l-.8-2.2L16 18l2.2-.8z" /></>,
  template: <><rect x="4" y="3" width="16" height="18" rx="2" /><path d="M8 7h8M8 11h8M8 15h5" /></>,
  code: <><path d="m8 5-6 7 6 7M16 5l6 7-6 7M14 3l-4 18" /></>,
  target: <><circle cx="12" cy="12" r="9" /><circle cx="12" cy="12" r="4" /><path d="M12 3v3M21 12h-3M12 21v-3M3 12h3" /></>,
  chat: <><path d="M4 4h16v13H8l-4 4z" /><path d="M8 9h8M8 13h5" /></>,
  download: <><path d="M12 3v12M7 10l5 5 5-5M4 21h16" /></>,
  refresh: <><path d="M20 7V3h-4M4 17v4h4" /><path d="M19 10a7 7 0 0 0-12-4L4 9M5 14a7 7 0 0 0 12 4l3-3" /></>,
  check: <path d="m5 12 4 4L19 6" />,
  search: <><circle cx="10" cy="10" r="6" /><path d="m15 15 5 5" /></>,
  database: <><ellipse cx="12" cy="5" rx="8" ry="3" /><path d="M4 5v6c0 1.7 3.6 3 8 3s8-1.3 8-3V5M4 11v6c0 1.7 3.6 3 8 3s8-1.3 8-3v-6" /></>,
};

function Icon({ name, size = 18 }: { name: IconName; size?: number }) {
  return <svg className="icon" width={size} height={size} viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true">{iconPaths[name]}</svg>;
}

function TopToolbar({ profile, templateName, compiling, canExport, onCompile, onExport }: { profile: Profile; templateName: string; compiling: boolean; canExport: boolean; onCompile: () => void; onExport: () => void }) {
  return <header className="top-toolbar">
    <div className="toolbar-brand"><span className="brand-mark">T</span><strong>TailorCV</strong></div>
    <span className="toolbar-divider" />
    <div className="document-title"><strong>{profile.headline || "Backend Engineer Resume"}</strong><small>{templateName} · profile saved locally</small></div>
    <div className="toolbar-spacer" />
    <div className="ats-pill"><span className={`status-dot ${canExport ? "" : "muted"}`} /> {canExport ? "PDF ready" : "Source draft"}</div>
    <button className="toolbar-button" disabled={compiling} onClick={onCompile}><Icon name="refresh" size={16} />{compiling ? "Compiling…" : "Compile"}</button>
    <button className="export-button" disabled={!canExport || compiling} title={canExport ? "Export the latest compiled PDF" : "Compile the current source before exporting"} onClick={onExport}><Icon name="download" size={16} />Export PDF</button>
  </header>;
}

function PanelHeader({ eyebrow, title, description, action }: { eyebrow: string; title: string; description: string; action?: React.ReactNode }) {
  return <header className="panel-header"><div><p className="eyebrow">{eyebrow}</p><h1>{title}</h1><p>{description}</p></div>{action}</header>;
}

function WorkspaceOverview({ profile, completion, experiences, projects, applications, onOpen }: { profile: Profile; completion: number; experiences: ExperienceDraft[]; projects: ProjectDraft[]; applications: Application[]; onOpen: (view: View) => void }) {
  const versionCount = applications.reduce((sum, application) => sum + application.versions.length, 0);
  return <section className="workspace-panel scroll-panel">
    <PanelHeader eyebrow="Your workspace" title={profile.name ? `Welcome back, ${profile.name.split(" ")[0]}.` : "Build your source of truth."} description="Keep one reliable career profile, then tailor the evidence for each role." action={<button className="primary-button" onClick={() => onOpen("job")}>Match a job</button>} />
    <div className="metric-grid">
      <article className="metric-card accent"><span>Profile readiness</span><strong>{completion}%</strong><div className="progress"><i style={{ width: `${completion}%` }} /></div><button onClick={() => onOpen("profile")}>Complete profile →</button></article>
      <article className="metric-card"><span>Evidence</span><strong>{experiences.reduce((sum, item) => sum + item.bullets.length, 0)}</strong><small>reviewable career claims</small></article>
      <article className="metric-card"><span>Resume versions</span><strong>{versionCount}</strong><small>{applications.length} saved applications · {projects.filter((project) => project.resumeEligible).length} eligible projects</small></article>
    </div>
    <div className="overview-callout"><span className="callout-index">01</span><div><p className="eyebrow">Recommended next step</p><h2>{completion < 70 ? "Finish the profile foundation" : "Select evidence for your next application"}</h2><p>{completion < 70 ? "Add your identity, positioning, and core skills before tailoring." : "Choose projects and experience that directly support the target role."}</p></div><button className="secondary-button" onClick={() => onOpen(completion < 70 ? "profile" : "projects")}>Open workspace</button></div>
  </section>;
}

function ProfileWorkspace({ profile, busy, message, onChange, onSubmit }: { profile: Profile; busy: boolean; message: string; onChange: <K extends keyof Profile>(field:K,value:Profile[K])=>void; onSubmit: (event: FormEvent) => void }) {
  return <section className="workspace-panel scroll-panel">
    <PanelHeader eyebrow="Career profile" title="Profile" description="Identity and positioning used across every resume." />
    <form className="compact-form" onSubmit={onSubmit}>
      <FormBlock title="Identity" description="Shown in the resume header."><div className="field-grid two"><Field label="Full name" value={profile.name} onChange={(value) => onChange("name", value)} placeholder="Ada Lovelace" /><Field label="Headline" value={profile.headline} onChange={(value) => onChange("headline", value)} placeholder="Backend engineer" /><Field label="Email" type="email" value={profile.email} onChange={(value) => onChange("email", value)} placeholder="ada@example.com" /><Field label="Phone" value={profile.phone} onChange={(value) => onChange("phone", value)} placeholder="+91 98765 43210" /><Field label="Location" value={profile.location} onChange={(value) => onChange("location", value)} placeholder="Bengaluru, India" /><Field label="Website" type="url" value={profile.website} onChange={(value) => onChange("website", value)} placeholder="https://example.com" /></div></FormBlock>
      <FormBlock title="Presence" description="Professional links and repository identity."><div className="field-grid two"><Field label="GitHub username" value={profile.githubUsername} onChange={(value) => onChange("githubUsername", value)} placeholder="octocat" prefix="github.com/" /><Field label="LinkedIn URL" type="url" value={profile.linkedInUrl} onChange={(value) => onChange("linkedInUrl", value)} placeholder="https://linkedin.com/in/..." /></div><div className="contact-link-list">{profile.contactLinks.map((link,index)=><div className="field-grid two" key={link.id||index}><Field label="Link label" value={link.label} onChange={value=>onChange("contactLinks",profile.contactLinks.map((item,itemIndex)=>itemIndex===index?{...item,label:value}:item))} placeholder="Portfolio"/><Field label="URL" type="url" value={link.url} onChange={value=>onChange("contactLinks",profile.contactLinks.map((item,itemIndex)=>itemIndex===index?{...item,url:value}:item))} placeholder="https://example.com"/><button type="button" className="remove-button" onClick={()=>onChange("contactLinks",profile.contactLinks.filter((_,itemIndex)=>itemIndex!==index))}>Remove link</button></div>)}<button type="button" className="text-button" onClick={()=>onChange("contactLinks",[...profile.contactLinks,{id:"",label:"",url:""}])}>+ Add professional link</button></div></FormBlock>
      <FormBlock title="Positioning" description="A factual summary, never generated without evidence."><label className="field"><span>Professional summary</span><textarea rows={7} maxLength={2400} value={profile.summary} onChange={(event) => onChange("summary", event.target.value)} placeholder="Summarize your experience, strengths, and outcomes." /><small>{profile.summary.length}/2400</small></label></FormBlock>
      <div className="sticky-form-actions"><span>{message}</span><button className="primary-button" disabled={busy}>{busy ? "Saving…" : "Save profile"}</button></div>
    </form>
  </section>;
}

function FormBlock({ title, description, children }: { title: string; description: string; children: React.ReactNode }) {
  return <section className="form-block"><header><h2>{title}</h2><p>{description}</p></header><div>{children}</div></section>;
}

function ProjectWorkspace({ projects, selectedKeys, busyKey, readmeBusyKey, aiReady, githubUsername, githubBusy, onToggle, onAdd, onUpdate, onSave, onDelete, onGenerateReadmeBullets, onImportGitHubProject, onSyncGitHub, onOpenProfile }: { projects: ProjectDraft[]; selectedKeys: string[]; busyKey: string; readmeBusyKey: string; aiReady: boolean; githubUsername: string; githubBusy: boolean; onToggle: (key: string) => void; onAdd: () => void; onUpdate: (key: string, project: ProjectDraft) => void; onSave: (event: FormEvent, project: ProjectDraft) => void; onDelete: (project: ProjectDraft) => void; onGenerateReadmeBullets: (project: ProjectDraft) => void; onImportGitHubProject: (repositoryURL: string) => void; onSyncGitHub: () => void; onOpenProfile: () => void }) {
  const [tab, setTab] = useState<"select" | "manage">("select");
  const [query, setQuery] = useState("");
  const [repositoryURL, setRepositoryURL] = useState("");
  const [reviewKey, setReviewKey] = useState("");
  const filteredProjects = filterProjects(projects, query);
  useEffect(() => {
    if (tab !== "manage" || !reviewKey) return;
    window.requestAnimationFrame(() => document.getElementById(`project-review-${reviewKey}`)?.scrollIntoView({ behavior: "smooth", block: "start" }));
  }, [tab, reviewKey]);
  const openReview = (key: string) => {
    setReviewKey(key);
    setTab("manage");
  };
  return <section className="workspace-panel scroll-panel">
    <PanelHeader eyebrow="Selected work" title="Projects" description="Choose the strongest evidence for this resume." action={<div className="panel-actions"><button className="secondary-button" onClick={githubUsername ? onSyncGitHub : onOpenProfile} disabled={githubBusy}>{githubBusy ? "Syncing…" : githubUsername ? "Sync GitHub" : "Connect GitHub"}</button><button className="secondary-button" onClick={() => { onAdd(); setTab("manage"); }}>Add project</button></div>} />
    <div className="panel-tabs" role="group" aria-label="Project workspace"><button aria-pressed={tab === "select"} className={tab === "select" ? "active" : ""} onClick={() => setTab("select")}>Select for resume <span>{selectedKeys.length}</span></button><button aria-pressed={tab === "manage"} className={tab === "manage" ? "active" : ""} onClick={() => setTab("manage")}>Manage evidence</button></div>
    {tab === "select" ? <div className="project-selector">
      <div className="github-sync-note"><span className="github-mark">GH</span><div><strong>{githubUsername ? `github.com/${githubUsername}` : "Connect your GitHub profile"}</strong><p>{githubUsername ? "Public, owned repositories sync into Manage evidence for review." : "Add a username in Profile to import your public repositories."}</p></div><button onClick={githubUsername ? onSyncGitHub : onOpenProfile} disabled={githubBusy}>{githubBusy ? "Syncing…" : githubUsername ? "Refresh" : "Open profile"}</button></div>
      <form className="github-import-form" onSubmit={(event) => { event.preventDefault(); onImportGitHubProject(repositoryURL); }}><input aria-label="Public GitHub repository URL" type="url" required value={repositoryURL} onChange={(event) => setRepositoryURL(event.target.value)} placeholder="https://github.com/owner/repository" /><button className="secondary-button" disabled={githubBusy || !repositoryURL.trim()}>{githubBusy ? "Importing…" : "Import repository"}</button></form>
      <label className="search-field"><Icon name="search" size={16} /><input aria-label="Search projects" placeholder="Search projects, roles, or skills" value={query} onChange={(event) => setQuery(event.target.value)} /></label>
      <div className="selection-note"><span className="status-dot" /><div><strong>{selectedKeys.length} projects selected</strong><p>The preview updates as you select evidence.</p></div></div>
      {projects.length === 0 ? <div className="panel-empty"><span className="empty-icon"><Icon name="folder" size={22} /></span><strong>No projects yet</strong><p>Add a project with evidence bullets, technologies, and a review state.</p><button className="primary-button" onClick={() => { onAdd(); setTab("manage"); }}>Add first project</button></div> : filteredProjects.length === 0 ? <div className="panel-empty search-empty"><span className="empty-icon"><Icon name="search" size={22} /></span><strong>No matching projects</strong><p>Try a project name, role, description, or technology.</p><button className="text-button" onClick={() => setQuery("")}>Clear search</button></div> : filteredProjects.map((project) => {
        const selected = selectedKeys.includes(project.key);
        const selectable = isProjectSelectable(project);
        return <article className={`project-select-card ${selected ? "selected" : ""} ${!selectable ? "locked" : ""}`} key={project.key}>
          <button className={`project-check ${!selectable ? "review" : ""}`} aria-label={selectable ? `${selected ? "Remove" : "Add"} ${project.name || "project"} ${selected ? "from" : "to"} resume` : `Review ${project.name || "project"} before selection`} onClick={() => selectable ? onToggle(project.key) : openReview(project.key)}>{selected ? <Icon name="check" size={14} /> : !selectable ? "!" : null}</button>
          <div className="project-card-copy"><div className="project-title-row"><strong>{project.name || "Untitled project"}</strong><span>{selectable ? selected ? "Selected" : "Available" : "Review required"}</span></div><p>{project.description || "Add a concise description of the problem, your contribution, and the outcome."}</p><div className="tag-row">{project.skills.slice(0, 4).map((skill) => <span key={skill}>{skill}</span>)}{project.skills.length === 0 && <span>Skills not added</span>}</div><small>{project.verification === "verified" ? "Verified evidence" : "Needs review"} · {project.bullets.length} bullets</small>{!selectable && <button className="review-project-link" onClick={() => openReview(project.key)}>Review details and enable selection →</button>}</div>
        </article>;
      })}
    </div> : <ProjectSection projects={projects} busyKey={busyKey} readmeBusyKey={readmeBusyKey} aiReady={aiReady} onAdd={onAdd} onUpdate={onUpdate} onSave={onSave} onDelete={onDelete} onGenerateReadmeBullets={onGenerateReadmeBullets} />}
  </section>;
}

function EducationWorkspace({ educations, busyKey, onAdd, onUpdate, onSave, onDelete }: { educations: EducationDraft[]; busyKey: string; onAdd: () => void; onUpdate: (key: string, education: EducationDraft) => void; onSave: (event: FormEvent, education: EducationDraft) => void; onDelete: (education: EducationDraft) => void }) {
  return <section className="workspace-panel scroll-panel education-workspace"><PanelHeader eyebrow="Background" title="Education" description="Academic history that supports your application." action={<button className="secondary-button" onClick={onAdd}>Add education</button>} /><div className="education-list">{educations.length === 0 ? <div className="panel-empty tall"><span className="empty-icon"><Icon name="education" size={24} /></span><strong>No education records yet</strong><p>Add a degree, program, or current course of study. Saved records appear in the resume preview immediately.</p><button className="primary-button" onClick={onAdd}>Add education</button></div> : educations.map((education) => <EducationCard key={education.key} education={education} busy={busyKey === education.key} onUpdate={(next) => onUpdate(education.key, next)} onSave={(event) => onSave(event, education)} onDelete={() => onDelete(education)} />)}</div></section>;
}

function EducationCard({ education, busy, onUpdate, onSave, onDelete }: { education: EducationDraft; busy: boolean; onUpdate: (education: EducationDraft) => void; onSave: (event: FormEvent) => void; onDelete: () => void }) {
  const updateField = <K extends keyof EducationDraft>(field: K, value: EducationDraft[K]) => onUpdate({ ...education, [field]: value });
  return <form className="education-card" onSubmit={onSave}><header><div><span>{education.id ? "Saved education" : "New education"}</span><strong>{education.degree || "Untitled degree"}{education.institution ? ` · ${education.institution}` : ""}</strong></div><button className="danger-button" type="button" disabled={busy} onClick={onDelete}>Delete</button></header><div className="education-fields"><div className="field-grid two"><Field label="Institution" value={education.institution} onChange={(value) => updateField("institution", value)} placeholder="Example Institute" required /><Field label="Degree" value={education.degree} onChange={(value) => updateField("degree", value)} placeholder="Bachelor of Science" required /><Field label="Field of study" value={education.fieldOfStudy} onChange={(value) => updateField("fieldOfStudy", value)} placeholder="Computer Science" /><Field label="Location" value={education.location} onChange={(value) => updateField("location", value)} placeholder="Bengaluru, India" /><Field label="Start month" type="month" value={education.startDate} onChange={(value) => updateField("startDate", value)} placeholder="YYYY-MM" /><Field label="End month" type="month" value={education.endDate} onChange={(value) => updateField("endDate", value)} placeholder="YYYY-MM" disabled={education.current} /></div><label className="checkbox-field"><input type="checkbox" checked={education.current} onChange={(event) => updateField("current", event.target.checked)} /><span>I currently study here</span></label><label className="field"><span>Details (optional)</span><textarea rows={4} maxLength={1200} value={education.details} onChange={(event) => updateField("details", event.target.value)} placeholder="Honors, relevant coursework, thesis, or leadership." /><small>{education.details.length}/1200</small></label></div><footer><small>{education.updatedAt ? `Last saved ${new Date(education.updatedAt).toLocaleString()}` : "Not saved yet"}</small><button className="primary-button" disabled={busy}>{busy ? "Saving…" : "Save education"}</button></footer></form>;
}

function CredentialsWorkspace({certifications,achievements,busyKey,onCertifications,onAchievements,onSaveCertification,onDeleteCertification,onSaveAchievement,onDeleteAchievement}:{certifications:CertificationDraft[];achievements:AchievementDraft[];busyKey:string;onCertifications:React.Dispatch<React.SetStateAction<CertificationDraft[]>>;onAchievements:React.Dispatch<React.SetStateAction<AchievementDraft[]>>;onSaveCertification:(event:FormEvent,item:CertificationDraft)=>void;onDeleteCertification:(item:CertificationDraft)=>void;onSaveAchievement:(event:FormEvent,item:AchievementDraft)=>void;onDeleteAchievement:(item:AchievementDraft)=>void}){
  const updateCertification=(key:string,next:CertificationDraft)=>onCertifications(items=>items.map(item=>item.key===key?next:item));
  const updateAchievement=(key:string,next:AchievementDraft)=>onAchievements(items=>items.map(item=>item.key===key?next:item));
  return <section className="workspace-panel scroll-panel credentials-workspace"><PanelHeader eyebrow="Additional evidence" title="Credentials & achievements" description="Store defensible recognition and qualifications with source and review state." action={<div className="panel-actions"><button className="secondary-button" onClick={()=>onCertifications(items=>[...items,newCertificationDraft()])}>Add certification</button><button className="secondary-button" onClick={()=>onAchievements(items=>[...items,newAchievementDraft()])}>Add achievement</button></div>}/><div className="credential-grid"><section><h2>Certifications</h2>{certifications.length===0&&<p className="evidence-empty">No certifications saved.</p>}{certifications.map(item=><form className="education-card" key={item.key} onSubmit={event=>onSaveCertification(event,item)}><header><strong>{item.name||"New certification"}</strong><button type="button" className="danger-button" onClick={()=>onDeleteCertification(item)}>Delete</button></header><div className="field-grid two"><Field label="Certification" value={item.name} onChange={value=>updateCertification(item.key,{...item,name:value})} placeholder="Cloud Professional" required/><Field label="Issuer" value={item.issuer} onChange={value=>updateCertification(item.key,{...item,issuer:value})} placeholder="Example Institute" required/><Field label="Issued" type="month" value={item.issueDate} onChange={value=>updateCertification(item.key,{...item,issueDate:value})} placeholder="YYYY-MM"/><Field label="Expires" type="month" value={item.expiryDate} onChange={value=>updateCertification(item.key,{...item,expiryDate:value})} placeholder="YYYY-MM"/><Field label="Credential ID" value={item.credentialId} onChange={value=>updateCertification(item.key,{...item,credentialId:value})} placeholder="ABC-123"/><Field label="Credential URL" type="url" value={item.credentialUrl} onChange={value=>updateCertification(item.key,{...item,credentialUrl:value})} placeholder="https://..."/></div><label className="field"><span>Description</span><textarea rows={3} maxLength={1200} value={item.description} onChange={event=>updateCertification(item.key,{...item,description:event.target.value})}/></label><EvidenceState value={item.verification} onChange={verification=>updateCertification(item.key,{...item,verification})}/><button className="primary-button" disabled={busyKey===item.key}>{busyKey===item.key?"Saving…":"Save certification"}</button></form>)}</section><section><h2>Achievements</h2>{achievements.length===0&&<p className="evidence-empty">No achievements saved.</p>}{achievements.map(item=><form className="education-card" key={item.key} onSubmit={event=>onSaveAchievement(event,item)}><header><strong>{item.title||"New achievement"}</strong><button type="button" className="danger-button" onClick={()=>onDeleteAchievement(item)}>Delete</button></header><div className="field-grid two"><Field label="Title" value={item.title} onChange={value=>updateAchievement(item.key,{...item,title:value})} placeholder="Engineering award" required/><Field label="Date" type="month" value={item.date} onChange={value=>updateAchievement(item.key,{...item,date:value})} placeholder="YYYY-MM"/><Field label="Source URL" type="url" value={item.sourceUrl} onChange={value=>updateAchievement(item.key,{...item,sourceUrl:value})} placeholder="https://..."/></div><label className="field"><span>Evidence-backed description</span><textarea required rows={4} maxLength={1200} value={item.description} onChange={event=>updateAchievement(item.key,{...item,description:event.target.value})}/></label><EvidenceState value={item.verification} onChange={verification=>updateAchievement(item.key,{...item,verification})}/><button className="primary-button" disabled={busyKey===item.key}>{busyKey===item.key?"Saving…":"Save achievement"}</button></form>)}</section></div></section>;
}

function EvidenceState({value,onChange}:{value:"verified"|"unverified";onChange:(value:"verified"|"unverified")=>void}){return <label className="field"><span>Review state</span><select value={value} onChange={event=>onChange(event.target.value as "verified"|"unverified")}><option value="unverified">Needs review</option><option value="verified">Verified</option></select></label>}

function SkillsWorkspace({ skillsText, busy, message, onChange, onSubmit }: { skillsText: string; busy: boolean; message: string; onChange: (value: string) => void; onSubmit: (event: FormEvent) => void }) {
  const skills = parseSkills(skillsText);
  return <section className="workspace-panel scroll-panel"><PanelHeader eyebrow="Capabilities" title="Skills" description="A normalized vocabulary for matching and evidence selection." /><form className="compact-form skills-workspace" onSubmit={onSubmit}><FormBlock title="Skill inventory" description="Separate skills with commas. Duplicates are removed automatically."><label className="field"><span>Skills</span><textarea rows={6} value={skillsText} onChange={(event) => onChange(event.target.value)} placeholder="Go, TypeScript, PostgreSQL, Docker" /></label><div className="skill-chip-grid">{skills.map((skill) => <span key={skill}>{skill}<button type="button" aria-label={`Remove ${skill}`} onClick={() => onChange(skills.filter((item) => item !== skill).join(", "))}>×</button></span>)}</div></FormBlock><div className="sticky-form-actions"><span>{message}</span><button className="primary-button" disabled={busy}>{busy ? "Saving…" : "Save skills"}</button></div></form></section>;
}

function TemplatesWorkspace({ templates, selectedID, busyKey, onImport, onUse, onEdit, onDuplicate, onDelete }: { templates: domain.ResumeTemplate[]; selectedID: string; busyKey: string; onImport: () => void; onUse: (id: string) => void; onEdit: (id: string) => void; onDuplicate: (template: domain.ResumeTemplate) => void; onDelete: (template: domain.ResumeTemplate) => void }) {
  return <section className="workspace-panel scroll-panel templates-workspace">
    <PanelHeader eyebrow="Document system" title="Templates" description="Choose a built-in layout or import a complete, single-file LaTeX template." action={<button className="primary-button" disabled={busyKey !== ""} onClick={onImport}><Icon name="download" size={16} />{busyKey === "import" ? "Importing…" : "Import .tex"}</button>} />
    <div className="template-guidance"><span className="empty-icon"><Icon name="template" size={22} /></span><div><strong>Optional TailorCV markers</strong><p>Imported documents compile as-is. Add markers such as <code>{"{{TAILORCV_NAME}}"}</code>, <code>{"{{TAILORCV_EXPERIENCE_SECTION}}"}</code>, or <code>{"{{TAILORCV_PROJECTS_SECTION}}"}</code> to populate saved profile data when the template is selected.</p></div></div>
    <div className="template-grid">{templates.map((template) => {
      const selected = template.id === selectedID;
      const busy = busyKey === template.id;
      return <article className={`template-card ${selected ? "selected" : ""}`} key={template.id}>
        <div className="template-paper"><span>{template.name.split(/\s+/).slice(0, 2).map((part) => part[0]).join("")}</span><i /><i /><i /><i className="short" /></div>
        <div className="template-card-copy"><div className="template-title"><strong>{template.name}</strong><span>{template.builtIn ? "Built in" : "Custom"}</span></div><p>{template.description || "User-owned LaTeX resume template."}</p><small>{selected ? "Currently selected" : template.builtIn ? "Read-only · duplicate to customize" : "Stored locally"}</small></div>
        <div className="template-actions"><button className={selected ? "primary-button" : "secondary-button"} disabled={busy || selected} onClick={() => onUse(template.id)}>{selected ? "Selected" : busy ? "Loading…" : "Use"}</button><button className="text-button" disabled={busy} onClick={() => onEdit(template.id)}>Edit source</button><button className="text-button" disabled={busy} onClick={() => onDuplicate(template)}>Duplicate</button>{!template.builtIn && <button className="danger-button" disabled={busy} onClick={() => onDelete(template)}>Delete</button>}</div>
      </article>;
    })}</div>
  </section>;
}

function LatexWorkspace({ source, template, result, busy, compiling, versionBusy, hasOpenVersion, versionDirty, onChange, onSave, onSaveVersion, onReload, onCompile, onExport }: { source: string; template?: domain.ResumeTemplate; result: domain.CompileResult | null; busy: boolean; compiling: boolean; versionBusy: boolean; hasOpenVersion: boolean; versionDirty: boolean; onChange: (value: string) => void; onSave: () => void; onSaveVersion: () => void; onReload: () => void; onCompile: () => void; onExport: () => void }) {
  const diagnostics = result?.diagnostics ?? [];
  const compileState = compiling ? "Compiling…" : result?.success ? "PDF ready" : result ? "Needs attention" : "Draft";
  const [diagnosticLine, setDiagnosticLine] = useState(0);
  const focusLine = (line: number) => {
    setDiagnosticLine(line > 0 ? line : 0);
  };
  return <section className="workspace-panel latex-workspace"><PanelHeader eyebrow="Source editor" title="LaTeX" description={`Editing the rendered source from ${template?.name ?? "the selected template"}. Changes compile locally after a short pause.`} action={<div className="panel-actions latex-actions"><button className="text-button" disabled={busy || compiling || !template} onClick={onReload}>Reload data</button><button className="secondary-button" disabled={busy || compiling || !template} onClick={onSave}>{template?.builtIn ? "Save editable copy" : "Save template"}</button>{hasOpenVersion && <button className="secondary-button" disabled={versionBusy || compiling || !versionDirty} onClick={onSaveVersion}>{versionBusy ? "Saving…" : versionDirty ? "Save new version" : "Version saved"}</button>}<button className="secondary-button" disabled={compiling} onClick={onExport}>Export .tex</button><button className="primary-button" disabled={compiling} onClick={onCompile}>{compiling ? "Compiling…" : "Compile PDF"}</button></div>} /><div className="editor-tabs"><button className="active">resume.tex</button><span>{hasOpenVersion ? versionDirty ? "Saved version · edited draft" : "Immutable saved version" : template?.builtIn ? "Built-in source · edits are draft-only" : "Custom template · local"}</span></div><div className="latex-editor-body"><LatexCodeEditor value={source} focusLine={diagnosticLine} onChange={onChange} />{diagnostics.length > 0 && <section className="compiler-diagnostics" aria-label="Compiler diagnostics" aria-live="polite"><header><strong>Compiler diagnostics</strong><span>{diagnostics.length} issue{diagnostics.length === 1 ? "" : "s"}</span></header><div>{diagnostics.map((diagnostic, index) => <button type="button" className={`compiler-diagnostic ${diagnostic.severity}`} key={`${diagnostic.line}-${diagnostic.message}-${index}`} onClick={() => focusLine(diagnostic.line)}><span>{diagnostic.severity}</span><strong>{diagnostic.line > 0 ? `Line ${diagnostic.line}` : "Source"}</strong><p>{diagnostic.message}</p></button>)}</div></section>}</div><footer className="editor-status" role="status" aria-live="polite"><span>{compileState}</span><span>UTF-8</span><span>{template?.name ?? "LaTeX draft"}</span><span>{source.split("\n").length} lines</span></footer></section>;
}

function LatexCodeEditor({ value, focusLine, onChange }: { value: string; focusLine: number; onChange: (value: string) => void }) {
  const hostRef = useRef<HTMLDivElement>(null);
  const viewRef = useRef<EditorView | null>(null);
  const onChangeRef = useRef(onChange);
  onChangeRef.current = onChange;

  useEffect(() => {
    if (!hostRef.current) return;
    const view = new EditorView({
      parent: hostRef.current,
      state: EditorState.create({
        doc: value,
        extensions: [
          lineNumbers(), highlightSpecialChars(), history(), drawSelection(), highlightActiveLine(),
          syntaxHighlighting(defaultHighlightStyle), StreamLanguage.define(stex),
          keymap.of([...defaultKeymap, ...historyKeymap, indentWithTab]),
          EditorView.contentAttributes.of({ "aria-label": "LaTeX source editor" }),
          EditorView.lineWrapping,
          EditorView.theme({ "&": { height: "100%" }, ".cm-scroller": { overflow: "auto" } }),
          EditorView.updateListener.of((update) => {
            if (update.docChanged) onChangeRef.current(update.state.doc.toString());
          }),
        ],
      }),
    });
    viewRef.current = view;
    return () => { view.destroy(); viewRef.current = null; };
  }, []);

  useEffect(() => {
    const view = viewRef.current;
    if (!view || view.state.doc.toString() === value) return;
    view.dispatch({ changes: { from: 0, to: view.state.doc.length, insert: value } });
  }, [value]);

  useEffect(() => {
    const view = viewRef.current;
    if (!view || focusLine < 1 || focusLine > view.state.doc.lines) return;
    const line = view.state.doc.line(focusLine);
    view.dispatch({ selection: { anchor: line.from, head: line.to }, scrollIntoView: true });
    view.focus();
  }, [focusLine]);

  return <div className="code-editor" ref={hostRef} role="region" aria-label="LaTeX source" />;
}

function DataWorkspace({ profile, experiences, projects, educations, jobs, applications, busy, lastResult, onExport, onImport }: { profile: Profile; experiences: ExperienceDraft[]; projects: ProjectDraft[]; educations: EducationDraft[]; jobs: Job[]; applications: Application[]; busy: "export" | "import" | ""; lastResult: domain.BackupResult | null; onExport: () => void; onImport: () => void }) {
  const evidenceCount = experiences.reduce((sum, experience) => sum + experience.bullets.length, 0) + projects.reduce((sum, project) => sum + project.bullets.length, 0);
  const versionCount = applications.reduce((sum, application) => sum + application.versions.length, 0);
  return <section className="workspace-panel scroll-panel data-workspace"><PanelHeader eyebrow="Local data" title="Backup & restore" description="Keep a portable, versioned copy of your complete TailorCV profile." /><div className="backup-content"><section className="backup-summary"><header><span className="empty-icon"><Icon name="database" size={22} /></span><div><h2>Current local profile</h2><p>Everything listed here is included in one JSON backup.</p></div></header><div className="backup-stat-grid"><div><strong>{profile.name ? "1" : "0"}</strong><span>profile</span></div><div><strong>{experiences.length}</strong><span>roles</span></div><div><strong>{projects.length}</strong><span>projects</span></div><div><strong>{educations.length}</strong><span>education</span></div><div><strong>{jobs.length}</strong><span>jobs</span></div><div><strong>{applications.length}</strong><span>applications</span></div><div><strong>{versionCount}</strong><span>versions</span></div><div><strong>{profile.skills.length}</strong><span>skills</span></div><div><strong>{evidenceCount}</strong><span>evidence</span></div></div></section><section className="backup-action-card"><div><span className="backup-action-icon"><Icon name="download" size={20} /></span><h2>Export backup</h2><p>Write an owner-readable JSON snapshot using a native save dialog. IDs, ordering, verification state, ranking priority, and timestamps are preserved.</p></div><button className="primary-button" disabled={busy !== ""} onClick={onExport}>{busy === "export" ? "Exporting…" : "Choose destination"}</button></section><section className="backup-action-card restore"><div><span className="backup-action-icon"><Icon name="refresh" size={20} /></span><h2>Restore backup</h2><p>The entire file is validated before a single transaction replaces current data. Invalid or unsupported backups leave this profile untouched.</p></div><button className="secondary-button" disabled={busy !== ""} onClick={onImport}>{busy === "import" ? "Restoring…" : "Choose backup"}</button></section>{lastResult && <section className="backup-result"><span className="status-dot" /><div><strong>Last operation completed</strong><p>{lastResult.certificationCount} certifications · {lastResult.achievementCount} achievements · {lastResult.resumeVersionCount} resume versions</p><small>{lastResult.path}</small></div></section>}<div className="backup-safety"><strong>Backup format v6</strong><p>Backups include evidence ranking priorities, professional links, certifications, achievements, GitHub metadata, templates, AI audit history, and non-secret preferences, but never provider credentials, generated PDFs, compiler caches, or local model data.</p></div></div></section>;
}

function ResumePreview({ profile, experiences, projects, educations,certifications,achievements, compileResult }: { profile: Profile; experiences: ExperienceDraft[]; projects: ProjectDraft[]; educations: EducationDraft[];certifications:CertificationDraft[];achievements:AchievementDraft[]; compileResult: domain.CompileResult | null }) {
  const visibleExperiences = experiences.slice(0, 2);
  const visibleProjects = projects.slice(0, 3);
  const compiled = Boolean(compileResult?.success && compileResult.pdfBase64);
  const previewTitle = compiled ? "Compiled PDF" : compileResult ? "Compilation failed" : "Layout preview";
  const previewDetail = compiled ? `${compileResult?.engine} · ${compileResult?.durationMs} ms` : compileResult ? `${compileResult.diagnostics?.length || 1} compiler issue${compileResult.diagnostics?.length === 1 ? "" : "s"}` : "Draft · compile to verify";
  return <aside className="preview-pane" aria-label="Resume preview"><header className="preview-toolbar"><div role="status" aria-live="polite"><strong>{previewTitle}</strong><small>{previewDetail}</small></div>{!compiled && <div className="preview-controls"><span>{compileResult ? "Error" : "Draft"}</span></div>}</header>{compiled ? <PDFPreview data={compileResult?.pdfBase64 ?? ""} /> : <div className="preview-stage"><article className="resume-paper">
    <header className="resume-header"><h1>{profile.name || "Your Name"}</h1><p>{profile.headline || "Backend Engineer"}</p><small>{[profile.email || "your@email.com", profile.phone, profile.location, profile.website].filter(Boolean).join("  ·  ")}</small></header>
    {profile.summary && <ResumeSection title="Summary"><p className="resume-summary">{profile.summary}</p></ResumeSection>}
    <ResumeSection title="Experience">{visibleExperiences.length ? visibleExperiences.map((experience) => <div className="resume-entry" key={experience.key}><div><strong>{experience.title || "Role"} · {experience.company || "Company"}</strong><span>{experience.startDate} — {experience.current ? "Present" : experience.endDate}</span></div>{experience.location && <em>{experience.location}</em>}<ul>{experience.bullets.slice(0, 3).map((bullet, index) => <li key={bullet.id || index}>{bullet.text}</li>)}</ul></div>) : <ResumePlaceholder text="Add experience and evidence bullets" />}</ResumeSection>
    <ResumeSection title="Projects">{visibleProjects.length ? visibleProjects.map((project) => <div className="resume-entry project" key={project.key}><div><strong>{project.name || "Project"}</strong><span>{project.skills.slice(0, 3).join(" · ")}</span></div>{project.description && <p>{project.description}</p>}<ul>{project.bullets.slice(0, 2).map((bullet, index) => <li key={bullet.id || index}>{bullet.text}</li>)}</ul></div>) : <ResumePlaceholder text="Select projects to include them" />}</ResumeSection>
    {educations.length > 0 && <ResumeSection title="Education">{educations.slice(0, 2).map((education) => <div className="resume-entry education" key={education.key}><div><strong>{education.degree}{education.fieldOfStudy ? `, ${education.fieldOfStudy}` : ""}</strong><span>{education.startDate}{education.startDate && (education.current || education.endDate) ? " — " : ""}{education.current ? "Present" : education.endDate}</span></div><em>{education.institution}{education.location ? ` · ${education.location}` : ""}</em>{education.details && <p>{education.details}</p>}</div>)}</ResumeSection>}
    {certifications.length>0&&<ResumeSection title="Certifications">{certifications.slice(0,3).map(item=><div className="resume-entry" key={item.key}><div><strong>{item.name} · {item.issuer}</strong><span>{item.issueDate}</span></div>{item.description&&<p>{item.description}</p>}</div>)}</ResumeSection>}
    {achievements.length>0&&<ResumeSection title="Achievements"><ul>{achievements.slice(0,3).map(item=><li key={item.key}><strong>{item.title}:</strong> {item.description}</li>)}</ul></ResumeSection>}
    <ResumeSection title="Skills"><p className="resume-skills">{profile.skills.length ? profile.skills.join("  ·  ") : "Add skills to your profile"}</p></ResumeSection>
  </article></div>}</aside>;
}

function PDFPreview({ data }: { data: string }) {
  const canvasRef = useRef<HTMLCanvasElement>(null);
  const [zoom, setZoom] = useState(1);
  const [pageCount, setPageCount] = useState(0);
  const [renderError, setRenderError] = useState("");

  useEffect(() => {
    let cancelled = false;
    const bytes = Uint8Array.from(atob(data), (character) => character.charCodeAt(0));
    let loadingTask: { promise: Promise<any>; destroy: () => Promise<void> } | undefined;
    let renderTask: { cancel: () => void; promise: Promise<unknown> } | undefined;
    import("pdfjs-dist").then(({ getDocument, GlobalWorkerOptions }) => {
      GlobalWorkerOptions.workerSrc = pdfWorkerURL;
      loadingTask = getDocument({ data: bytes });
      return loadingTask.promise;
    }).then(async (document) => {
        if (cancelled) return;
        setPageCount(document.numPages);
        const page = await document.getPage(1);
        if (cancelled || !canvasRef.current) return;
        const viewport = page.getViewport({ scale: zoom * 1.35 });
        const pixelRatio = window.devicePixelRatio || 1;
        const canvas = canvasRef.current;
        canvas.width = Math.floor(viewport.width * pixelRatio);
        canvas.height = Math.floor(viewport.height * pixelRatio);
        canvas.style.width = `${viewport.width}px`;
        canvas.style.height = `${viewport.height}px`;
        const context = canvas.getContext("2d");
        if (!context) throw new Error("Canvas rendering is unavailable");
        const activeRenderTask = page.render({ canvasContext: context, viewport, transform: pixelRatio === 1 ? undefined : [pixelRatio, 0, 0, pixelRatio, 0, 0] });
        renderTask = activeRenderTask;
        await activeRenderTask.promise;
        setRenderError("");
      }).catch((reason) => {
      if (!cancelled && reason?.name !== "RenderingCancelledException") setRenderError(errorMessage(reason));
    });
    return () => {
      cancelled = true;
      renderTask?.cancel();
      if (loadingTask) void loadingTask.destroy();
    };
  }, [data, zoom]);

  return <div className="pdf-preview"><div className="pdf-controls" role="group" aria-label="PDF preview controls"><button aria-label="Zoom out" disabled={zoom <= 0.7} onClick={() => setZoom((value) => Math.max(0.7, value - 0.15))}>−</button><output aria-live="polite">{Math.round(zoom * 100)}%</output><button aria-label="Zoom in" disabled={zoom >= 1.6} onClick={() => setZoom((value) => Math.min(1.6, value + 0.15))}>+</button><i /><span>{pageCount || 1} page{pageCount === 1 ? "" : "s"}</span></div><div className="preview-stage compiled">{renderError ? <div className="pdf-error" role="alert">{renderError}</div> : <canvas ref={canvasRef} aria-label="Compiled resume PDF page 1" />}</div></div>;
}

function ResumeSection({ title, children }: { title: string; children: React.ReactNode }) {
  return <section className="resume-section"><h2>{title}</h2>{children}</section>;
}

function ResumePlaceholder({ text }: { text: string }) {
  return <p className="resume-placeholder">{text}</p>;
}

function Overview({ profile, completion, onProfile, onTailor }: { profile: Profile; completion: number; onProfile: () => void; onTailor: () => void }) {
  return (
    <section className="page overview-page">
      <header className="page-header">
        <div>
          <p className="eyebrow">Your workspace</p>
          <h1>{profile.name ? `Welcome back, ${profile.name.split(" ")[0]}.` : "Build the resume that fits."}</h1>
          <p>Keep one reliable career profile. Tailor the story for every opportunity.</p>
        </div>
        <button className="primary-button" onClick={onTailor}>Tailor a resume <span>→</span></button>
      </header>

      <div className="stat-grid">
        <article className="stat-card accent">
          <span className="stat-label">Profile readiness</span>
          <strong>{completion}%</strong>
          <div className="progress"><span style={{ width: `${completion}%` }} /></div>
          <button className="text-button" onClick={onProfile}>Complete your profile →</button>
        </article>
        <article className="stat-card">
          <span className="stat-label">Skills on record</span>
          <strong>{profile.skills.length}</strong>
          <p>Used for transparent job matching.</p>
        </article>
        <article className="stat-card">
          <span className="stat-label">Saved resumes</span>
          <strong>0</strong>
          <p>Every saved edit creates a recoverable immutable version.</p>
        </article>
      </div>

      <section className="next-step">
        <div className="step-number">01</div>
        <div>
          <p className="eyebrow">Recommended next step</p>
          <h2>{completion < 70 ? "Create your source of truth" : "Compare your first job"}</h2>
          <p>{completion < 70 ? "Add your headline, skills, links, and a factual professional summary. TailorCV will only generate from information you approve." : "Paste a job description to see which profile skills are explicitly requested."}</p>
        </div>
        <button className="secondary-button" onClick={completion < 70 ? onProfile : onTailor}>
          {completion < 70 ? "Open profile" : "Analyze a job"}
        </button>
      </section>
    </section>
  );
}

function ProfileEditor({
  profile,
  experiences,
  projects,
  skillsText,
  busy,
  experienceBusyKey,
  projectBusyKey,
  message,
  onChange,
  onSkillsChange,
  onSubmit,
  onAddExperience,
  onUpdateExperience,
  onSaveExperience,
  onDeleteExperience,
  onAddProject,
  onUpdateProject,
  onSaveProject,
  onDeleteProject,
}: {
  profile: Profile;
  experiences: ExperienceDraft[];
  projects: ProjectDraft[];
  skillsText: string;
  busy: boolean;
  experienceBusyKey: string;
  projectBusyKey: string;
  message: string;
  onChange: (field: keyof Profile, value: string) => void;
  onSkillsChange: (value: string) => void;
  onSubmit: (event: FormEvent) => void;
  onAddExperience: () => void;
  onUpdateExperience: (key: string, experience: ExperienceDraft) => void;
  onSaveExperience: (event: FormEvent, experience: ExperienceDraft) => void;
  onDeleteExperience: (experience: ExperienceDraft) => void;
  onAddProject: () => void;
  onUpdateProject: (key: string, project: ProjectDraft) => void;
  onSaveProject: (event: FormEvent, project: ProjectDraft) => void;
  onDeleteProject: (project: ProjectDraft) => void;
}) {
  return (
    <section className="page form-page">
      <header className="page-header compact">
        <div><p className="eyebrow">Career profile</p><h1>Your source of truth</h1><p>Only add information you would be comfortable defending in an interview.</p></div>
      </header>
      <form className="profile-form" onSubmit={onSubmit}>
        <FormSection title="Identity" description="The details shown at the top of your resume.">
          <div className="field-grid two">
            <Field label="Full name" value={profile.name} onChange={(value) => onChange("name", value)} placeholder="Ada Lovelace" />
            <Field label="Professional headline" value={profile.headline} onChange={(value) => onChange("headline", value)} placeholder="Backend engineer building reliable systems" />
            <Field label="Email" type="email" value={profile.email} onChange={(value) => onChange("email", value)} placeholder="ada@example.com" />
            <Field label="Phone" value={profile.phone} onChange={(value) => onChange("phone", value)} placeholder="+91 98765 43210" />
            <Field label="Location" value={profile.location} onChange={(value) => onChange("location", value)} placeholder="Bengaluru, India" />
            <Field label="Website" type="url" value={profile.website} onChange={(value) => onChange("website", value)} placeholder="https://example.com" />
          </div>
        </FormSection>
        <FormSection title="Professional presence" description="These links also help connect future repository imports.">
          <div className="field-grid two">
            <Field label="GitHub username" value={profile.githubUsername} onChange={(value) => onChange("githubUsername", value)} placeholder="octocat" prefix="github.com/" />
            <Field label="LinkedIn URL" type="url" value={profile.linkedInUrl} onChange={(value) => onChange("linkedInUrl", value)} placeholder="https://linkedin.com/in/..." />
          </div>
        </FormSection>
        <FormSection title="Positioning" description="Give the matcher strong, factual material to work with.">
          <label className="field full">
            <span>Professional summary</span>
            <textarea rows={5} maxLength={2400} value={profile.summary} onChange={(event) => onChange("summary", event.target.value)} placeholder="Summarize your experience, strengths, and the outcomes you create." />
            <small>{profile.summary.length}/2400</small>
          </label>
          <label className="field full">
            <span>Skills</span>
            <textarea rows={3} value={skillsText} onChange={(event) => onSkillsChange(event.target.value)} placeholder="Go, TypeScript, PostgreSQL, Docker, Kubernetes" />
            <small>Separate skills with commas. Duplicates are removed automatically.</small>
          </label>
        </FormSection>
        <div className="form-actions">
          <span className="success-message">{message}</span>
          <button className="primary-button" disabled={busy} type="submit">{busy ? "Saving…" : "Save profile"}</button>
        </div>
      </form>
      <ExperienceSection
        experiences={experiences}
        busyKey={experienceBusyKey}
        onAdd={onAddExperience}
        onUpdate={onUpdateExperience}
        onSave={onSaveExperience}
        onDelete={onDeleteExperience}
      />
    </section>
  );
}

function ExperienceSection({ experiences, busyKey, onAdd, onUpdate, onSave, onDelete }: {
  experiences: ExperienceDraft[];
  busyKey: string;
  onAdd: () => void;
  onUpdate: (key: string, experience: ExperienceDraft) => void;
  onSave: (event: FormEvent, experience: ExperienceDraft) => void;
  onDelete: (experience: ExperienceDraft) => void;
}) {
  return (
    <section className="experience-section">
      <div className="experience-heading">
        <div>
          <p className="eyebrow">Career evidence</p>
          <h2>Experience and outcomes</h2>
          <p>Record factual bullets with their origin and review state. Their order becomes the default resume order.</p>
        </div>
        <button className="secondary-button" type="button" onClick={onAdd}>Add experience</button>
      </div>
      {experiences.length === 0 ? (
        <div className="experience-empty">
          <strong>No experience entries yet.</strong>
          <p>Add a role and the concrete work or outcomes you can support.</p>
          <button className="primary-button" type="button" onClick={onAdd}>Add your first role</button>
        </div>
      ) : experiences.map((experience) => (
        <ExperienceCard
          key={experience.key}
          experience={experience}
          busy={busyKey === experience.key}
          onUpdate={(next) => onUpdate(experience.key, next)}
          onSave={(event) => onSave(event, experience)}
          onDelete={() => onDelete(experience)}
        />
      ))}
    </section>
  );
}

function ExperienceCard({ experience, busy, onUpdate, onSave, onDelete }: {
  experience: ExperienceDraft;
  busy: boolean;
  onUpdate: (experience: ExperienceDraft) => void;
  onSave: (event: FormEvent) => void;
  onDelete: () => void;
}) {
  const updateField = <K extends keyof ExperienceDraft>(field: K, value: ExperienceDraft[K]) => {
    onUpdate({ ...experience, [field]: value });
  };

  return (
    <form className="experience-card" onSubmit={onSave}>
      <div className="experience-card-heading">
        <div>
          <span>{experience.id ? "Saved role" : "New role"}</span>
          <strong>{experience.title || "Untitled role"}{experience.company ? ` · ${experience.company}` : ""}</strong>
        </div>
        <button className="danger-button" type="button" disabled={busy} onClick={onDelete}>Delete role</button>
      </div>
      <div className="field-grid three">
        <Field label="Company" value={experience.company} onChange={(value) => updateField("company", value)} placeholder="Example Systems" required />
        <Field label="Job title" value={experience.title} onChange={(value) => updateField("title", value)} placeholder="Senior Software Engineer" required />
        <Field label="Location" value={experience.location} onChange={(value) => updateField("location", value)} placeholder="Remote" />
        <Field label="Start month" type="month" value={experience.startDate} onChange={(value) => updateField("startDate", value)} placeholder="YYYY-MM" required />
        <Field label="End month" type="month" value={experience.endDate} onChange={(value) => updateField("endDate", value)} placeholder="YYYY-MM" disabled={experience.current} />
        <label className="checkbox-field">
          <input type="checkbox" checked={experience.current} onChange={(event) => updateField("current", event.target.checked)} />
          <span>I currently work here</span>
        </label>
      </div>

      <EvidenceEditor bullets={experience.bullets} emptyLabel="The role can still be saved." onChange={(bullets) => updateField("bullets", bullets)} />
      <div className="experience-actions">
        <small>{experience.updatedAt ? `Last saved ${new Date(experience.updatedAt).toLocaleString()}` : "Not saved yet"}</small>
        <button className="primary-button" disabled={busy} type="submit">{busy ? "Saving…" : "Save experience"}</button>
      </div>
    </form>
  );
}

function ProjectSection({ projects, busyKey, readmeBusyKey, aiReady, onAdd, onUpdate, onSave, onDelete, onGenerateReadmeBullets }: {
  projects: ProjectDraft[];
  busyKey: string;
  readmeBusyKey: string;
  aiReady: boolean;
  onAdd: () => void;
  onUpdate: (key: string, project: ProjectDraft) => void;
  onSave: (event: FormEvent, project: ProjectDraft) => void;
  onDelete: (project: ProjectDraft) => void;
  onGenerateReadmeBullets: (project: ProjectDraft) => void;
}) {
  return (
    <section className="experience-section project-section">
      <div className="experience-heading">
        <div>
          <p className="eyebrow">Selected work</p>
          <h2>Projects</h2>
          <p>Capture the work, technologies, and defensible outcomes that can support a tailored resume.</p>
        </div>
        <button className="secondary-button" type="button" onClick={onAdd}>Add project</button>
      </div>
      {projects.length === 0 ? (
        <div className="experience-empty">
          <strong>No projects yet.</strong>
          <p>Add a manual project now. GitHub imports will enter this same review workflow later.</p>
          <button className="primary-button" type="button" onClick={onAdd}>Add your first project</button>
        </div>
      ) : projects.map((project) => (
        <ProjectCard
          key={project.key}
          project={project}
          busy={busyKey === project.key}
          readmeBusy={readmeBusyKey === project.key}
          aiReady={aiReady}
          onUpdate={(next) => onUpdate(project.key, next)}
          onSave={(event) => onSave(event, project)}
          onDelete={() => onDelete(project)}
          onGenerateReadmeBullets={() => onGenerateReadmeBullets(project)}
        />
      ))}
    </section>
  );
}

function ProjectCard({ project, busy, readmeBusy, aiReady, onUpdate, onSave, onDelete, onGenerateReadmeBullets }: {
  project: ProjectDraft;
  busy: boolean;
  readmeBusy: boolean;
  aiReady: boolean;
  onUpdate: (project: ProjectDraft) => void;
  onSave: (event: FormEvent) => void;
  onDelete: () => void;
  onGenerateReadmeBullets: () => void;
}) {
  const updateField = <K extends keyof ProjectDraft>(field: K, value: ProjectDraft[K]) => {
    onUpdate({ ...project, [field]: value });
  };
  const selectedSkills = parseSkills(project.skillsText);
  const totalLanguageBytes = project.detectedLanguages.reduce((sum, language) => sum + language.bytes, 0);
  const approveGitHubProject = () => onUpdate({ ...project, verification: "verified", resumeEligible: true });

  return (
    <form className="experience-card project-card" id={`project-review-${project.key}`} onSubmit={onSave}>
      <div className="experience-card-heading">
        <div>
          <span>{project.id ? `${project.provenance} project` : "New project"}</span>
          <strong>{project.name || "Untitled project"}{project.role ? ` · ${project.role}` : ""}</strong>
        </div>
        <button className="danger-button" type="button" disabled={busy} onClick={onDelete}>Delete project</button>
      </div>
      <div className="project-fields">
        <div className="field-grid three">
          <Field label="Project name" value={project.name} onChange={(value) => updateField("name", value)} placeholder="Release Console" required />
          <Field label="Your role" value={project.role} onChange={(value) => updateField("role", value)} placeholder="Creator and maintainer" />
          <label className="checkbox-field eligible-field">
            <input type="checkbox" checked={project.resumeEligible} onChange={(event) => updateField("resumeEligible", event.target.checked)} />
            <span>Eligible for resumes</span>
          </label>
          <Field label="Start month" type="month" value={project.startDate} onChange={(value) => updateField("startDate", value)} placeholder="YYYY-MM" />
          <Field label="End month" type="month" value={project.endDate} onChange={(value) => updateField("endDate", value)} placeholder="YYYY-MM" disabled={project.ongoing} />
          <label className="checkbox-field">
            <input type="checkbox" checked={project.ongoing} onChange={(event) => updateField("ongoing", event.target.checked)} />
            <span>Ongoing project</span>
          </label>
        </div>
        <label className="field full">
          <span>Description</span>
          <textarea rows={4} maxLength={2400} value={project.description} onChange={(event) => updateField("description", event.target.value)} placeholder="What the project does, who it serves, and why it matters." />
          <small>{project.description.length}/2400</small>
        </label>
        <div className="field-grid two">
          <Field label="Project URL" type="url" value={project.url} onChange={(value) => updateField("url", value)} placeholder="https://example.com/project" />
          <Field label="Repository URL" type="url" value={project.repositoryUrl} onChange={(value) => updateField("repositoryUrl", value)} placeholder="https://github.com/owner/repository" />
        </div>
        {project.provenance === "github" && <section className="repository-metadata"><header><div><strong>GitHub repository snapshot</strong><p>Reference metadata from the latest successful public sync.</p></div><span>{project.repositoryVisibility || "public"}</span></header><div><small>Repository ID</small><strong>{project.repositoryId || "Unavailable"}</strong><small>Updated on GitHub</small><strong>{project.repositoryUpdatedAt ? new Date(project.repositoryUpdatedAt).toLocaleString() : "Unavailable"}</strong></div><details><summary>README snapshot{project.repositoryReadme ? "" : " unavailable"}</summary><pre>{project.repositoryReadme || "No README content was returned by GitHub."}</pre></details>{project.repositoryReadme && <button className="secondary-button" type="button" disabled={busy || readmeBusy || !aiReady} title={aiReady ? "Generate editable, unverified project bullets from this saved README" : "Check an AI provider and select a model first"} onClick={onGenerateReadmeBullets}>{readmeBusy ? "Generating README bullets…" : "Generate editable bullets from README"}</button>}</section>}
        <label className="field full">
          <span>Technologies and skills</span>
          <textarea rows={2} value={project.skillsText} onChange={(event) => updateField("skillsText", event.target.value)} placeholder="Go, React, SQLite, Docker" />
          <small>Separate skills with commas. Duplicates are removed when saved.</small>
        </label>
        {project.detectedLanguages.length > 0 && <section className="language-picker"><header><div><strong>Languages detected by GitHub</strong><p>Choose which languages should appear in this project's resume skills.</p></div><span>{project.detectedLanguages.length} detected</span></header><div>{project.detectedLanguages.map((language) => {
          const selected = selectedSkills.some((skill) => skill.toLocaleLowerCase() === language.name.toLocaleLowerCase());
          const percentage = totalLanguageBytes > 0 ? Math.max(0.1, language.bytes / totalLanguageBytes * 100) : 0;
          return <button className={selected ? "selected" : ""} type="button" aria-pressed={selected} key={language.name} onClick={() => updateField("skillsText", toggleProjectLanguage(project.skillsText, language.name))}><span>{selected && <Icon name="check" size={13} />}{language.name}</span>{totalLanguageBytes > 0 && <small>{percentage.toFixed(percentage >= 10 ? 0 : 1)}%</small>}</button>;
        })}</div></section>}
        <div className="review-fields">
          <label className="field">
            <span>Project review state</span>
            <select value={project.verification} onChange={(event) => updateField("verification", event.target.value as ProjectDraft["verification"])}>
              <option value="unverified">Needs review</option>
              <option value="verified">Verified</option>
            </select>
          </label>
          <div className="provenance-field"><span>Origin</span><strong>{project.provenance}</strong></div>
          {project.provenance === "github" && !isProjectSelectable(project) ? <div className="review-guidance"><p>Review the imported details and language choices, then approve this project.</p><button className="secondary-button" type="button" onClick={approveGitHubProject}>Approve for resume</button></div> : !project.resumeEligible && <p>Excluded from resume selection until you mark it eligible.</p>}
        </div>
      </div>
      <EvidenceEditor bullets={project.bullets} emptyLabel="The project can still be saved." onChange={(bullets) => updateField("bullets", bullets)} />
      <div className="experience-actions">
        <small>{project.updatedAt ? `Last saved ${new Date(project.updatedAt).toLocaleString()}` : "Not saved yet"}</small>
        <button className="primary-button" disabled={busy} type="submit">{busy ? "Saving…" : "Save project"}</button>
      </div>
    </form>
  );
}

function EvidenceEditor({ bullets, emptyLabel, onChange }: {
  bullets: EvidenceBullet[];
  emptyLabel: string;
  onChange: (bullets: EvidenceBullet[]) => void;
}) {
  const updateBullet = (index: number, patch: Partial<EvidenceBullet>) => {
    onChange(bullets.map((bullet, bulletIndex) => bulletIndex === index ? { ...bullet, ...patch } : bullet));
  };
  const removeBullet = (index: number) => {
    onChange(bullets.filter((_, bulletIndex) => bulletIndex !== index));
  };

  return (
    <>
      <div className="evidence-heading">
        <div><strong>Evidence bullets</strong><span>Use specific, defensible work and outcomes.</span></div>
        <button className="text-button" type="button" onClick={() => onChange([...bullets, newEvidenceBullet()])}>+ Add bullet</button>
      </div>
      <div className="evidence-list">
        {bullets.length === 0 && <p className="evidence-empty">No evidence bullets. {emptyLabel}</p>}
        {bullets.map((bullet, index) => (
          <div className="evidence-row" key={bullet.id || `new-bullet-${index}`}>
            <div className="evidence-order" aria-label={`Reorder evidence bullet ${index + 1}`}>
              <span>{index + 1}</span>
              <button type="button" aria-label="Move bullet up" disabled={index === 0} onClick={() => onChange(moveItem(bullets, index, index - 1))}>↑</button>
              <button type="button" aria-label="Move bullet down" disabled={index === bullets.length - 1} onClick={() => onChange(moveItem(bullets, index, index + 1))}>↓</button>
            </div>
            <div className="evidence-fields">
              <label className="field full">
                <span>Claim or outcome</span>
                <textarea rows={3} maxLength={1200} required value={bullet.text} onChange={(event) => updateBullet(index, { text: event.target.value })} placeholder="Reduced deployment time by 40% by replacing manual release steps with an audited pipeline." />
                <small>{bullet.text.length}/1200</small>
              </label>
              <div className="evidence-metadata">
                <label className="field">
                  <span>Review state</span>
                  <select value={bullet.verification} onChange={(event) => updateBullet(index, { verification: event.target.value as EvidenceBullet["verification"] })}>
                    <option value="unverified">Needs review</option>
                    <option value="verified">Verified</option>
                  </select>
                </label>
                <label className="field">
                  <span>Ranking priority</span>
                  <select value={bullet.importance} onChange={(event) => updateBullet(index, { importance: event.target.value as EvidenceBullet["importance"] })}>
                    <option value="standard">Standard</option>
                    <option value="important">Important (+6)</option>
                    <option value="essential">Essential (+12)</option>
                  </select>
                </label>
                <Field label="Source URL (optional)" type="url" value={bullet.sourceUrl} onChange={(value) => updateBullet(index, { sourceUrl: value })} placeholder="https://github.com/…" />
                <div className="provenance-field"><span>Origin</span><strong>{bullet.provenance}</strong></div>
                <button className="remove-button" type="button" onClick={() => removeBullet(index)}>Remove</button>
              </div>
            </div>
          </div>
        ))}
      </div>
    </>
  );
}

function FormSection({ title, description, children }: { title: string; description: string; children: React.ReactNode }) {
  return <section className="form-section"><div className="section-copy"><h2>{title}</h2><p>{description}</p></div><div className="section-fields">{children}</div></section>;
}

function Field({ label, value, placeholder, type = "text", prefix, disabled = false, required = false, onChange }: { label: string; value: string; placeholder: string; type?: string; prefix?: string; disabled?: boolean; required?: boolean; onChange: (value: string) => void }) {
  return <label className="field"><span>{label}</span><div className={prefix ? "prefixed-input" : ""}>{prefix && <em>{prefix}</em>}<input type={type} value={value} placeholder={placeholder} disabled={disabled} required={required} onChange={(event) => onChange(event.target.value)} /></div></label>;
}

function JobTailor({ job, jobs, application, analysis, selectedFactIDs, busy, versionBusy, statusBusy, hasEvidence, templateName, onChange, onNew, onOpen, onDelete, onToggleEvidence, onCreateVersion, onOpenVersion, onStatusChange, onSubmit, onProfile }: {
  job: Job;
  jobs: Job[];
  application?: Application;
  analysis: JobAnalysis | null;
  selectedFactIDs: string[];
  busy: boolean;
  versionBusy: boolean;
  statusBusy: boolean;
  hasEvidence: boolean;
  templateName: string;
  onChange: (field: "company" | "role" | "description", value: string) => void;
  onNew: () => void;
  onOpen: (job: Job) => void;
  onDelete: (job: Job) => void;
  onToggleEvidence: (factID: string) => void;
  onCreateVersion: () => void;
  onOpenVersion: (version: ResumeVersion) => void;
  onStatusChange: (status: ApplicationStatus) => void;
  onSubmit: (event: FormEvent) => void;
  onProfile: () => void;
}) {
  return (
    <section className="page form-page">
      <header className="page-header compact job-page-header"><div><p className="eyebrow">Job workspace</p><h1>Find the strongest fit</h1><p>Save opportunities and rank only the career evidence you control.</p></div><button className="secondary-button" onClick={onNew}>New job</button></header>
      {jobs.length > 0 && <section className="saved-jobs" aria-label="Saved jobs"><div className="saved-jobs-heading"><strong>Saved opportunities</strong><span>{jobs.length}</span></div><div className="saved-job-list">{jobs.map((saved) => <article className={saved.id === job.id ? "active" : ""} key={saved.id}><button className="saved-job-open" onClick={() => onOpen(saved)}><strong>{saved.role || "Untitled opportunity"}</strong><span>{saved.company || "Company not specified"}</span><small>Updated {new Date(saved.updatedAt).toLocaleDateString()}</small></button><button className="saved-job-delete" aria-label={`Delete ${saved.role || "saved job"}`} onClick={() => onDelete(saved)}>×</button></article>)}</div></section>}
      {!hasEvidence && <div className="empty-callout"><strong>Add career evidence for useful rankings.</strong><span>The job will still be saved, but matching improves with skills and factual bullets.</span><button onClick={onProfile}>Open skills</button></div>}
      <div className="tailor-grid">
        <form className="job-card" onSubmit={onSubmit}>
          <div className="card-heading"><span className="step-pill">Step 1</span><h2>{job.id ? "Update saved job" : "Describe the opportunity"}</h2></div>
          <div className="field-grid two job-identity"><Field label="Company" value={job.company} onChange={(value) => onChange("company", value)} placeholder="Example Systems" /><Field label="Role" value={job.role} onChange={(value) => onChange("role", value)} placeholder="Platform Engineer" /></div>
          <textarea aria-label="Job description" value={job.description} onChange={(event) => onChange("description", event.target.value)} placeholder="Paste the complete role description here…" rows={18} />
          <div className="job-actions"><small>{job.description.length.toLocaleString()} characters · saved locally when analyzed</small><button className="primary-button" disabled={busy} type="submit">{busy ? "Ranking…" : "Save & rank evidence"}</button></div>
        </form>
        <section className="analysis-card">
          <div className="card-heading"><span className="step-pill muted">Step 2</span><h2>Review the evidence</h2></div>
          {!analysis ? <div className="analysis-empty"><div className="radar">✦</div><strong>Your ranked evidence will appear here</strong><p>TailorCV shows skill, term, importance, and recency signals with a reason for every suggested fact.</p></div> : <AnalysisResult analysis={analysis} selectedFactIDs={selectedFactIDs} onToggle={onToggleEvidence} />}
          {analysis && <div className="version-action"><div><strong>{selectedFactIDs.length} facts selected</strong><span>Render with {templateName} and preserve an immutable job snapshot.</span></div><button className="primary-button" disabled={versionBusy || selectedFactIDs.length === 0} onClick={onCreateVersion}>{versionBusy ? "Saving version…" : "Save resume version"}</button></div>}
        </section>
      </div>
      {application && application.versions.length > 0 && <section className="version-history"><div className="application-history-heading"><div className="saved-jobs-heading"><strong>Resume history</strong><span>{application.versions.length}</span></div><ApplicationStatusControl status={application.status} busy={statusBusy} onChange={onStatusChange} /></div><div>{application.versions.map((version) => <article key={version.id}><div><strong>Version {version.versionNumber}</strong><span>{version.selectedFactIds.length} facts · {new Date(version.createdAt).toLocaleString()}</span><small>{version.pdfAvailable ? `PDF saved · ${version.compileEngine}` : version.compiledAt ? "Compilation needs attention" : "Not compiled as a saved artifact"}</small></div><button className="text-button" onClick={() => onOpenVersion(version)}>Open source</button></article>)}</div></section>}
    </section>
  );
}

function AnalysisResult({ analysis, selectedFactIDs, onToggle }: { analysis: JobAnalysis; selectedFactIDs: string[]; onToggle: (factID: string) => void }) {
  return <div className="analysis-result"><div className="score-row"><div className="score-ring" style={{ "--score": `${analysis.score * 3.6}deg` } as React.CSSProperties}><span>{analysis.score}%</span></div><div><strong>Profile skill alignment</strong><p>{analysis.explanation}</p></div></div><SkillGroup title="Required skills" skills={analysis.requiredSkills} tone="match" />{analysis.preferredSkills.length > 0 && <SkillGroup title="Preferred skills" skills={analysis.preferredSkills} tone="match" />}<SkillGroup title="Profile skills not mentioned" skills={analysis.unmentionedSkills} tone="neutral" />{analysis.responsibilities.length > 0 && <section className="requirement-summary"><div className="skill-heading"><h3>Responsibilities extracted</h3><span>{analysis.responsibilities.length}</span></div><ul>{analysis.responsibilities.map((responsibility) => <li key={responsibility}>{responsibility}</li>)}</ul></section>}<div className="keyword-summary"><strong>Search vocabulary</strong><div className="skill-list">{analysis.keywords.map((keyword) => <span className={`skill neutral`} key={keyword}>{keyword}</span>)}</div></div><div className="ranked-evidence"><div className="skill-heading"><h3>Ranked career evidence</h3><span>{analysis.rankedEvidence.length}</span></div>{analysis.rankedEvidence.length ? analysis.rankedEvidence.map((evidence) => {
    const selected = selectedFactIDs.includes(evidence.factId);
    return <article className={`${selected ? "selected" : ""} ${!evidence.selectable ? "locked" : ""}`} key={`${evidence.sourceType}-${evidence.factId}`}><header><label><input aria-label={`${selected ? "Remove" : "Add"} ${evidence.sourceLabel} evidence ${selected ? "from" : "to"} resume`} type="checkbox" checked={selected} disabled={!evidence.selectable} onChange={() => onToggle(evidence.factId)} /><div><span>{evidence.sourceType}</span><strong>{evidence.sourceLabel}</strong></div></label><b>{evidence.score}</b></header><p>{evidence.text}</p><div className="evidence-reasons">{evidence.reasons.map((reason) => <span key={reason}>{reason}</span>)}{!evidence.selectable && <span>Review project to unlock</span>}</div></article>;
  }) : <p className="no-ranked-evidence">No saved evidence shares meaningful skills or terms with this description yet.</p>}</div></div>;
}

function SkillGroup({ title, skills, tone }: { title: string; skills: string[]; tone: "match" | "neutral" }) {
  return <div className="skill-group"><div className="skill-heading"><h3>{title}</h3><span>{skills.length}</span></div><div className="skill-list">{skills.length ? skills.map((skill) => <span className={`skill ${tone}`} key={skill}>{skill}</span>) : <p>None yet.</p>}</div></div>;
}

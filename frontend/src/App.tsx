import { FormEvent, useEffect, useMemo, useState } from "react";
import {
  AnalyzeJobDescription,
  DeleteEducation,
  DeleteExperience,
  DeleteProject,
  ExportProfileBackup,
  GetProfile,
  ImportGitHubProjects,
  ImportProfileBackup,
  ListEducations,
  ListExperiences,
  ListProjects,
  SaveEducation,
  SaveExperience,
  SaveProject,
  SaveProfile,
} from "../wailsjs/go/main/App";
import { domain } from "../wailsjs/go/models";
import {
  Education,
  EducationDraft,
  educationToDraft,
  newEducationDraft,
  toEducationInput,
} from "./lib/education";
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
  newProjectDraft,
  projectToDraft,
  reconcileSelectedProjectKeys,
  removeSelectedProjectKey,
  toProjectInput,
} from "./lib/project";
import {
  emptyProfile,
  parseSkills,
  Profile,
  profileCompletion,
} from "./lib/profile";

type View = "overview" | "profile" | "experience" | "projects" | "education" | "skills" | "latex" | "job" | "ai" | "data";

type AIProvider = "Ollama" | "Gemini" | "Claude" | "OpenAI";

type ChatMessage = {
  role: "assistant" | "user";
  text: string;
};

type JobAnalysis = {
  score: number;
  matchedSkills: string[];
  unmentionedSkills: string[];
  explanation: string;
};

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

export default function App() {
  const [view, setView] = useState<View>("projects");
  const [profile, setProfile] = useState<Profile>(emptyProfile);
  const [experiences, setExperiences] = useState<ExperienceDraft[]>([]);
  const [projects, setProjects] = useState<ProjectDraft[]>([]);
  const [educations, setEducations] = useState<EducationDraft[]>([]);
  const [skillsText, setSkillsText] = useState("");
  const [jobDescription, setJobDescription] = useState("");
  const [analysis, setAnalysis] = useState<JobAnalysis | null>(null);
  const [busy, setBusy] = useState(true);
  const [experienceBusyKey, setExperienceBusyKey] = useState("");
  const [projectBusyKey, setProjectBusyKey] = useState("");
  const [educationBusyKey, setEducationBusyKey] = useState("");
  const [backupBusy, setBackupBusy] = useState<"export" | "import" | "">("");
  const [githubBusy, setGitHubBusy] = useState(false);
  const [lastBackupResult, setLastBackupResult] = useState<domain.BackupResult | null>(null);
  const [onboardingDismissed, setOnboardingDismissed] = useState(false);
  const [message, setMessage] = useState("");
  const [error, setError] = useState("");
  const [selectedProjectKeys, setSelectedProjectKeys] = useState<string[]>([]);
  const [latexSource, setLatexSource] = useState(DEFAULT_LATEX);
  const [aiProvider, setAIProvider] = useState<AIProvider>("Ollama");
  const [chatDraft, setChatDraft] = useState("");
  const [chatMessages, setChatMessages] = useState<ChatMessage[]>([
    { role: "assistant", text: "I can tighten bullets, compare evidence to a role, or explain what changed. I will not invent facts." },
  ]);

  const loadWorkspaceData = async () => {
    const [result, savedExperiences, savedProjects, savedEducations] = await Promise.all([GetProfile(), ListExperiences(), ListProjects(), ListEducations()]);
    const loaded = { ...emptyProfile, ...result } as Profile;
    loaded.skills ??= [];
    setProfile(loaded);
    setSkillsText(loaded.skills.join(", "));
    setExperiences((savedExperiences as unknown as Experience[]).map(experienceToDraft));
    const projectDrafts = (savedProjects as unknown as Project[]).map(projectToDraft);
    setProjects(projectDrafts);
    setSelectedProjectKeys(projectDrafts.filter((project) => project.resumeEligible).slice(0, 3).map((project) => project.key));
    setEducations((savedEducations as unknown as Education[]).map(educationToDraft));
  };

  useEffect(() => {
    loadWorkspaceData().catch((reason) => setError(errorMessage(reason))).finally(() => setBusy(false));
  }, []);

  const completion = useMemo(() => profileCompletion(profile), [profile]);

  const updateProfile = (field: keyof Profile, value: string) => {
    setProfile((current) => ({ ...current, [field]: value }));
    setMessage("");
  };

  const saveProfile = async (event: FormEvent) => {
    event.preventDefault();
    setBusy(true);
    setError("");
    setMessage("");
    try {
      const saved = await SaveProfile({
        ...profile,
        skills: parseSkills(skillsText),
      });
      const normalized = { ...emptyProfile, ...saved } as Profile;
      setProfile(normalized);
      setSkillsText(normalized.skills.join(", "));
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
      const result = await AnalyzeJobDescription({ description: jobDescription });
      setAnalysis(result as JobAnalysis);
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
    if (!next.resumeEligible) {
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
      setSelectedProjectKeys((current) => reconcileSelectedProjectKeys(current, draft.key, normalized.key, normalized.resumeEligible));
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

  const toggleProject = (key: string) => {
    if (!projects.some((project) => project.key === key && project.resumeEligible)) return;
    setSelectedProjectKeys((current) => current.includes(key) ? current.filter((item) => item !== key) : [...current, key]);
  };

  const sendChatMessage = (event: FormEvent) => {
    event.preventDefault();
    const text = chatDraft.trim();
    if (!text) return;
    setChatMessages((current) => [
      ...current,
      { role: "user", text },
      { role: "assistant", text: `${aiProvider} is not connected yet. This workspace is ready for the provider adapter; your message has not left this device.` },
    ]);
    setChatDraft("");
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
      setSelectedProjectKeys((current) => current.filter((key) => projectDrafts.some((project) => project.key === key && project.resumeEligible)));
      setMessage(`GitHub sync complete: ${result.imported} imported, ${result.updated} refreshed, ${result.skipped} skipped.`);
    } catch (reason) {
      setError(errorMessage(reason));
    } finally {
      setGitHubBusy(false);
    }
  };

  const selectedProjects = projects.filter((project) => selectedProjectKeys.includes(project.key));
  const hasProfileData = Boolean(profile.name || profile.email || profile.headline || profile.phone || profile.location || profile.website || profile.githubUsername || profile.linkedInUrl || profile.summary || profile.skills.length);
  const showOnboarding = !busy && !onboardingDismissed && !hasProfileData && experiences.length === 0 && projects.length === 0 && educations.length === 0;

  return (
    <div className="studio-shell">
      {showOnboarding && <Onboarding profile={profile} skillsText={skillsText} busy={busy} onChange={updateProfile} onSkillsChange={setSkillsText} onSubmit={saveProfile} onSkip={() => setOnboardingDismissed(true)} />}
      <TopToolbar profile={profile} />

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
            <NavButton active={view === "skills"} label="Skills" icon="sparkles" badge={profile.skills.length || undefined} onClick={() => setView("skills")} />
            <p className="nav-section-label nav-section-spaced">Tailor</p>
            <NavButton active={view === "latex"} label="LaTeX source" icon="code" onClick={() => setView("latex")} />
            <NavButton active={view === "job"} label="Job match" icon="target" badge={analysis ? `${analysis.score}%` : undefined} onClick={() => setView("job")} />
            <NavButton active={view === "ai"} label="AI assistant" icon="chat" badge="Setup" onClick={() => setView("ai")} />
            <p className="nav-section-label nav-section-spaced">System</p>
            <NavButton active={view === "data"} label="Backup & restore" icon="database" onClick={() => setView("data")} />
          </nav>

          <button className="provider-status" onClick={() => setView("ai")}>
            <span className="status-dot muted" />
            <span><strong>AI provider</strong><small>{aiProvider} · not connected</small></span>
            <span aria-hidden="true">›</span>
          </button>
        </aside>

        <main className="studio-workspace">
          {error && <div className="notice error" role="alert">{error}<button aria-label="Dismiss error" onClick={() => setError("")}>×</button></div>}
          {message && <div className="notice success" role="status">{message}<button aria-label="Dismiss message" onClick={() => setMessage("")}>×</button></div>}

          <div className="editor-pane">
            {view === "overview" && <WorkspaceOverview profile={profile} completion={completion} experiences={experiences} projects={projects} onOpen={setView} />}
            {view === "profile" && <ProfileWorkspace profile={profile} busy={busy} message={message} onChange={updateProfile} onSubmit={saveProfile} />}
            {view === "experience" && <section className="workspace-panel scroll-panel"><PanelHeader eyebrow="Career evidence" title="Experience" description="Keep every claim factual, ordered, and reviewable." action={<button className="secondary-button" onClick={addExperience}>Add role</button>} /><ExperienceSection experiences={experiences} busyKey={experienceBusyKey} onAdd={addExperience} onUpdate={updateExperience} onSave={saveExperience} onDelete={deleteExperience} /></section>}
            {view === "projects" && <ProjectWorkspace projects={projects} selectedKeys={selectedProjectKeys} busyKey={projectBusyKey} githubUsername={profile.githubUsername} githubBusy={githubBusy} onToggle={toggleProject} onAdd={addProject} onUpdate={updateProject} onSave={saveProject} onDelete={deleteProject} onSyncGitHub={syncGitHubProjects} onOpenProfile={() => setView("profile")} />}
            {view === "education" && <EducationWorkspace educations={educations} busyKey={educationBusyKey} onAdd={addEducation} onUpdate={updateEducation} onSave={saveEducation} onDelete={deleteEducation} />}
            {view === "skills" && <SkillsWorkspace skillsText={skillsText} busy={busy} message={message} onChange={setSkillsText} onSubmit={saveProfile} />}
            {view === "latex" && <LatexWorkspace source={latexSource} onChange={setLatexSource} />}
            {view === "job" && <JobTailor description={jobDescription} analysis={analysis} busy={busy} hasSkills={profile.skills.length > 0} onDescriptionChange={setJobDescription} onSubmit={analyzeJob} onProfile={() => setView("skills")} />}
            {view === "ai" && <AIWorkspace provider={aiProvider} draft={chatDraft} messages={chatMessages} onProviderChange={setAIProvider} onDraftChange={setChatDraft} onSubmit={sendChatMessage} />}
            {view === "data" && <DataWorkspace profile={profile} experiences={experiences} projects={projects} educations={educations} busy={backupBusy} lastResult={lastBackupResult} onExport={exportBackup} onImport={importBackup} />}
          </div>

          <ResumePreview profile={profile} experiences={experiences} projects={selectedProjects} educations={educations} />
        </main>
      </div>
    </div>
  );
}

function NavButton({ active, label, icon, badge, onClick }: { active: boolean; label: string; icon: IconName; badge?: string | number; onClick: () => void }) {
  return (
    <button className={`nav-button ${active ? "active" : ""}`} onClick={onClick}>
      <Icon name={icon} />
      <span className="nav-button-label">{label}</span>
      {badge !== undefined && <span className="nav-badge">{badge}</span>}
    </button>
  );
}

type IconName = "home" | "user" | "briefcase" | "folder" | "education" | "sparkles" | "code" | "target" | "chat" | "download" | "refresh" | "check" | "search" | "database";

const iconPaths: Record<IconName, React.ReactNode> = {
  home: <><path d="m3 10 9-7 9 7" /><path d="M5 9v11h14V9" /><path d="M9 20v-6h6v6" /></>,
  user: <><circle cx="12" cy="8" r="4" /><path d="M4 21c.8-4.2 3.5-6 8-6s7.2 1.8 8 6" /></>,
  briefcase: <><rect x="3" y="7" width="18" height="13" rx="2" /><path d="M8 7V4h8v3M3 12h18M10 12v2h4v-2" /></>,
  folder: <><path d="M3 6h7l2 2h9v11H3z" /><path d="M3 10h18" /></>,
  education: <><path d="m2 9 10-5 10 5-10 5z" /><path d="M6 11v5c3 2 9 2 12 0v-5M22 9v7" /></>,
  sparkles: <><path d="m12 2 1.5 5.5L19 9l-5.5 1.5L12 16l-1.5-5.5L5 9l5.5-1.5z" /><path d="m19 15 .8 2.2L22 18l-2.2.8L19 21l-.8-2.2L16 18l2.2-.8z" /></>,
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

function TopToolbar({ profile }: { profile: Profile }) {
  return <header className="top-toolbar">
    <div className="toolbar-brand"><span className="brand-mark">T</span><strong>TailorCV</strong></div>
    <span className="toolbar-divider" />
    <div className="document-title"><strong>{profile.headline || "Backend Engineer Resume"}</strong><small>Layout draft · profile saved locally</small></div>
    <div className="toolbar-spacer" />
    <div className="ats-pill"><span className="status-dot muted" /> Draft preview</div>
    <button className="toolbar-button" disabled title="LaTeX compilation is not connected yet"><Icon name="refresh" size={16} />Compile</button>
    <button className="export-button" disabled title="PDF export is available after compilation is implemented"><Icon name="download" size={16} />Export PDF</button>
    <button className="icon-button" aria-label="More document options">•••</button>
  </header>;
}

function Onboarding({ profile, skillsText, busy, onChange, onSkillsChange, onSubmit, onSkip }: { profile: Profile; skillsText: string; busy: boolean; onChange: (field: keyof Profile, value: string) => void; onSkillsChange: (value: string) => void; onSubmit: (event: FormEvent) => void; onSkip: () => void }) {
  return <div className="onboarding-overlay"><section className="onboarding-card"><aside><span className="brand-mark">T</span><p className="eyebrow">Welcome to TailorCV</p><h1>Start with facts you control.</h1><p>Create your local career profile. TailorCV uses it as the source of truth for every resume.</p><div className="onboarding-promise"><span className="status-dot" /><div><strong>Stored on this device</strong><small>No account or cloud sync required.</small></div></div></aside><form onSubmit={onSubmit}><header><div><span>01</span><p>Profile foundation</p></div><small>About 2 minutes</small></header><div className="onboarding-fields"><div className="field-grid two"><Field label="Full name" value={profile.name} onChange={(value) => onChange("name", value)} placeholder="Ada Lovelace" required /><Field label="Email" type="email" value={profile.email} onChange={(value) => onChange("email", value)} placeholder="ada@example.com" required /><Field label="Professional headline" value={profile.headline} onChange={(value) => onChange("headline", value)} placeholder="Backend engineer building reliable systems" required /><Field label="Location" value={profile.location} onChange={(value) => onChange("location", value)} placeholder="Bengaluru, India" /><Field label="GitHub username" value={profile.githubUsername} onChange={(value) => onChange("githubUsername", value)} placeholder="octocat" prefix="github.com/" /><Field label="LinkedIn URL" type="url" value={profile.linkedInUrl} onChange={(value) => onChange("linkedInUrl", value)} placeholder="https://linkedin.com/in/..." /></div><label className="field"><span>Core skills</span><textarea rows={3} value={skillsText} onChange={(event) => onSkillsChange(event.target.value)} placeholder="Go, TypeScript, PostgreSQL, Docker" /><small>Separate skills with commas.</small></label></div><footer><button className="text-button" type="button" onClick={onSkip}>Explore first</button><button className="primary-button" disabled={busy}>{busy ? "Creating profile…" : "Create local profile"}</button></footer></form></section></div>;
}

function PanelHeader({ eyebrow, title, description, action }: { eyebrow: string; title: string; description: string; action?: React.ReactNode }) {
  return <header className="panel-header"><div><p className="eyebrow">{eyebrow}</p><h1>{title}</h1><p>{description}</p></div>{action}</header>;
}

function WorkspaceOverview({ profile, completion, experiences, projects, onOpen }: { profile: Profile; completion: number; experiences: ExperienceDraft[]; projects: ProjectDraft[]; onOpen: (view: View) => void }) {
  return <section className="workspace-panel scroll-panel">
    <PanelHeader eyebrow="Your workspace" title={profile.name ? `Welcome back, ${profile.name.split(" ")[0]}.` : "Build your source of truth."} description="Keep one reliable career profile, then tailor the evidence for each role." action={<button className="primary-button" onClick={() => onOpen("job")}>Match a job</button>} />
    <div className="metric-grid">
      <article className="metric-card accent"><span>Profile readiness</span><strong>{completion}%</strong><div className="progress"><i style={{ width: `${completion}%` }} /></div><button onClick={() => onOpen("profile")}>Complete profile →</button></article>
      <article className="metric-card"><span>Evidence</span><strong>{experiences.reduce((sum, item) => sum + item.bullets.length, 0)}</strong><small>reviewable career claims</small></article>
      <article className="metric-card"><span>Projects</span><strong>{projects.length}</strong><small>{projects.filter((project) => project.resumeEligible).length} resume eligible</small></article>
    </div>
    <div className="overview-callout"><span className="callout-index">01</span><div><p className="eyebrow">Recommended next step</p><h2>{completion < 70 ? "Finish the profile foundation" : "Select evidence for your next application"}</h2><p>{completion < 70 ? "Add your identity, positioning, and core skills before tailoring." : "Choose projects and experience that directly support the target role."}</p></div><button className="secondary-button" onClick={() => onOpen(completion < 70 ? "profile" : "projects")}>Open workspace</button></div>
  </section>;
}

function ProfileWorkspace({ profile, busy, message, onChange, onSubmit }: { profile: Profile; busy: boolean; message: string; onChange: (field: keyof Profile, value: string) => void; onSubmit: (event: FormEvent) => void }) {
  return <section className="workspace-panel scroll-panel">
    <PanelHeader eyebrow="Career profile" title="Profile" description="Identity and positioning used across every resume." />
    <form className="compact-form" onSubmit={onSubmit}>
      <FormBlock title="Identity" description="Shown in the resume header."><div className="field-grid two"><Field label="Full name" value={profile.name} onChange={(value) => onChange("name", value)} placeholder="Ada Lovelace" /><Field label="Headline" value={profile.headline} onChange={(value) => onChange("headline", value)} placeholder="Backend engineer" /><Field label="Email" type="email" value={profile.email} onChange={(value) => onChange("email", value)} placeholder="ada@example.com" /><Field label="Phone" value={profile.phone} onChange={(value) => onChange("phone", value)} placeholder="+91 98765 43210" /><Field label="Location" value={profile.location} onChange={(value) => onChange("location", value)} placeholder="Bengaluru, India" /><Field label="Website" type="url" value={profile.website} onChange={(value) => onChange("website", value)} placeholder="https://example.com" /></div></FormBlock>
      <FormBlock title="Presence" description="Professional links and repository identity."><div className="field-grid two"><Field label="GitHub username" value={profile.githubUsername} onChange={(value) => onChange("githubUsername", value)} placeholder="octocat" prefix="github.com/" /><Field label="LinkedIn URL" type="url" value={profile.linkedInUrl} onChange={(value) => onChange("linkedInUrl", value)} placeholder="https://linkedin.com/in/..." /></div></FormBlock>
      <FormBlock title="Positioning" description="A factual summary, never generated without evidence."><label className="field"><span>Professional summary</span><textarea rows={7} maxLength={2400} value={profile.summary} onChange={(event) => onChange("summary", event.target.value)} placeholder="Summarize your experience, strengths, and outcomes." /><small>{profile.summary.length}/2400</small></label></FormBlock>
      <div className="sticky-form-actions"><span>{message}</span><button className="primary-button" disabled={busy}>{busy ? "Saving…" : "Save profile"}</button></div>
    </form>
  </section>;
}

function FormBlock({ title, description, children }: { title: string; description: string; children: React.ReactNode }) {
  return <section className="form-block"><header><h2>{title}</h2><p>{description}</p></header><div>{children}</div></section>;
}

function ProjectWorkspace({ projects, selectedKeys, busyKey, githubUsername, githubBusy, onToggle, onAdd, onUpdate, onSave, onDelete, onSyncGitHub, onOpenProfile }: { projects: ProjectDraft[]; selectedKeys: string[]; busyKey: string; githubUsername: string; githubBusy: boolean; onToggle: (key: string) => void; onAdd: () => void; onUpdate: (key: string, project: ProjectDraft) => void; onSave: (event: FormEvent, project: ProjectDraft) => void; onDelete: (project: ProjectDraft) => void; onSyncGitHub: () => void; onOpenProfile: () => void }) {
  const [tab, setTab] = useState<"select" | "manage">("select");
  const [query, setQuery] = useState("");
  const filteredProjects = filterProjects(projects, query);
  return <section className="workspace-panel scroll-panel">
    <PanelHeader eyebrow="Selected work" title="Projects" description="Choose the strongest evidence for this resume." action={<div className="panel-actions"><button className="secondary-button" onClick={githubUsername ? onSyncGitHub : onOpenProfile} disabled={githubBusy}>{githubBusy ? "Syncing…" : githubUsername ? "Sync GitHub" : "Connect GitHub"}</button><button className="secondary-button" onClick={() => { onAdd(); setTab("manage"); }}>Add project</button></div>} />
    <div className="panel-tabs"><button className={tab === "select" ? "active" : ""} onClick={() => setTab("select")}>Select for resume <span>{selectedKeys.length}</span></button><button className={tab === "manage" ? "active" : ""} onClick={() => setTab("manage")}>Manage evidence</button></div>
    {tab === "select" ? <div className="project-selector">
      <div className="github-sync-note"><span className="github-mark">GH</span><div><strong>{githubUsername ? `github.com/${githubUsername}` : "Connect your GitHub profile"}</strong><p>{githubUsername ? "Public, owned repositories sync into Manage evidence for review." : "Add a username in Profile to import your public repositories."}</p></div><button onClick={githubUsername ? onSyncGitHub : onOpenProfile} disabled={githubBusy}>{githubBusy ? "Syncing…" : githubUsername ? "Refresh" : "Open profile"}</button></div>
      <label className="search-field"><Icon name="search" size={16} /><input aria-label="Search projects" placeholder="Search projects, roles, or skills" value={query} onChange={(event) => setQuery(event.target.value)} /></label>
      <div className="selection-note"><span className="status-dot" /><div><strong>{selectedKeys.length} projects selected</strong><p>The preview updates as you select evidence.</p></div></div>
      {projects.length === 0 ? <div className="panel-empty"><span className="empty-icon"><Icon name="folder" size={22} /></span><strong>No projects yet</strong><p>Add a project with evidence bullets, technologies, and a review state.</p><button className="primary-button" onClick={() => { onAdd(); setTab("manage"); }}>Add first project</button></div> : filteredProjects.length === 0 ? <div className="panel-empty search-empty"><span className="empty-icon"><Icon name="search" size={22} /></span><strong>No matching projects</strong><p>Try a project name, role, description, or technology.</p><button className="text-button" onClick={() => setQuery("")}>Clear search</button></div> : filteredProjects.map((project) => {
        const selected = selectedKeys.includes(project.key);
        const selectable = project.resumeEligible;
        return <article className={`project-select-card ${selected ? "selected" : ""} ${!selectable ? "locked" : ""}`} key={project.key}>
          <button className="project-check" disabled={!selectable} aria-label={selectable ? `${selected ? "Remove" : "Add"} ${project.name || "project"} ${selected ? "from" : "to"} resume` : `${project.name || "Project"} must be reviewed before selection`} onClick={() => onToggle(project.key)}>{selected && <Icon name="check" size={14} />}</button>
          <div className="project-card-copy"><div className="project-title-row"><strong>{project.name || "Untitled project"}</strong><span>{selectable ? selected ? "Selected" : "Available" : "Review required"}</span></div><p>{project.description || "Add a concise description of the problem, your contribution, and the outcome."}</p><div className="tag-row">{project.skills.slice(0, 4).map((skill) => <span key={skill}>{skill}</span>)}{project.skills.length === 0 && <span>Skills not added</span>}</div><small>{project.verification === "verified" ? "Verified evidence" : "Needs review"} · {project.bullets.length} bullets</small></div>
        </article>;
      })}
    </div> : <ProjectSection projects={projects} busyKey={busyKey} onAdd={onAdd} onUpdate={onUpdate} onSave={onSave} onDelete={onDelete} />}
  </section>;
}

function EducationWorkspace({ educations, busyKey, onAdd, onUpdate, onSave, onDelete }: { educations: EducationDraft[]; busyKey: string; onAdd: () => void; onUpdate: (key: string, education: EducationDraft) => void; onSave: (event: FormEvent, education: EducationDraft) => void; onDelete: (education: EducationDraft) => void }) {
  return <section className="workspace-panel scroll-panel education-workspace"><PanelHeader eyebrow="Background" title="Education" description="Academic history that supports your application." action={<button className="secondary-button" onClick={onAdd}>Add education</button>} /><div className="education-list">{educations.length === 0 ? <div className="panel-empty tall"><span className="empty-icon"><Icon name="education" size={24} /></span><strong>No education records yet</strong><p>Add a degree, program, or current course of study. Saved records appear in the resume preview immediately.</p><button className="primary-button" onClick={onAdd}>Add education</button></div> : educations.map((education) => <EducationCard key={education.key} education={education} busy={busyKey === education.key} onUpdate={(next) => onUpdate(education.key, next)} onSave={(event) => onSave(event, education)} onDelete={() => onDelete(education)} />)}</div></section>;
}

function EducationCard({ education, busy, onUpdate, onSave, onDelete }: { education: EducationDraft; busy: boolean; onUpdate: (education: EducationDraft) => void; onSave: (event: FormEvent) => void; onDelete: () => void }) {
  const updateField = <K extends keyof EducationDraft>(field: K, value: EducationDraft[K]) => onUpdate({ ...education, [field]: value });
  return <form className="education-card" onSubmit={onSave}><header><div><span>{education.id ? "Saved education" : "New education"}</span><strong>{education.degree || "Untitled degree"}{education.institution ? ` · ${education.institution}` : ""}</strong></div><button className="danger-button" type="button" disabled={busy} onClick={onDelete}>Delete</button></header><div className="education-fields"><div className="field-grid two"><Field label="Institution" value={education.institution} onChange={(value) => updateField("institution", value)} placeholder="Example Institute" required /><Field label="Degree" value={education.degree} onChange={(value) => updateField("degree", value)} placeholder="Bachelor of Science" required /><Field label="Field of study" value={education.fieldOfStudy} onChange={(value) => updateField("fieldOfStudy", value)} placeholder="Computer Science" /><Field label="Location" value={education.location} onChange={(value) => updateField("location", value)} placeholder="Bengaluru, India" /><Field label="Start month" type="month" value={education.startDate} onChange={(value) => updateField("startDate", value)} placeholder="YYYY-MM" /><Field label="End month" type="month" value={education.endDate} onChange={(value) => updateField("endDate", value)} placeholder="YYYY-MM" disabled={education.current} /></div><label className="checkbox-field"><input type="checkbox" checked={education.current} onChange={(event) => updateField("current", event.target.checked)} /><span>I currently study here</span></label><label className="field"><span>Details (optional)</span><textarea rows={4} maxLength={1200} value={education.details} onChange={(event) => updateField("details", event.target.value)} placeholder="Honors, relevant coursework, thesis, or leadership." /><small>{education.details.length}/1200</small></label></div><footer><small>{education.updatedAt ? `Last saved ${new Date(education.updatedAt).toLocaleString()}` : "Not saved yet"}</small><button className="primary-button" disabled={busy}>{busy ? "Saving…" : "Save education"}</button></footer></form>;
}

function SkillsWorkspace({ skillsText, busy, message, onChange, onSubmit }: { skillsText: string; busy: boolean; message: string; onChange: (value: string) => void; onSubmit: (event: FormEvent) => void }) {
  const skills = parseSkills(skillsText);
  return <section className="workspace-panel scroll-panel"><PanelHeader eyebrow="Capabilities" title="Skills" description="A normalized vocabulary for matching and evidence selection." /><form className="compact-form skills-workspace" onSubmit={onSubmit}><FormBlock title="Skill inventory" description="Separate skills with commas. Duplicates are removed automatically."><label className="field"><span>Skills</span><textarea rows={6} value={skillsText} onChange={(event) => onChange(event.target.value)} placeholder="Go, TypeScript, PostgreSQL, Docker" /></label><div className="skill-chip-grid">{skills.map((skill) => <span key={skill}>{skill}<button type="button" aria-label={`Remove ${skill}`} onClick={() => onChange(skills.filter((item) => item !== skill).join(", "))}>×</button></span>)}</div></FormBlock><div className="sticky-form-actions"><span>{message}</span><button className="primary-button" disabled={busy}>{busy ? "Saving…" : "Save skills"}</button></div></form></section>;
}

function LatexWorkspace({ source, onChange }: { source: string; onChange: (value: string) => void }) {
  return <section className="workspace-panel latex-workspace"><PanelHeader eyebrow="Source editor" title="LaTeX" description="Edit the draft template directly. Compilation is not connected yet." /><div className="editor-tabs"><button className="active">resume.tex</button><button disabled title="Custom template files are not implemented yet">template.cls</button><span>Draft · unsaved</span></div><div className="code-editor"><div className="line-numbers">{source.split("\n").map((_, index) => <span key={index}>{index + 1}</span>)}</div><textarea spellCheck={false} value={source} onChange={(event) => onChange(event.target.value)} aria-label="LaTeX source" /></div><footer className="editor-status"><span>UTF-8</span><span>LaTeX draft</span><span>{source.split("\n").length} lines</span></footer></section>;
}

function AIWorkspace({ provider, draft, messages, onProviderChange, onDraftChange, onSubmit }: { provider: AIProvider; draft: string; messages: ChatMessage[]; onProviderChange: (provider: AIProvider) => void; onDraftChange: (value: string) => void; onSubmit: (event: FormEvent) => void }) {
  const providers: AIProvider[] = ["Ollama", "Gemini", "Claude", "OpenAI"];
  return <section className="workspace-panel ai-workspace"><PanelHeader eyebrow="Evidence-aware assistant" title="AI assistant" description="Use local or cloud models without letting them invent facts." /><div className="provider-picker">{providers.map((item) => <button key={item} className={provider === item ? "active" : ""} onClick={() => onProviderChange(item)}><span className="provider-logo">{item.slice(0, 2).toUpperCase()}</span><span><strong>{item}</strong><small>{item === "Ollama" ? "Local · private" : "Cloud · setup required"}</small></span>{provider === item && <Icon name="check" size={15} />}</button>)}</div><div className="chat-thread">{messages.map((item, index) => <div className={`chat-message ${item.role}`} key={`${item.role}-${index}`}><span>{item.role === "assistant" ? "AI" : "You"}</span><p>{item.text}</p></div>)}</div><form className="chat-composer" onSubmit={onSubmit}><textarea rows={3} value={draft} onChange={(event) => onDraftChange(event.target.value)} placeholder="Ask for a tighter bullet or compare selected evidence…" /><div><small>Only selected evidence will be shared.</small><button className="primary-button" type="submit">Send</button></div></form></section>;
}

function DataWorkspace({ profile, experiences, projects, educations, busy, lastResult, onExport, onImport }: { profile: Profile; experiences: ExperienceDraft[]; projects: ProjectDraft[]; educations: EducationDraft[]; busy: "export" | "import" | ""; lastResult: domain.BackupResult | null; onExport: () => void; onImport: () => void }) {
  const evidenceCount = experiences.reduce((sum, experience) => sum + experience.bullets.length, 0) + projects.reduce((sum, project) => sum + project.bullets.length, 0);
  return <section className="workspace-panel scroll-panel data-workspace"><PanelHeader eyebrow="Local data" title="Backup & restore" description="Keep a portable, versioned copy of your complete TailorCV profile." /><div className="backup-content"><section className="backup-summary"><header><span className="empty-icon"><Icon name="database" size={22} /></span><div><h2>Current local profile</h2><p>Everything listed here is included in one JSON backup.</p></div></header><div className="backup-stat-grid"><div><strong>{profile.name ? "1" : "0"}</strong><span>profile</span></div><div><strong>{experiences.length}</strong><span>roles</span></div><div><strong>{projects.length}</strong><span>projects</span></div><div><strong>{educations.length}</strong><span>education</span></div><div><strong>{profile.skills.length}</strong><span>skills</span></div><div><strong>{evidenceCount}</strong><span>evidence</span></div></div></section><section className="backup-action-card"><div><span className="backup-action-icon"><Icon name="download" size={20} /></span><h2>Export backup</h2><p>Write an owner-readable JSON snapshot using a native save dialog. IDs, ordering, verification state, and timestamps are preserved.</p></div><button className="primary-button" disabled={busy !== ""} onClick={onExport}>{busy === "export" ? "Exporting…" : "Choose destination"}</button></section><section className="backup-action-card restore"><div><span className="backup-action-icon"><Icon name="refresh" size={20} /></span><h2>Restore backup</h2><p>The entire file is validated before a single transaction replaces current data. Invalid or unsupported backups leave this profile untouched.</p></div><button className="secondary-button" disabled={busy !== ""} onClick={onImport}>{busy === "import" ? "Restoring…" : "Choose backup"}</button></section>{lastResult && <section className="backup-result"><span className="status-dot" /><div><strong>Last operation completed</strong><p>{lastResult.experienceCount} roles · {lastResult.projectCount} projects · {lastResult.educationCount} education records</p><small>{lastResult.path}</small></div></section>}<div className="backup-safety"><strong>Backup format v1</strong><p>Backups never include provider credentials, generated PDFs, compiler caches, or local model data.</p></div></div></section>;
}

function ResumePreview({ profile, experiences, projects, educations }: { profile: Profile; experiences: ExperienceDraft[]; projects: ProjectDraft[]; educations: EducationDraft[] }) {
  const visibleExperiences = experiences.slice(0, 2);
  const visibleProjects = projects.slice(0, 3);
  return <aside className="preview-pane"><header className="preview-toolbar"><div><strong>Layout preview</strong><small>Draft · not compiled</small></div><div className="preview-controls"><button aria-label="Zoom out" disabled>−</button><span>Fit</span><button aria-label="Zoom in" disabled>+</button><i /><span>1 page</span></div></header><div className="preview-stage"><article className="resume-paper">
    <header className="resume-header"><h1>{profile.name || "Your Name"}</h1><p>{profile.headline || "Backend Engineer"}</p><small>{[profile.email || "your@email.com", profile.phone, profile.location, profile.website].filter(Boolean).join("  ·  ")}</small></header>
    {profile.summary && <ResumeSection title="Summary"><p className="resume-summary">{profile.summary}</p></ResumeSection>}
    <ResumeSection title="Experience">{visibleExperiences.length ? visibleExperiences.map((experience) => <div className="resume-entry" key={experience.key}><div><strong>{experience.title || "Role"} · {experience.company || "Company"}</strong><span>{experience.startDate} — {experience.current ? "Present" : experience.endDate}</span></div>{experience.location && <em>{experience.location}</em>}<ul>{experience.bullets.slice(0, 3).map((bullet, index) => <li key={bullet.id || index}>{bullet.text}</li>)}</ul></div>) : <ResumePlaceholder text="Add experience and evidence bullets" />}</ResumeSection>
    <ResumeSection title="Projects">{visibleProjects.length ? visibleProjects.map((project) => <div className="resume-entry project" key={project.key}><div><strong>{project.name || "Project"}</strong><span>{project.skills.slice(0, 3).join(" · ")}</span></div>{project.description && <p>{project.description}</p>}<ul>{project.bullets.slice(0, 2).map((bullet, index) => <li key={bullet.id || index}>{bullet.text}</li>)}</ul></div>) : <ResumePlaceholder text="Select projects to include them" />}</ResumeSection>
    {educations.length > 0 && <ResumeSection title="Education">{educations.slice(0, 2).map((education) => <div className="resume-entry education" key={education.key}><div><strong>{education.degree}{education.fieldOfStudy ? `, ${education.fieldOfStudy}` : ""}</strong><span>{education.startDate}{education.startDate && (education.current || education.endDate) ? " — " : ""}{education.current ? "Present" : education.endDate}</span></div><em>{education.institution}{education.location ? ` · ${education.location}` : ""}</em>{education.details && <p>{education.details}</p>}</div>)}</ResumeSection>}
    <ResumeSection title="Skills"><p className="resume-skills">{profile.skills.length ? profile.skills.join("  ·  ") : "Add skills to your profile"}</p></ResumeSection>
  </article></div></aside>;
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
          <p>Version history arrives in milestone four.</p>
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
      <ProjectSection
        projects={projects}
        busyKey={projectBusyKey}
        onAdd={onAddProject}
        onUpdate={onUpdateProject}
        onSave={onSaveProject}
        onDelete={onDeleteProject}
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

function ProjectSection({ projects, busyKey, onAdd, onUpdate, onSave, onDelete }: {
  projects: ProjectDraft[];
  busyKey: string;
  onAdd: () => void;
  onUpdate: (key: string, project: ProjectDraft) => void;
  onSave: (event: FormEvent, project: ProjectDraft) => void;
  onDelete: (project: ProjectDraft) => void;
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
          onUpdate={(next) => onUpdate(project.key, next)}
          onSave={(event) => onSave(event, project)}
          onDelete={() => onDelete(project)}
        />
      ))}
    </section>
  );
}

function ProjectCard({ project, busy, onUpdate, onSave, onDelete }: {
  project: ProjectDraft;
  busy: boolean;
  onUpdate: (project: ProjectDraft) => void;
  onSave: (event: FormEvent) => void;
  onDelete: () => void;
}) {
  const updateField = <K extends keyof ProjectDraft>(field: K, value: ProjectDraft[K]) => {
    onUpdate({ ...project, [field]: value });
  };

  return (
    <form className="experience-card project-card" onSubmit={onSave}>
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
        <label className="field full">
          <span>Technologies and skills</span>
          <textarea rows={2} value={project.skillsText} onChange={(event) => updateField("skillsText", event.target.value)} placeholder="Go, React, SQLite, Docker" />
          <small>Separate skills with commas. Duplicates are removed when saved.</small>
        </label>
        <div className="review-fields">
          <label className="field">
            <span>Project review state</span>
            <select value={project.verification} onChange={(event) => updateField("verification", event.target.value as ProjectDraft["verification"])}>
              <option value="unverified">Needs review</option>
              <option value="verified">Verified</option>
            </select>
          </label>
          <div className="provenance-field"><span>Origin</span><strong>{project.provenance}</strong></div>
          {!project.resumeEligible && <p>Excluded from resume selection until you mark it eligible.</p>}
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

function JobTailor({ description, analysis, busy, hasSkills, onDescriptionChange, onSubmit, onProfile }: {
  description: string;
  analysis: JobAnalysis | null;
  busy: boolean;
  hasSkills: boolean;
  onDescriptionChange: (value: string) => void;
  onSubmit: (event: FormEvent) => void;
  onProfile: () => void;
}) {
  return (
    <section className="page form-page">
      <header className="page-header compact"><div><p className="eyebrow">Job workspace</p><h1>Find the strongest fit</h1><p>Start with a transparent comparison before generating or rewriting anything.</p></div></header>
      {!hasSkills && <div className="empty-callout"><strong>Your profile has no skills yet.</strong><span>Add skills first so the comparison has reliable evidence.</span><button onClick={onProfile}>Open career profile</button></div>}
      <div className="tailor-grid">
        <form className="job-card" onSubmit={onSubmit}>
          <div className="card-heading"><span className="step-pill">Step 1</span><h2>Paste the job description</h2></div>
          <textarea value={description} onChange={(event) => onDescriptionChange(event.target.value)} placeholder="Paste the complete role description here…" rows={18} />
          <div className="job-actions"><small>{description.length.toLocaleString()} characters</small><button className="primary-button" disabled={busy || !hasSkills} type="submit">{busy ? "Analyzing…" : "Analyze fit"}</button></div>
        </form>
        <section className="analysis-card">
          <div className="card-heading"><span className="step-pill muted">Step 2</span><h2>Review the evidence</h2></div>
          {!analysis ? <div className="analysis-empty"><div className="radar">✦</div><strong>Your comparison will appear here</strong><p>TailorCV will show exactly which stored skills were found—without pretending this is a hiring score.</p></div> : <AnalysisResult analysis={analysis} />}
        </section>
      </div>
    </section>
  );
}

function AnalysisResult({ analysis }: { analysis: JobAnalysis }) {
  return <div className="analysis-result"><div className="score-row"><div className="score-ring" style={{ "--score": `${analysis.score * 3.6}deg` } as React.CSSProperties}><span>{analysis.score}%</span></div><div><strong>Profile skill alignment</strong><p>{analysis.explanation}</p></div></div><SkillGroup title="Explicit matches" skills={analysis.matchedSkills} tone="match" /><SkillGroup title="Not mentioned" skills={analysis.unmentionedSkills} tone="neutral" /></div>;
}

function SkillGroup({ title, skills, tone }: { title: string; skills: string[]; tone: "match" | "neutral" }) {
  return <div className="skill-group"><div className="skill-heading"><h3>{title}</h3><span>{skills.length}</span></div><div className="skill-list">{skills.length ? skills.map((skill) => <span className={`skill ${tone}`} key={skill}>{skill}</span>) : <p>None yet.</p>}</div></div>;
}

import { FormEvent, useEffect, useMemo, useState } from "react";
import {
  AnalyzeJobDescription,
  GetProfile,
  SaveProfile,
} from "../wailsjs/go/main/App";
import {
  emptyProfile,
  parseSkills,
  Profile,
  profileCompletion,
} from "./lib/profile";

type View = "home" | "profile" | "tailor";

type JobAnalysis = {
  score: number;
  matchedSkills: string[];
  unmentionedSkills: string[];
  explanation: string;
};

function errorMessage(error: unknown): string {
  if (error instanceof Error) return error.message;
  if (typeof error === "string") return error;
  return "Something went wrong. Please try again.";
}

export default function App() {
  const [view, setView] = useState<View>("home");
  const [profile, setProfile] = useState<Profile>(emptyProfile);
  const [skillsText, setSkillsText] = useState("");
  const [jobDescription, setJobDescription] = useState("");
  const [analysis, setAnalysis] = useState<JobAnalysis | null>(null);
  const [busy, setBusy] = useState(true);
  const [message, setMessage] = useState("");
  const [error, setError] = useState("");

  useEffect(() => {
    GetProfile()
      .then((result) => {
        const loaded = { ...emptyProfile, ...result } as Profile;
        loaded.skills ??= [];
        setProfile(loaded);
        setSkillsText(loaded.skills.join(", "));
      })
      .catch((reason) => setError(errorMessage(reason)))
      .finally(() => setBusy(false));
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

  return (
    <div className="app-shell">
      <aside className="sidebar">
        <button className="brand" onClick={() => setView("home")}>
          <span className="brand-mark">T</span>
          <span>
            TailorCV
            <small>Local resume studio</small>
          </span>
        </button>

        <nav aria-label="Main navigation">
          <NavButton active={view === "home"} label="Overview" icon="⌂" onClick={() => setView("home")} />
          <NavButton active={view === "profile"} label="Career profile" icon="◎" onClick={() => setView("profile")} />
          <NavButton active={view === "tailor"} label="Tailor a resume" icon="✦" onClick={() => setView("tailor")} />
        </nav>

        <div className="privacy-note">
          <span className="privacy-dot" />
          <div>
            <strong>Local first</strong>
            <p>Your profile is stored on this device.</p>
          </div>
        </div>
      </aside>

      <main>
        {error && (
          <div className="notice error" role="alert">
            {error}
            <button aria-label="Dismiss error" onClick={() => setError("")}>×</button>
          </div>
        )}
        {view === "home" && (
          <Overview
            profile={profile}
            completion={completion}
            onProfile={() => setView("profile")}
            onTailor={() => setView("tailor")}
          />
        )}
        {view === "profile" && (
          <ProfileEditor
            profile={profile}
            skillsText={skillsText}
            busy={busy}
            message={message}
            onChange={updateProfile}
            onSkillsChange={setSkillsText}
            onSubmit={saveProfile}
          />
        )}
        {view === "tailor" && (
          <JobTailor
            description={jobDescription}
            analysis={analysis}
            busy={busy}
            hasSkills={profile.skills.length > 0}
            onDescriptionChange={setJobDescription}
            onSubmit={analyzeJob}
            onProfile={() => setView("profile")}
          />
        )}
      </main>
    </div>
  );
}

function NavButton({ active, label, icon, onClick }: { active: boolean; label: string; icon: string; onClick: () => void }) {
  return (
    <button className={`nav-button ${active ? "active" : ""}`} onClick={onClick}>
      <span>{icon}</span>{label}
    </button>
  );
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

function ProfileEditor({ profile, skillsText, busy, message, onChange, onSkillsChange, onSubmit }: {
  profile: Profile;
  skillsText: string;
  busy: boolean;
  message: string;
  onChange: (field: keyof Profile, value: string) => void;
  onSkillsChange: (value: string) => void;
  onSubmit: (event: FormEvent) => void;
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
    </section>
  );
}

function FormSection({ title, description, children }: { title: string; description: string; children: React.ReactNode }) {
  return <section className="form-section"><div className="section-copy"><h2>{title}</h2><p>{description}</p></div><div className="section-fields">{children}</div></section>;
}

function Field({ label, value, placeholder, type = "text", prefix, onChange }: { label: string; value: string; placeholder: string; type?: string; prefix?: string; onChange: (value: string) => void }) {
  return <label className="field"><span>{label}</span><div className={prefix ? "prefixed-input" : ""}>{prefix && <em>{prefix}</em>}<input type={type} value={value} placeholder={placeholder} onChange={(event) => onChange(event.target.value)} /></div></label>;
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

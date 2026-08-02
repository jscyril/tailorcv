import { FormEvent, useEffect, useMemo, useState } from "react";
import {
  AnalyzeJobDescription,
  DeleteExperience,
  GetProfile,
  ListExperiences,
  SaveExperience,
  SaveProfile,
} from "../wailsjs/go/main/App";
import { domain } from "../wailsjs/go/models";
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
  const [experiences, setExperiences] = useState<ExperienceDraft[]>([]);
  const [skillsText, setSkillsText] = useState("");
  const [jobDescription, setJobDescription] = useState("");
  const [analysis, setAnalysis] = useState<JobAnalysis | null>(null);
  const [busy, setBusy] = useState(true);
  const [experienceBusyKey, setExperienceBusyKey] = useState("");
  const [message, setMessage] = useState("");
  const [error, setError] = useState("");

  useEffect(() => {
    Promise.all([GetProfile(), ListExperiences()])
      .then(([result, savedExperiences]) => {
        const loaded = { ...emptyProfile, ...result } as Profile;
        loaded.skills ??= [];
        setProfile(loaded);
        setSkillsText(loaded.skills.join(", "));
        setExperiences((savedExperiences as unknown as Experience[]).map(experienceToDraft));
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
            experiences={experiences}
            skillsText={skillsText}
            busy={busy}
            experienceBusyKey={experienceBusyKey}
            message={message}
            onChange={updateProfile}
            onSkillsChange={setSkillsText}
            onSubmit={saveProfile}
            onAddExperience={addExperience}
            onUpdateExperience={updateExperience}
            onSaveExperience={saveExperience}
            onDeleteExperience={deleteExperience}
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

function ProfileEditor({
  profile,
  experiences,
  skillsText,
  busy,
  experienceBusyKey,
  message,
  onChange,
  onSkillsChange,
  onSubmit,
  onAddExperience,
  onUpdateExperience,
  onSaveExperience,
  onDeleteExperience,
}: {
  profile: Profile;
  experiences: ExperienceDraft[];
  skillsText: string;
  busy: boolean;
  experienceBusyKey: string;
  message: string;
  onChange: (field: keyof Profile, value: string) => void;
  onSkillsChange: (value: string) => void;
  onSubmit: (event: FormEvent) => void;
  onAddExperience: () => void;
  onUpdateExperience: (key: string, experience: ExperienceDraft) => void;
  onSaveExperience: (event: FormEvent, experience: ExperienceDraft) => void;
  onDeleteExperience: (experience: ExperienceDraft) => void;
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
  const updateBullet = (index: number, patch: Partial<EvidenceBullet>) => {
    updateField("bullets", experience.bullets.map((bullet, bulletIndex) => bulletIndex === index ? { ...bullet, ...patch } : bullet));
  };
  const removeBullet = (index: number) => {
    updateField("bullets", experience.bullets.filter((_, bulletIndex) => bulletIndex !== index));
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

      <div className="evidence-heading">
        <div><strong>Evidence bullets</strong><span>Use specific, defensible work and outcomes.</span></div>
        <button className="text-button" type="button" onClick={() => updateField("bullets", [...experience.bullets, newEvidenceBullet()])}>+ Add bullet</button>
      </div>
      <div className="evidence-list">
        {experience.bullets.length === 0 && <p className="evidence-empty">No evidence bullets. The role can still be saved.</p>}
        {experience.bullets.map((bullet, index) => (
          <div className="evidence-row" key={bullet.id || `new-bullet-${index}`}>
            <div className="evidence-order" aria-label={`Reorder evidence bullet ${index + 1}`}>
              <span>{index + 1}</span>
              <button type="button" aria-label="Move bullet up" disabled={index === 0} onClick={() => updateField("bullets", moveItem(experience.bullets, index, index - 1))}>↑</button>
              <button type="button" aria-label="Move bullet down" disabled={index === experience.bullets.length - 1} onClick={() => updateField("bullets", moveItem(experience.bullets, index, index + 1))}>↓</button>
            </div>
            <div className="evidence-fields">
              <label className="field full">
                <span>Claim or outcome</span>
                <textarea rows={3} maxLength={1200} value={bullet.text} onChange={(event) => updateBullet(index, { text: event.target.value })} placeholder="Reduced deployment time by 40% by replacing manual release steps with an audited pipeline." />
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
      <div className="experience-actions">
        <small>{experience.updatedAt ? `Last saved ${new Date(experience.updatedAt).toLocaleString()}` : "Not saved yet"}</small>
        <button className="primary-button" disabled={busy} type="submit">{busy ? "Saving…" : "Save experience"}</button>
      </div>
    </form>
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

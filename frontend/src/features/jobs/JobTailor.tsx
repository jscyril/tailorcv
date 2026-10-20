import { history } from "@codemirror/commands";
import { FormEvent } from "react";
import { Profile } from "../../lib/profile";
import { type Application, type ResumeVersion } from "../../lib/resume";
import { Field } from "../../components/ui";
import { type Job, type JobAnalysis } from "../ai/AIWorkspace";
import { type ApplicationStatus, ApplicationStatusControl } from "../applications/ApplicationStatusControl";

export function JobTailor({ job, jobs, application, analysis, selectedFactIDs, busy, versionBusy, statusBusy, hasEvidence, templateName, onChange, onNew, onOpen, onDelete, onToggleEvidence, onCreateVersion, onOpenVersion, onStatusChange, onSubmit, onProfile }: {
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

export function AnalysisResult({ analysis, selectedFactIDs, onToggle }: { analysis: JobAnalysis; selectedFactIDs: string[]; onToggle: (factID: string) => void }) {
  return <div className="analysis-result"><div className="score-row"><div className="score-ring" style={{ "--score": `${analysis.score * 3.6}deg` } as React.CSSProperties}><span>{analysis.score}%</span></div><div><strong>Profile skill alignment</strong><p>{analysis.explanation}</p></div></div><SkillGroup title="Required skills" skills={analysis.requiredSkills} tone="match" />{analysis.preferredSkills.length > 0 && <SkillGroup title="Preferred skills" skills={analysis.preferredSkills} tone="match" />}<SkillGroup title="Profile skills not mentioned" skills={analysis.unmentionedSkills} tone="neutral" />{analysis.responsibilities.length > 0 && <section className="requirement-summary"><div className="skill-heading"><h3>Responsibilities extracted</h3><span>{analysis.responsibilities.length}</span></div><ul>{analysis.responsibilities.map((responsibility) => <li key={responsibility}>{responsibility}</li>)}</ul></section>}<div className="keyword-summary"><strong>Search vocabulary</strong><div className="skill-list">{analysis.keywords.map((keyword) => <span className={`skill neutral`} key={keyword}>{keyword}</span>)}</div></div><div className="ranked-evidence"><div className="skill-heading"><h3>Ranked career evidence</h3><span>{analysis.rankedEvidence.length}</span></div>{analysis.rankedEvidence.length ? analysis.rankedEvidence.map((evidence) => {
    const selected = selectedFactIDs.includes(evidence.factId);
    return <article className={`${selected ? "selected" : ""} ${!evidence.selectable ? "locked" : ""}`} key={`${evidence.sourceType}-${evidence.factId}`}><header><label><input aria-label={`${selected ? "Remove" : "Add"} ${evidence.sourceLabel} evidence ${selected ? "from" : "to"} resume`} type="checkbox" checked={selected} disabled={!evidence.selectable} onChange={() => onToggle(evidence.factId)} /><div><span>{evidence.sourceType}</span><strong>{evidence.sourceLabel}</strong></div></label><b>{evidence.score}</b></header><p>{evidence.text}</p><div className="evidence-reasons">{evidence.reasons.map((reason) => <span key={reason}>{reason}</span>)}{!evidence.selectable && <span>Review project to unlock</span>}</div></article>;
  }) : <p className="no-ranked-evidence">No saved evidence shares meaningful skills or terms with this description yet.</p>}</div></div>;
}

export function SkillGroup({ title, skills, tone }: { title: string; skills: string[]; tone: "match" | "neutral" }) {
  return <div className="skill-group"><div className="skill-heading"><h3>{title}</h3><span>{skills.length}</span></div><div className="skill-list">{skills.length ? skills.map((skill) => <span className={`skill ${tone}`} key={skill}>{skill}</span>) : <p>None yet.</p>}</div></div>;
}

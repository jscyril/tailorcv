import { useState } from "react";
import { ExperienceDraft } from "../../lib/experience";
import { Profile } from "../../lib/profile";
import { ProjectDraft } from "../../lib/project";
import { type Application, type ResumeVersion, type View } from "../../lib/resume";
import { PanelHeader } from "../../components/ui";
import { type Job } from "../ai/AIWorkspace";
import { VersionCompare } from "./VersionCompare";

export function resumeTitle(application: Application, jobs: Job[]): string {
  if (application.name) return application.name;
  const job = jobs.find((item) => item.id === application.jobId);
  return [job?.role, job?.company].filter(Boolean).join(" at ") || "Untitled resume";
}

export function WorkspaceOverview({ profile, completion, experiences, projects, applications, jobs, busy, onOpen, onNewResume, onNewGeneralResume, onOpenJob, onOpenVersion, onRename, onDuplicate, onDelete }: { profile: Profile; completion: number; experiences: ExperienceDraft[]; projects: ProjectDraft[]; applications: Application[]; jobs: Job[]; busy: boolean; onOpen: (view: View) => void; onNewResume: () => void; onNewGeneralResume: () => void; onOpenJob: (job: Job) => void; onOpenVersion: (version: ResumeVersion) => void; onRename: (application: Application, name: string) => Promise<boolean>; onDuplicate: (application: Application) => void; onDelete: (application: Application, title: string) => void }) {
  const versionCount = applications.reduce((sum, application) => sum + application.versions.length, 0);
  const [renaming, setRenaming] = useState<{ id: string; name: string } | null>(null);
  const [comparing, setComparing] = useState<string | null>(null);
  const resumes = [...applications].sort((a, b) => b.updatedAt.localeCompare(a.updatedAt));
  return <section className="workspace-panel scroll-panel">
    <PanelHeader eyebrow="Your workspace" title={profile.name ? `Welcome back, ${profile.name.split(" ")[0]}.` : "Build your source of truth."} description="Keep one reliable career profile, then create general or job-tailored resumes from it." action={<div className="panel-actions"><button className="secondary-button" disabled={busy} onClick={onNewGeneralResume}>New general resume</button><button className="primary-button" onClick={onNewResume}>New tailored resume</button></div>} />
    <section className="resume-library" aria-labelledby="resume-library-title">
      <header><h2 id="resume-library-title">Your resumes</h2><span>{resumes.length} resume{resumes.length === 1 ? "" : "s"} · {versionCount} version{versionCount === 1 ? "" : "s"}</span></header>
      {resumes.length ? <ul>{resumes.map((application) => {
        const title = resumeTitle(application, jobs);
        const job = jobs.find((item) => item.id === application.jobId);
        const latest = application.versions[0];
        const isRenaming = renaming?.id === application.id;
        return <li key={application.id}>
          <div className="resume-library-row">
            <div className="resume-library-title">
              {isRenaming ? <form onSubmit={async (event) => { event.preventDefault(); if (await onRename(application, renaming.name)) setRenaming(null); }}>
                <label className="visually-hidden" htmlFor={`rename-${application.id}`}>Resume name</label>
                <input id={`rename-${application.id}`} autoFocus maxLength={120} value={renaming.name} onChange={(event) => setRenaming({ id: application.id, name: event.target.value })} onKeyDown={(event) => { if (event.key === "Escape") setRenaming(null); }} />
                <button className="primary-button" type="submit" disabled={!renaming.name.trim()}>Save</button>
                <button className="text-button" type="button" onClick={() => setRenaming(null)}>Cancel</button>
              </form> : <strong>{title}</strong>}
              <small>{job ? `Tailored to ${[job.role, job.company].filter(Boolean).join(" at ") || "a saved job"}` : "General resume"}{latest ? ` · version ${latest.versionNumber} · ${new Date(latest.createdAt).toLocaleDateString()}${latest.pdfAvailable ? " · PDF ready" : ""}` : ""}</small>
            </div>
            <span className={`status-pill ${application.status}`}>{application.status}</span>
            <div className="resume-library-actions">
              {latest && <button className="primary-button" aria-label={`Open latest version of ${title}`} onClick={() => onOpenVersion(latest)}>Open</button>}
              {job && <button className="secondary-button" aria-label={`Tailor ${title}`} onClick={() => onOpenJob(job)}>Tailor</button>}
              <button className="secondary-button" aria-label={`Rename ${title}`} onClick={() => setRenaming({ id: application.id, name: title })}>Rename</button>
              <button className="secondary-button" aria-label={`Duplicate ${title}`} disabled={!latest} onClick={() => onDuplicate(application)}>Duplicate</button>
              <button className="secondary-button danger" aria-label={`Delete ${title}`} onClick={() => onDelete(application, title)}>Delete</button>
            </div>
          </div>
          {application.versions.length > 0 && <details className="resume-versions">
            <summary>{application.versions.length} version{application.versions.length === 1 ? "" : "s"}</summary>
            <ol>{application.versions.map((version) => <li key={version.id}>
              <span>Version {version.versionNumber}</span>
              <small>{new Date(version.createdAt).toLocaleString()}{version.pdfAvailable ? " · PDF" : version.compiledAt && !version.compileSuccess ? " · failed to compile" : ""}</small>
              <button className="text-button" aria-label={`Open version ${version.versionNumber} of ${title}`} onClick={() => onOpenVersion(version)}>Open</button>
            </li>)}</ol>
            {application.versions.length > 1 && (comparing === application.id
              ? <VersionCompare versions={application.versions} title={title} onClose={() => setComparing(null)} />
              : <button className="secondary-button" aria-label={`Compare versions of ${title}`} onClick={() => setComparing(application.id)}>Compare versions</button>)}
          </details>}
        </li>;
      })}</ul> : <div className="resume-library-empty"><p>No resumes yet. Create a general resume from your profile and the selected template, or paste a job description to build a tailored one.</p><div className="panel-actions"><button className="secondary-button" disabled={busy} onClick={onNewGeneralResume}>New general resume</button><button className="primary-button" onClick={onNewResume}>New tailored resume</button></div></div>}
    </section>
    <div className="metric-grid">
      <article className="metric-card accent"><span>Profile readiness</span><strong>{completion}%</strong><div className="progress"><i style={{ width: `${completion}%` }} /></div><button onClick={() => onOpen("profile")}>Complete profile →</button></article>
      <article className="metric-card"><span>Evidence</span><strong>{experiences.reduce((sum, item) => sum + item.bullets.length, 0)}</strong><small>reviewable career claims</small></article>
      <article className="metric-card"><span>Resume versions</span><strong>{versionCount}</strong><small>{applications.length} saved applications · {projects.filter((project) => project.resumeEligible).length} eligible projects</small></article>
    </div>
    <div className="overview-callout"><span className="callout-index">01</span><div><p className="eyebrow">Recommended next step</p><h2>{completion < 70 ? "Finish the profile foundation" : "Select evidence for your next application"}</h2><p>{completion < 70 ? "Add your identity, positioning, and core skills before tailoring." : "Choose projects and experience that directly support the target role."}</p></div><button className="secondary-button" onClick={() => onOpen(completion < 70 ? "profile" : "projects")}>Open workspace</button></div>
  </section>;
}

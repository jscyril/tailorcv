import { useState } from "react";
import { ExperienceDraft } from "../../lib/experience";
import { Profile } from "../../lib/profile";
import { ProjectDraft } from "../../lib/project";
import { type Application, type ResumeVersion, type View } from "../../lib/resume";
import { PanelHeader } from "../../components/ui";
import { type Job } from "../ai/AIWorkspace";
import { VersionCompare } from "./VersionCompare";
import { ApplicationTracker, followUpDue, type ApplicationDetails } from "../applications/ApplicationTracker";
import { CoverLetterPanel } from "../coverletter/CoverLetterPanel";
import { STATUS_OPTIONS, type ApplicationStatus } from "../applications/ApplicationStatusControl";

export type CoverLetterHandlers = {
  aiModel: string;
  busyId: string;
  draft: (application: Application) => Promise<string | null>;
  generate: (application: Application) => Promise<string | null>;
  save: (application: Application, text: string) => Promise<boolean>;
  export: (application: Application, text: string) => void;
};

export function resumeTitle(application: Application, jobs: Job[]): string {
  if (application.name) return application.name;
  const job = jobs.find((item) => item.id === application.jobId);
  return [job?.role, job?.company].filter(Boolean).join(" at ") || "Untitled resume";
}

export function WorkspaceOverview({ profile, completion, experiences, projects, applications, jobs, busy, onOpen, onNewResume, onNewGeneralResume, onOpenJob, onOpenVersion, onRename, onDuplicate, onDelete, statusBusy, onStatusChange, onSaveDetails, coverLetters }: { profile: Profile; completion: number; experiences: ExperienceDraft[]; projects: ProjectDraft[]; applications: Application[]; jobs: Job[]; busy: boolean; onOpen: (view: View) => void; onNewResume: () => void; onNewGeneralResume: () => void; onOpenJob: (job: Job) => void; onOpenVersion: (version: ResumeVersion) => void; onRename: (application: Application, name: string) => Promise<boolean>; onDuplicate: (application: Application) => void; onDelete: (application: Application, title: string) => void; statusBusy: boolean; onStatusChange: (status: ApplicationStatus, application: Application) => void; onSaveDetails: (application: Application, details: ApplicationDetails) => Promise<boolean>; coverLetters: CoverLetterHandlers }) {
  const versionCount = applications.reduce((sum, application) => sum + application.versions.length, 0);
  const [renaming, setRenaming] = useState<{ id: string; name: string } | null>(null);
  const [comparing, setComparing] = useState<string | null>(null);
  const [search, setSearch] = useState("");
  const [statusFilter, setStatusFilter] = useState<ApplicationStatus | "">("");
  const query = search.trim().toLowerCase();
  const resumes = [...applications].sort((a, b) => b.updatedAt.localeCompare(a.updatedAt)).filter((application) => {
    if (statusFilter && application.status !== statusFilter) return false;
    if (!query) return true;
    const job = jobs.find((item) => item.id === application.jobId);
    return [resumeTitle(application, jobs), job?.role, job?.company, application.notes].some((text) => text?.toLowerCase().includes(query));
  });
  return <section className="workspace-panel scroll-panel">
    <PanelHeader eyebrow="Your workspace" title={profile.name ? `Welcome back, ${profile.name.split(" ")[0]}.` : "Build your source of truth."} description="Keep one reliable career profile, then create general or job-tailored resumes from it." action={<div className="panel-actions"><button className="secondary-button" disabled={busy} onClick={onNewGeneralResume}>New general resume</button><button className="primary-button" onClick={onNewResume}>New tailored resume</button></div>} />
    <section className="resume-library" aria-labelledby="resume-library-title">
      <header><h2 id="resume-library-title">Your resumes</h2><span>{applications.length} resume{applications.length === 1 ? "" : "s"} · {versionCount} version{versionCount === 1 ? "" : "s"}</span></header>
      {applications.length > 0 && <div className="resume-library-filters">
        <input type="search" aria-label="Search resumes" placeholder="Search by name, company, role, or notes" value={search} onChange={(event) => setSearch(event.target.value)} />
        <select aria-label="Filter by status" value={statusFilter} onChange={(event) => setStatusFilter(event.target.value as ApplicationStatus | "")}>
          <option value="">All statuses</option>
          {STATUS_OPTIONS.map((option) => <option key={option.value} value={option.value}>{option.label} ({applications.filter((application) => application.status === option.value).length})</option>)}
        </select>
      </div>}
      {applications.length > 0 && resumes.length === 0 && <p className="resume-library-no-match">No resumes match. <button className="text-button" onClick={() => { setSearch(""); setStatusFilter(""); }}>Clear filters</button></p>}
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
              <small>{job ? `Tailored to ${[job.role, job.company].filter(Boolean).join(" at ") || "a saved job"}` : "General resume"}{latest ? ` · version ${latest.versionNumber} · ${new Date(latest.createdAt).toLocaleDateString()}${latest.pdfAvailable ? " · PDF ready" : ""}` : ""}{application.appliedOn ? ` · applied ${application.appliedOn}` : ""}</small>
            </div>
            <span className="resume-library-badges">{followUpDue(application) && <span className="status-pill follow-up">Follow-up due</span>}<span className={`status-pill ${application.status}`}>{application.status}</span></span>
            <div className="resume-library-actions">
              {latest && <button className="primary-button" aria-label={`Open latest version of ${title}`} onClick={() => onOpenVersion(latest)}>Open</button>}
              {job && <button className="secondary-button" aria-label={`Tailor ${title}`} onClick={() => onOpenJob(job)}>Tailor</button>}
              <button className="secondary-button" aria-label={`Rename ${title}`} onClick={() => setRenaming({ id: application.id, name: title })}>Rename</button>
              <button className="secondary-button" aria-label={`Duplicate ${title}`} disabled={!latest} onClick={() => onDuplicate(application)}>Duplicate</button>
              <button className="secondary-button danger" aria-label={`Delete ${title}`} onClick={() => onDelete(application, title)}>Delete</button>
            </div>
          </div>
          <details className="resume-versions">
            <summary>Track application</summary>
            <ApplicationTracker application={application} title={title} busy={statusBusy} onStatusChange={(status) => onStatusChange(status, application)} onSave={(details) => onSaveDetails(application, details)} />
          </details>
          {job && <details className="resume-versions">
            <summary>Cover letter{application.coverLetter ? "" : " (none yet)"}</summary>
            <CoverLetterPanel application={application} title={title} aiModel={coverLetters.aiModel} busy={coverLetters.busyId === application.id} onDraft={() => coverLetters.draft(application)} onGenerate={() => coverLetters.generate(application)} onSave={(text) => coverLetters.save(application, text)} onExport={(text) => coverLetters.export(application, text)} />
          </details>}
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
      })}</ul> : applications.length > 0 ? null : <div className="resume-library-empty"><p>No resumes yet. Create a general resume from your profile and the selected template, or paste a job description to build a tailored one.</p><div className="panel-actions"><button className="secondary-button" disabled={busy} onClick={onNewGeneralResume}>New general resume</button><button className="primary-button" onClick={onNewResume}>New tailored resume</button></div></div>}
    </section>
    <div className="metric-grid">
      <article className="metric-card accent"><span>Profile readiness</span><strong>{completion}%</strong><div className="progress"><i style={{ width: `${completion}%` }} /></div><button onClick={() => onOpen("profile")}>Complete profile →</button></article>
      <article className="metric-card"><span>Evidence</span><strong>{experiences.reduce((sum, item) => sum + item.bullets.length, 0)}</strong><small>reviewable career claims</small></article>
      <article className="metric-card"><span>Resume versions</span><strong>{versionCount}</strong><small>{applications.length} saved applications · {projects.filter((project) => project.resumeEligible).length} eligible projects</small></article>
    </div>
    <div className="overview-callout"><span className="callout-index">01</span><div><p className="eyebrow">Recommended next step</p><h2>{completion < 70 ? "Finish the profile foundation" : "Select evidence for your next application"}</h2><p>{completion < 70 ? "Add your identity, positioning, and core skills before tailoring." : "Choose projects and experience that directly support the target role."}</p></div><button className="secondary-button" onClick={() => onOpen(completion < 70 ? "profile" : "projects")}>Open workspace</button></div>
  </section>;
}

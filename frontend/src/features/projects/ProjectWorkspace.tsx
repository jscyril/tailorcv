import { FormEvent, useEffect, useState } from "react";
import { Profile, parseSkills } from "../../lib/profile";
import { Project, ProjectDraft, filterProjects, isProjectSelectable, toggleProjectLanguage } from "../../lib/project";
import { EvidenceEditor } from "../../components/EvidenceEditor";
import { Field, Icon, PanelHeader } from "../../components/ui";

export function ProjectWorkspace({ projects, selectedKeys, busyKey, readmeBusyKey, aiReady, githubUsername, githubBusy, onToggle, onAdd, onUpdate, onSave, onDelete, onGenerateReadmeBullets, onImportGitHubProject, onSyncGitHub, onOpenProfile }: { projects: ProjectDraft[]; selectedKeys: string[]; busyKey: string; readmeBusyKey: string; aiReady: boolean; githubUsername: string; githubBusy: boolean; onToggle: (key: string) => void; onAdd: () => void; onUpdate: (key: string, project: ProjectDraft) => void; onSave: (event: FormEvent, project: ProjectDraft) => void; onDelete: (project: ProjectDraft) => void; onGenerateReadmeBullets: (project: ProjectDraft) => void; onImportGitHubProject: (repositoryURL: string) => void; onSyncGitHub: () => void; onOpenProfile: () => void }) {
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

export function ProjectSection({ projects, busyKey, readmeBusyKey, aiReady, onAdd, onUpdate, onSave, onDelete, onGenerateReadmeBullets }: {
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

export function ProjectCard({ project, busy, readmeBusy, aiReady, onUpdate, onSave, onDelete, onGenerateReadmeBullets }: {
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

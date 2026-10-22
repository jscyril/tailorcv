import { useEffect, useState } from "react";
import { ListAutomaticBackups } from "../../../wailsjs/go/main/App";
import { domain } from "../../../wailsjs/go/models";
import { EducationDraft } from "../../lib/education";
import { ExperienceDraft } from "../../lib/experience";
import { Profile } from "../../lib/profile";
import { ProjectDraft } from "../../lib/project";
import { type Application } from "../../lib/resume";
import { Icon, PanelHeader } from "../../components/ui";
import { type Job } from "../ai/AIWorkspace";

export function DataWorkspace({ profile, experiences, projects, educations, jobs, applications, busy, lastResult, onExport, onImport, onRestoreAutomatic }: { profile: Profile; experiences: ExperienceDraft[]; projects: ProjectDraft[]; educations: EducationDraft[]; jobs: Job[]; applications: Application[]; busy: "export" | "import" | ""; lastResult: domain.BackupResult | null; onExport: () => void; onImport: () => void; onRestoreAutomatic: (backup: domain.AutomaticBackup) => void }) {
  const evidenceCount = experiences.reduce((sum, experience) => sum + experience.bullets.length, 0) + projects.reduce((sum, project) => sum + project.bullets.length, 0);
  const versionCount = applications.reduce((sum, application) => sum + application.versions.length, 0);
  return <section className="workspace-panel scroll-panel data-workspace"><PanelHeader eyebrow="Local data" title="Backup & restore" description="Keep a portable, versioned copy of your complete TailorCV profile." /><div className="backup-content"><section className="backup-summary"><header><span className="empty-icon"><Icon name="database" size={22} /></span><div><h2>Current local profile</h2><p>Everything listed here is included in one JSON backup.</p></div></header><div className="backup-stat-grid"><div><strong>{profile.name ? "1" : "0"}</strong><span>profile</span></div><div><strong>{experiences.length}</strong><span>roles</span></div><div><strong>{projects.length}</strong><span>projects</span></div><div><strong>{educations.length}</strong><span>education</span></div><div><strong>{jobs.length}</strong><span>jobs</span></div><div><strong>{applications.length}</strong><span>applications</span></div><div><strong>{versionCount}</strong><span>versions</span></div><div><strong>{profile.skills.length}</strong><span>skills</span></div><div><strong>{evidenceCount}</strong><span>evidence</span></div></div></section><section className="backup-action-card"><div><span className="backup-action-icon"><Icon name="download" size={20} /></span><h2>Export backup</h2><p>Write an owner-readable JSON snapshot using a native save dialog. IDs, ordering, verification state, ranking priority, and timestamps are preserved.</p></div><button className="primary-button" disabled={busy !== ""} onClick={onExport}>{busy === "export" ? "Exporting…" : "Choose destination"}</button></section><section className="backup-action-card restore"><div><span className="backup-action-icon"><Icon name="refresh" size={20} /></span><h2>Restore backup</h2><p>The entire file is validated before a single transaction replaces current data. Invalid or unsupported backups leave this profile untouched.</p></div><button className="secondary-button" disabled={busy !== ""} onClick={onImport}>{busy === "import" ? "Restoring…" : "Choose backup"}</button></section>{lastResult && <section className="backup-result"><span className="status-dot" /><div><strong>Last operation completed</strong><p>{lastResult.certificationCount} certifications · {lastResult.achievementCount} achievements · {lastResult.resumeVersionCount} resume versions</p><small>{lastResult.path}</small></div></section>}<AutomaticBackups busy={busy !== ""} refreshKey={lastResult} onRestore={onRestoreAutomatic} /><div className="backup-safety"><strong>Backup format v9</strong><p>Backups include evidence ranking priorities, professional links, certifications, achievements, GitHub metadata, templates, AI audit history, and non-secret preferences, but never provider credentials, generated PDFs, compiler caches, or local model data.</p></div></div></section>;
}

function AutomaticBackups({ busy, refreshKey, onRestore }: { busy: boolean; refreshKey: unknown; onRestore: (backup: domain.AutomaticBackup) => void }) {
  const [backups, setBackups] = useState<domain.AutomaticBackup[] | null>(null);
  const [loadError, setLoadError] = useState("");
  // Reload after each export or restore, since a restore adds a safety copy.
  useEffect(() => {
    ListAutomaticBackups().then((items) => { setBackups(items ?? []); setLoadError(""); }).catch((reason) => setLoadError(String(reason)));
  }, [refreshKey]);
  return <section className="automatic-backups" aria-labelledby="automatic-backups-title">
    <h2 id="automatic-backups-title">Automatic backups</h2>
    <p>TailorCV saves one backup a day (keeping the last 7) and a safety copy before every restore (keeping the last 5), so a restore can be undone.</p>
    {loadError ? <p role="alert">Could not list automatic backups: {loadError}</p>
      : backups === null ? <p>Loading…</p>
      : backups.length === 0 ? <p>No automatic backups yet. The first one is written shortly after TailorCV starts.</p>
      : <ul>{backups.map((backup) => {
        const label = `${backup.kind === "daily" ? "Daily backup" : "Before restore"} · ${new Date(backup.createdAt).toLocaleString()}`;
        return <li key={backup.name}><span>{label}</span><small>{Math.max(1, Math.round(backup.size / 1024))} KB</small><button className="text-button" disabled={busy} aria-label={`Restore ${label}`} onClick={() => onRestore(backup)}>Restore</button></li>;
      })}</ul>}
  </section>;
}

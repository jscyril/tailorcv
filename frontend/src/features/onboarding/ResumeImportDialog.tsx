import { useEffect, useRef, useState } from "react";
import { domain } from "../../../wailsjs/go/models";

type Section = "experiences" | "projects" | "educations";

// Shows what was read from a resume file and lets the user choose what to add.
export function ResumeImportDialog({ data, busy, onImport, onCancel, aiBusy = false, onReadWithAI }: {
  data: domain.ResumeImport;
  busy: boolean;
  aiBusy?: boolean;
  // Set only when an agent-capable provider is ready.
  onReadWithAI?: () => void;
  onImport: (selected: domain.ResumeImport) => void;
  onCancel: () => void;
}) {
  const dialogRef = useRef<HTMLElement>(null);
  const [excluded, setExcluded] = useState<Record<string, boolean>>({});
  useEffect(() => { dialogRef.current?.querySelector<HTMLButtonElement>("button.primary-button")?.focus(); }, []);

  const key = (section: Section, index: number) => `${section}:${index}`;
  const toggle = (id: string) => setExcluded((current) => ({ ...current, [id]: !current[id] }));
  const keep = <T,>(section: Section, items: T[]) => items.filter((_, index) => !excluded[key(section, index)]);
  const skills = excluded.skills ? [] : data.profile.skills;
  const selected = { ...data, profile: { ...data.profile, skills }, experiences: keep("experiences", data.experiences), projects: keep("projects", data.projects), educations: keep("educations", data.educations) };
  const count = selected.experiences.length + selected.projects.length + selected.educations.length;
  const profile = data.profile;
  const contact = [profile.name, profile.email, profile.phone, profile.linkedInUrl, profile.githubUsername && `github.com/${profile.githubUsername}`, profile.website].filter(Boolean);

  const item = (section: Section, index: number, title: string, detail: string, bullets?: number) => {
    const id = key(section, index);
    return <li key={id}><label><input type="checkbox" checked={!excluded[id]} onChange={() => toggle(id)} /><span><strong>{title}</strong><small>{detail}{bullets ? ` · ${bullets} bullet${bullets === 1 ? "" : "s"}` : ""}</small></span></label></li>;
  };
  // YYYY-MM as "Jun 2022", matching how the resume prints it.
  const month = (value?: string) => value ? new Date(`${value}-01T00:00:00`).toLocaleDateString("en-US", { month: "short", year: "numeric" }) : "";
  const dates = (start?: string, end?: string, current?: boolean) => [...new Set([month(start), current ? "Present" : month(end)])].filter(Boolean).join(" – ");

  return <div className="onboarding-overlay">
    <section ref={dialogRef} className="resume-import-dialog" role="dialog" aria-modal="true" aria-labelledby="resume-import-title" onKeyDown={(event) => { if (event.key === "Escape") onCancel(); }}>
      <header>
        <p className="eyebrow">Import resume</p>
        <h1 id="resume-import-title">Review what was found in {data.fileName}</h1>
        <p>Uncheck anything you do not want. Profile fields you already filled in are kept, and entries you already have are skipped. Imported bullets count as your own verified evidence, so check them later on each screen.</p>
      </header>
      <div className="resume-import-body">
        {contact.length > 0 && <section><h2>Profile</h2><p>{contact.join(" · ")}</p></section>}
        {data.experiences.length > 0 && <section><h2>Experience</h2><ul>{data.experiences.map((entry, index) => item("experiences", index, `${entry.title} · ${entry.company}`, [dates(entry.startDate, entry.endDate, entry.current), entry.location].filter(Boolean).join(" · "), entry.bullets.length))}</ul></section>}
        {data.projects.length > 0 && <section><h2>Projects</h2><ul>{data.projects.map((entry, index) => item("projects", index, entry.name, [dates(entry.startDate, entry.endDate, entry.ongoing), entry.skills.join(", ")].filter(Boolean).join(" · "), entry.bullets.length))}</ul></section>}
        {data.educations.length > 0 && <section><h2>Education</h2><ul>{data.educations.map((entry, index) => item("educations", index, entry.institution, [entry.degree, entry.fieldOfStudy, dates(entry.startDate, entry.endDate, entry.current)].filter(Boolean).join(" · ")))}</ul></section>}
        {data.profile.skills.length > 0 && <section><h2>Skills</h2><label className="resume-import-skills"><input type="checkbox" checked={!excluded.skills} onChange={() => toggle("skills")} /><span>{data.profile.skills.join(", ")}</span></label></section>}
        {data.warnings.length > 0 && <section className="resume-import-warnings" role="status"><h2>Not imported</h2><ul>{data.warnings.map((warning) => <li key={warning}>{warning}</li>)}</ul></section>}
      </div>
      <footer>
        <button type="button" className="text-button" onClick={onCancel}>Cancel</button>
        {onReadWithAI && <button type="button" className="secondary-button" disabled={busy || aiBusy} title="Sends this file to Gemini; entries must match the file word for word" onClick={onReadWithAI}>{aiBusy ? "Reading…" : "Read with AI instead"}</button>}
        <button type="button" className="primary-button" disabled={busy || (count === 0 && skills.length === 0 && contact.length === 0)} onClick={() => onImport(domain.ResumeImport.createFrom(selected))}>{busy ? "Importing…" : `Import ${count} entr${count === 1 ? "y" : "ies"}`}</button>
      </footer>
    </section>
  </div>;
}

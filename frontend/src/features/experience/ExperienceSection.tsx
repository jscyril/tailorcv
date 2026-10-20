import { FormEvent } from "react";
import { Experience, ExperienceDraft } from "../../lib/experience";
import { EvidenceEditor } from "../../components/EvidenceEditor";
import { Field } from "../../components/ui";
import { type Job } from "../ai/AIWorkspace";

export function ExperienceSection({ experiences, busyKey, onAdd, onUpdate, onSave, onDelete }: {
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

export function ExperienceCard({ experience, busy, onUpdate, onSave, onDelete }: {
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

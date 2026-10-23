import { domain } from "../../../wailsjs/go/models";
import { Icon, PanelHeader } from "../../components/ui";

export function TemplatesWorkspace({ templates, selectedID, busyKey, onImport, onUse, onEdit, onDuplicate, onDelete }: { templates: domain.ResumeTemplate[]; selectedID: string; busyKey: string; onImport: () => void; onUse: (id: string) => void; onEdit: (id: string) => void; onDuplicate: (template: domain.ResumeTemplate) => void; onDelete: (template: domain.ResumeTemplate) => void }) {
  return <section className="workspace-panel scroll-panel templates-workspace">
    <PanelHeader eyebrow="Document system" title="Templates" description="Choose a built-in layout or import a complete, single-file LaTeX template." action={<button className="primary-button" disabled={busyKey !== ""} onClick={onImport}><Icon name="download" size={16} />{busyKey === "import" ? "Importing…" : "Import .tex"}</button>} />
    <div className="template-guidance"><span className="empty-icon"><Icon name="template" size={22} /></span><div><strong>Optional TailorCV markers</strong><p>Imported documents compile as-is. Add markers such as <code>{"{{TAILORCV_NAME}}"}</code>, <code>{"{{TAILORCV_EXPERIENCE_SECTION}}"}</code>, or <code>{"{{TAILORCV_PROJECTS_SECTION}}"}</code> to populate saved profile data when the template is selected. Sections are written with <code>\tailorcvEntry</code>, <code>\tailorcvProject</code>, and <code>\tailorcvBullet</code>; define those in your preamble to restyle every entry, as the built-in Jake's Resume does.</p></div></div>
    <div className="template-grid">{templates.map((template) => {
      const selected = template.id === selectedID;
      const busy = busyKey === template.id;
      return <article className={`template-card ${selected ? "selected" : ""}`} key={template.id}>
        <div className="template-paper"><span>{template.name.split(/\s+/).slice(0, 2).map((part) => part[0]).join("")}</span><i /><i /><i /><i className="short" /></div>
        <div className="template-card-copy"><div className="template-title"><strong>{template.name}</strong><span>{template.builtIn ? "Built in" : "Custom"}</span></div><p>{template.description || "User-owned LaTeX resume template."}</p><small>{selected ? "Currently selected" : template.builtIn ? "Read-only · duplicate to customize" : "Stored locally"}</small></div>
        <div className="template-actions"><button className={selected ? "primary-button" : "secondary-button"} disabled={busy || selected} onClick={() => onUse(template.id)}>{selected ? "Selected" : busy ? "Loading…" : "Use"}</button><button className="text-button" disabled={busy} onClick={() => onEdit(template.id)}>Edit source</button><button className="text-button" disabled={busy} onClick={() => onDuplicate(template)}>Duplicate</button>{!template.builtIn && <button className="danger-button" disabled={busy} onClick={() => onDelete(template)}>Delete</button>}</div>
      </article>;
    })}</div>
  </section>;
}

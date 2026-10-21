import { Profile } from "../lib/profile";

export function NavButton({ active, label, icon, badge, onClick }: { active: boolean; label: string; icon: IconName; badge?: string | number; onClick: () => void }) {
  return (
    <button className={`nav-button ${active ? "active" : ""}`} aria-current={active ? "page" : undefined} onClick={onClick}>
      <Icon name={icon} />
      <span className="nav-button-label">{label}</span>
      {badge !== undefined && <span className="nav-badge">{badge}</span>}
    </button>
  );
}

export type IconName = "home" | "user" | "briefcase" | "folder" | "education" | "sparkles" | "template" | "code" | "target" | "chat" | "download" | "refresh" | "check" | "search" | "database" | "checklist" | "settings";

export const iconPaths: Record<IconName, React.ReactNode> = {
  home: <><path d="m3 10 9-7 9 7" /><path d="M5 9v11h14V9" /><path d="M9 20v-6h6v6" /></>,
  user: <><circle cx="12" cy="8" r="4" /><path d="M4 21c.8-4.2 3.5-6 8-6s7.2 1.8 8 6" /></>,
  briefcase: <><rect x="3" y="7" width="18" height="13" rx="2" /><path d="M8 7V4h8v3M3 12h18M10 12v2h4v-2" /></>,
  folder: <><path d="M3 6h7l2 2h9v11H3z" /><path d="M3 10h18" /></>,
  education: <><path d="m2 9 10-5 10 5-10 5z" /><path d="M6 11v5c3 2 9 2 12 0v-5M22 9v7" /></>,
  sparkles: <><path d="m12 2 1.5 5.5L19 9l-5.5 1.5L12 16l-1.5-5.5L5 9l5.5-1.5z" /><path d="m19 15 .8 2.2L22 18l-2.2.8L19 21l-.8-2.2L16 18l2.2-.8z" /></>,
  template: <><rect x="4" y="3" width="16" height="18" rx="2" /><path d="M8 7h8M8 11h8M8 15h5" /></>,
  code: <><path d="m8 5-6 7 6 7M16 5l6 7-6 7M14 3l-4 18" /></>,
  target: <><circle cx="12" cy="12" r="9" /><circle cx="12" cy="12" r="4" /><path d="M12 3v3M21 12h-3M12 21v-3M3 12h3" /></>,
  chat: <><path d="M4 4h16v13H8l-4 4z" /><path d="M8 9h8M8 13h5" /></>,
  download: <><path d="M12 3v12M7 10l5 5 5-5M4 21h16" /></>,
  refresh: <><path d="M20 7V3h-4M4 17v4h4" /><path d="M19 10a7 7 0 0 0-12-4L4 9M5 14a7 7 0 0 0 12 4l3-3" /></>,
  check: <path d="m5 12 4 4L19 6" />,
  search: <><circle cx="10" cy="10" r="6" /><path d="m15 15 5 5" /></>,
  checklist: <><path d="m4 6 2 2 3-3M4 13l2 2 3-3M4 20h.01M12 7h8M12 14h8M12 20h8" /></>,
  settings: <><circle cx="12" cy="12" r="3" /><path d="M12 2v3M12 19v3M4.9 4.9l2.1 2.1M17 17l2.1 2.1M2 12h3M19 12h3M4.9 19.1 7 17M17 7l2.1-2.1" /></>,
  database: <><ellipse cx="12" cy="5" rx="8" ry="3" /><path d="M4 5v6c0 1.7 3.6 3 8 3s8-1.3 8-3V5M4 11v6c0 1.7 3.6 3 8 3s8-1.3 8-3v-6" /></>,
};

export function Icon({ name, size = 18 }: { name: IconName; size?: number }) {
  return <svg className="icon" width={size} height={size} viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true">{iconPaths[name]}</svg>;
}

export function TopToolbar({ profile, templateName, compiling, canExport, onCompile, onExport }: { profile: Profile; templateName: string; compiling: boolean; canExport: boolean; onCompile: () => void; onExport: () => void }) {
  return <header className="top-toolbar">
    <div className="toolbar-brand"><span className="brand-mark">T</span><strong>TailorCV</strong></div>
    <span className="toolbar-divider" />
    <div className="document-title"><strong>{profile.headline || "Backend Engineer Resume"}</strong><small>{templateName} · profile saved locally</small></div>
    <div className="toolbar-spacer" />
    <div className="ats-pill"><span className={`status-dot ${canExport ? "" : "muted"}`} /> {canExport ? "PDF ready" : "Source draft"}</div>
    <button className="toolbar-button" disabled={compiling} onClick={onCompile}><Icon name="refresh" size={16} />{compiling ? "Compiling…" : "Compile"}</button>
    <button className="export-button" disabled={!canExport || compiling} title={canExport ? "Export the latest compiled PDF" : "Compile the current source before exporting"} onClick={onExport}><Icon name="download" size={16} />Export PDF</button>
  </header>;
}

export function PanelHeader({ eyebrow, title, description, action }: { eyebrow: string; title: string; description: string; action?: React.ReactNode }) {
  return <header className="panel-header"><div><p className="eyebrow">{eyebrow}</p><h1>{title}</h1><p>{description}</p></div>{action}</header>;
}

export function FormBlock({ title, description, children }: { title: string; description: string; children: React.ReactNode }) {
  return <section className="form-block"><header><h2>{title}</h2><p>{description}</p></header><div>{children}</div></section>;
}

export function FormSection({ title, description, children }: { title: string; description: string; children: React.ReactNode }) {
  return <section className="form-section"><div className="section-copy"><h2>{title}</h2><p>{description}</p></div><div className="section-fields">{children}</div></section>;
}

export function Field({ label, value, placeholder, type = "text", prefix, disabled = false, required = false, onChange }: { label: string; value: string; placeholder: string; type?: string; prefix?: string; disabled?: boolean; required?: boolean; onChange: (value: string) => void }) {
  return <label className="field"><span>{label}</span><div className={prefix ? "prefixed-input" : ""}>{prefix && <em>{prefix}</em>}<input type={type} value={value} placeholder={placeholder} disabled={disabled} required={required} onChange={(event) => onChange(event.target.value)} /></div></label>;
}

export function EvidenceState({value,onChange}:{value:"verified"|"unverified";onChange:(value:"verified"|"unverified")=>void}){return <label className="field"><span>Review state</span><select value={value} onChange={event=>onChange(event.target.value as "verified"|"unverified")}><option value="unverified">Needs review</option><option value="verified">Verified</option></select></label>}

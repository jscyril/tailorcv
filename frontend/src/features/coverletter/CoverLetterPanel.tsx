import { useState } from "react";
import { ClipboardSetText } from "../../../wailsjs/runtime/runtime";
import { type Application } from "../../lib/resume";

export function CoverLetterPanel({ application, title, aiModel, busy, onDraft, onGenerate, onSave, onExport }: {
  application: Application;
  title: string;
  aiModel: string;
  busy: boolean;
  onDraft: () => Promise<string | null>;
  onGenerate: () => Promise<string | null>;
  onSave: (text: string) => Promise<boolean>;
  onExport: (text: string) => void;
}) {
  const saved = application.coverLetter ?? "";
  // Unsaved text; null means the field shows the saved letter.
  const [edited, setEdited] = useState<string | null>(null);
  const [copied, setCopied] = useState(false);
  const text = edited ?? saved;
  const dirty = edited !== null && edited !== saved;
  const id = `cover-letter-${application.id}`;

  const replaceWith = async (draft: () => Promise<string | null>) => {
    if (text.trim() && !window.confirm("Replace the current cover letter with a new draft?")) return;
    const result = await draft();
    if (result !== null) setEdited(result);
  };
  const copy = async () => {
    setCopied(await ClipboardSetText(text).catch(() => false));
    window.setTimeout(() => setCopied(false), 2000);
  };

  return <section className="cover-letter" aria-label={`Cover letter for ${title}`}>
    <div className="cover-letter-actions">
      <button type="button" className="secondary-button" disabled={busy} onClick={() => replaceWith(onDraft)}>Draft from evidence</button>
      <button type="button" className="secondary-button" disabled={busy || !aiModel} title={aiModel ? `Uses ${aiModel}; paragraphs are checked against your selected evidence` : "Set up a model in AI assistant first"} onClick={() => replaceWith(onGenerate)}>{busy ? "Working…" : "Draft with AI"}</button>
    </div>
    <label htmlFor={id} className="visually-hidden">Cover letter text</label>
    <textarea id={id} rows={14} maxLength={20000} value={text} placeholder="Draft a letter from your selected evidence, or write one here." onChange={(event) => setEdited(event.target.value)} />
    <footer>
      <button type="button" className="primary-button" disabled={busy || !dirty} onClick={async () => { if (await onSave(text)) setEdited(null); }}>Save letter</button>
      {dirty && <button type="button" className="text-button" onClick={() => setEdited(null)}>Undo changes</button>}
      <span />
      <button type="button" className="secondary-button" disabled={!text.trim()} onClick={copy}>{copied ? "Copied" : "Copy text"}</button>
      <button type="button" className="secondary-button" disabled={busy || !text.trim()} onClick={() => onExport(text)}>Export PDF</button>
    </footer>
    {dirty && <small>Drafts are not saved until you choose Save letter.</small>}
  </section>;
}

import { EvidenceBullet, moveItem, newEvidenceBullet } from "../lib/experience";
import { Field } from "./ui";

export function EvidenceEditor({ bullets, emptyLabel, onChange }: {
  bullets: EvidenceBullet[];
  emptyLabel: string;
  onChange: (bullets: EvidenceBullet[]) => void;
}) {
  const updateBullet = (index: number, patch: Partial<EvidenceBullet>) => {
    onChange(bullets.map((bullet, bulletIndex) => bulletIndex === index ? { ...bullet, ...patch } : bullet));
  };
  const removeBullet = (index: number) => {
    onChange(bullets.filter((_, bulletIndex) => bulletIndex !== index));
  };

  return (
    <>
      <div className="evidence-heading">
        <div><strong>Evidence bullets</strong><span>Use specific, defensible work and outcomes.</span></div>
        <button className="text-button" type="button" onClick={() => onChange([...bullets, newEvidenceBullet()])}>+ Add bullet</button>
      </div>
      <div className="evidence-list">
        {bullets.length === 0 && <p className="evidence-empty">No evidence bullets. {emptyLabel}</p>}
        {bullets.map((bullet, index) => (
          <div className="evidence-row" key={bullet.id || `new-bullet-${index}`}>
            <div className="evidence-order" aria-label={`Reorder evidence bullet ${index + 1}`}>
              <span>{index + 1}</span>
              <button type="button" aria-label="Move bullet up" disabled={index === 0} onClick={() => onChange(moveItem(bullets, index, index - 1))}>↑</button>
              <button type="button" aria-label="Move bullet down" disabled={index === bullets.length - 1} onClick={() => onChange(moveItem(bullets, index, index + 1))}>↓</button>
            </div>
            <div className="evidence-fields">
              <label className="field full">
                <span>Claim or outcome</span>
                <textarea rows={3} maxLength={1200} required value={bullet.text} onChange={(event) => updateBullet(index, { text: event.target.value })} placeholder="Reduced deployment time by 40% by replacing manual release steps with an audited pipeline." />
                <small>{bullet.text.length}/1200</small>
              </label>
              <div className="evidence-metadata">
                <label className="field">
                  <span>Review state</span>
                  <select value={bullet.verification} onChange={(event) => updateBullet(index, { verification: event.target.value as EvidenceBullet["verification"] })}>
                    <option value="unverified">Needs review</option>
                    <option value="verified">Verified</option>
                  </select>
                </label>
                <label className="field">
                  <span>Ranking priority</span>
                  <select value={bullet.importance} onChange={(event) => updateBullet(index, { importance: event.target.value as EvidenceBullet["importance"] })}>
                    <option value="standard">Standard</option>
                    <option value="important">Important (+6)</option>
                    <option value="essential">Essential (+12)</option>
                  </select>
                </label>
                <Field label="Source URL (optional)" type="url" value={bullet.sourceUrl} onChange={(value) => updateBullet(index, { sourceUrl: value })} placeholder="https://github.com/…" />
                <div className="provenance-field"><span>Origin</span><strong>{bullet.provenance}</strong></div>
                <button className="remove-button" type="button" onClick={() => removeBullet(index)}>Remove</button>
              </div>
            </div>
          </div>
        ))}
      </div>
    </>
  );
}

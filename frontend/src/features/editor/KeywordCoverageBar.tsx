import { domain } from "../../../wailsjs/go/models";

// Shows which of the job's skills the resume source mentions.
export function KeywordCoverageBar({ coverage }: { coverage: domain.KeywordCoverage }) {
  const items = [...coverage.required.map((item) => ({ ...item, preferred: false })), ...coverage.preferred.map((item) => ({ ...item, preferred: true }))];
  if (!items.length) return null;
  const present = items.filter((item) => item.present).length;
  const missing = items.length - present;
  return <section className="keyword-coverage" aria-label="Job keyword coverage">
    <strong>{present} of {items.length} job skills in this resume</strong>
    {missing > 0 && <span>Missing ones are only worth adding if your evidence supports them.</span>}
    <ul>{items.map((item) => <li key={item.skill} className={item.present ? "present" : "missing"}>
      <span aria-hidden="true">{item.present ? "✓" : "–"}</span>{item.skill}{item.preferred && <small> preferred</small>}<span className="visually-hidden">{item.present ? " (in resume)" : " (missing)"}</span>
    </li>)}</ul>
  </section>;
}

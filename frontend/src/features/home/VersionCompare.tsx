import { useMemo, useState } from "react";
import { diffLines, type DiffLine } from "../../lib/diff";
import { type ResumeVersion } from "../../lib/resume";

const CONTEXT = 2;

// Keeps changed lines plus CONTEXT unchanged lines around them; null marks a skipped gap.
function withContext(lines: DiffLine[]): (DiffLine | null)[] {
  const keep = lines.map(() => false);
  lines.forEach((line, index) => {
    if (line.kind === "same") return;
    for (let k = Math.max(0, index - CONTEXT); k <= Math.min(lines.length - 1, index + CONTEXT); k++) keep[k] = true;
  });
  const out: (DiffLine | null)[] = [];
  lines.forEach((line, index) => {
    if (keep[index]) out.push(line);
    else if (out.length && out[out.length - 1] !== null) out.push(null);
  });
  if (out[out.length - 1] === null) out.pop();
  return out;
}

export function VersionCompare({ versions, title, onClose }: { versions: ResumeVersion[]; title: string; onClose: () => void }) {
  // versions arrive newest first.
  const [toId, setToId] = useState(versions[0].id);
  const [fromId, setFromId] = useState(versions[1].id);
  const from = versions.find((version) => version.id === fromId) ?? versions[1];
  const to = versions.find((version) => version.id === toId) ?? versions[0];
  const lines = useMemo(() => diffLines(from.latexSource, to.latexSource), [from, to]);
  const added = lines.filter((line) => line.kind === "added").length;
  const removed = lines.filter((line) => line.kind === "removed").length;
  const option = (version: ResumeVersion) => <option key={version.id} value={version.id}>Version {version.versionNumber} · {new Date(version.createdAt).toLocaleDateString()}</option>;
  return <section className="version-compare" aria-label={`Compare versions of ${title}`}>
    <header>
      <label>From <select value={from.id} onChange={(event) => setFromId(event.target.value)}>{versions.map(option)}</select></label>
      <label>To <select value={to.id} onChange={(event) => setToId(event.target.value)}>{versions.map(option)}</select></label>
      <span className="version-compare-summary">{added || removed ? <><b className="added">+{added}</b> <b className="removed">−{removed}</b> lines</> : "No changes"}</span>
      <button className="text-button" type="button" onClick={onClose}>Close</button>
    </header>
    {(added > 0 || removed > 0) && <pre className="version-diff">{withContext(lines).map((line, index) => line
      ? <div key={index} className={line.kind}><span aria-hidden="true">{line.kind === "added" ? "+" : line.kind === "removed" ? "−" : " "}</span><span className="visually-hidden">{line.kind === "same" ? "" : `${line.kind}: `}</span>{line.text || " "}</div>
      : <div key={index} className="gap">⋯</div>)}</pre>}
  </section>;
}

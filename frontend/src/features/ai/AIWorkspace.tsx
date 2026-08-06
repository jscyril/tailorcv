export type AIProvider = "ollama" | "gemini";

export type Job = {
  id: string;
  company: string;
  role: string;
  description: string;
  createdAt: string;
  updatedAt: string;
};

export type EvidenceMatch = {
  factId: string;
  sourceId: string;
  sourceType: "experience" | "project";
  sourceLabel: string;
  text: string;
  score: number;
  matchedSkills: string[];
  reasons: string[];
  verified: boolean;
  selectable: boolean;
};

export type JobAnalysis = {
  job: Job;
  score: number;
  matchedSkills: string[];
  unmentionedSkills: string[];
  detectedSkills: string[];
  requiredSkills: string[];
  preferredSkills: string[];
  responsibilities: string[];
  keywords: string[];
  searchTerms: string[];
  rankedEvidence: EvidenceMatch[];
  explanation: string;
};

export type AIProposal = {
  targetFactId: string;
  supportingFactIds: string[];
  text: string;
};

export type AIRun = {
  id: string;
  jobId: string;
  provider: string;
  model: string;
  promptVersion: string;
  schemaVersion: string;
  selectedFactIds: string[];
  validationPassed: boolean;
  failureCategory: string;
  validationErrors: string[];
  proposals: AIProposal[];
  resumeVersionId: string;
  createdAt: string;
  acceptedAt: string;
};

export type AIProviderStatus = {
  provider: string;
  endpoint: string;
  available: boolean;
  models: string[];
  message: string;
};

export type CredentialStatus = {
  configured: boolean;
  message: string;
};

export type AIBusyState = "check" | "credential" | "generate" | "accept" | "";

export type AIWorkspaceProps = {
  provider: AIProvider;
  endpoint: string;
  status: AIProviderStatus | null;
  model: string;
  credential: CredentialStatus;
  apiKey: string;
  run: AIRun | null;
  runs: AIRun[];
  proposals: AIProposal[];
  acceptedProposalIDs: string[];
  job: Job;
  analysis: JobAnalysis | null;
  selectedFactIDs: string[];
  templateName: string;
  busy: AIBusyState;
  onProviderChange: (value: AIProvider) => void;
  onEndpointChange: (value: string) => void;
  onModelChange: (value: string) => void;
  onAPIKeyChange: (value: string) => void;
  onSaveCredential: () => void;
  onDeleteCredential: () => void;
  onCheck: () => void;
  onGenerate: () => void;
  onCancel: () => void;
  onProposalChange: (targetFactID: string, text: string) => void;
  onToggleProposal: (targetFactID: string) => void;
  onAccept: () => void;
  onOpenJob: () => void;
};

export function AIWorkspace({ provider, endpoint, status, model, credential, apiKey, run, runs, proposals, acceptedProposalIDs, job, analysis, selectedFactIDs, templateName, busy, onProviderChange, onEndpointChange, onModelChange, onAPIKeyChange, onSaveCredential, onDeleteCredential, onCheck, onGenerate, onCancel, onProposalChange, onToggleProposal, onAccept, onOpenJob }: AIWorkspaceProps) {
  const canGenerate = Boolean(status?.available && model && job.id && analysis && selectedFactIDs.length);
  const selectedProposals = proposals.filter((proposal) => acceptedProposalIDs.includes(proposal.targetFactId));
  return (
    <section className="workspace-panel ai-workspace ai-tailoring-workspace">
      <header className="panel-header"><div><p className="eyebrow">Evidence-constrained tailoring</p><h1>{provider === "ollama" ? "Ollama" : "Gemini"} review</h1><p>Rewrite selected facts, validate every claim, and accept only wording you approve.</p></div></header>
      <div className="ai-setup-grid">
        <section className="ai-setup-card">
          <div className="ai-provider-switch" role="group" aria-label="AI provider"><button className={provider === "ollama" ? "active" : ""} disabled={busy !== ""} onClick={() => onProviderChange("ollama")}>Ollama</button><button className={provider === "gemini" ? "active" : ""} disabled={busy !== ""} onClick={() => onProviderChange("gemini")}>Gemini</button></div>
          <header><span className="provider-logo">{provider === "ollama" ? "OL" : "GM"}</span><div><strong>{provider === "ollama" ? "Ollama" : "Google Gemini"}</strong><small>{provider === "ollama" ? "Local provider · no credential required" : "Cloud provider · API key stays in the OS keyring"}</small></div><span className={`status-dot ${status?.available ? "" : "muted"}`} /></header>
          {provider === "ollama" ? <label className="field"><span>Endpoint</span><input value={endpoint} onChange={(event) => onEndpointChange(event.target.value)} placeholder="http://127.0.0.1:11434" /></label> : <div className="gemini-credential"><label className="field"><span>Gemini API key</span><input type="password" autoComplete="off" value={apiKey} onChange={(event) => onAPIKeyChange(event.target.value)} placeholder={credential.configured ? "Stored securely · enter a replacement" : "Enter API key"} /></label><div><button className="secondary-button" disabled={!apiKey.trim() || busy !== ""} onClick={onSaveCredential}>{busy === "credential" ? "Saving…" : credential.configured ? "Replace key" : "Save to keyring"}</button>{credential.configured && <button className="danger-button" disabled={busy !== ""} onClick={onDeleteCredential}>Remove key</button>}</div><small className={credential.configured ? "ai-status-ok" : "ai-status-error"}>{credential.message}</small></div>}
          <button className="secondary-button" disabled={busy !== ""} onClick={onCheck}>{busy === "check" ? "Checking…" : "Check connection"}</button>
          {status && <p className={status.available ? "ai-status-ok" : "ai-status-error"}>{status.message}</p>}
          <label className="field"><span>{provider === "ollama" ? "Local model" : "Gemini model"}</span><select value={model} disabled={!status?.models.length || busy !== ""} onChange={(event) => onModelChange(event.target.value)}><option value="">Select a model</option>{status?.models.map((item) => <option value={item} key={item}>{item}</option>)}</select></label>
        </section>
        <section className="ai-context-card">
          <span className="step-pill">Current context</span>
          <h2>{job.role || "Analyze a job first"}</h2>
          <p>{job.company || "No company selected"}</p>
          <div><strong>{selectedFactIDs.length}</strong><span>selected evidence facts</span></div>
          {!analysis ? <button className="secondary-button" onClick={onOpenJob}>Open job matching</button> : <small>Only normalized requirements and these selected facts will be sent to {provider === "ollama" ? "Ollama" : "Gemini"}.</small>}
          <div className="ai-generation-actions"><button className="primary-button" disabled={!canGenerate || busy !== ""} onClick={onGenerate}>{busy === "generate" ? "Generating…" : run ? "Retry tailoring" : "Generate proposals"}</button>{busy === "generate" && <button className="danger-button" onClick={onCancel}>Cancel</button>}</div>
        </section>
      </div>

      {run && !run.validationPassed && <section className="ai-validation-failure"><strong>Run blocked by {run.failureCategory || "validation"}</strong><p>No resume source was changed. The failure was recorded without storing credentials.</p><ul>{run.validationErrors.map((item) => <li key={item}>{item}</li>)}</ul></section>}

      {run?.validationPassed && <section className="ai-review-section">
        <header><div><p className="eyebrow">Review gate</p><h2>Original evidence and proposed wording</h2><p>Edits are revalidated before a version can be saved.</p></div><span>{acceptedProposalIDs.length}/{proposals.length} accepted</span></header>
        <div className="ai-proposal-list">{proposals.map((proposal) => {
          const original = analysis?.rankedEvidence.find((evidence) => evidence.factId === proposal.targetFactId);
          const accepted = acceptedProposalIDs.includes(proposal.targetFactId);
          return <article className={accepted ? "accepted" : ""} key={proposal.targetFactId}><header><label><input type="checkbox" checked={accepted} onChange={() => onToggleProposal(proposal.targetFactId)} /><span>{accepted ? "Include proposal" : "Keep original"}</span></label><small>{proposal.supportingFactIds.length} cited fact{proposal.supportingFactIds.length === 1 ? "" : "s"}</small></header><div className="ai-comparison"><div><span>Original evidence</span><strong>{original?.sourceLabel ?? "Selected evidence"}</strong><p>{original?.text ?? "The source fact remains stored locally and unchanged."}</p></div><label><span>Proposed wording</span><textarea aria-label={`Proposed wording for ${proposal.targetFactId}`} rows={5} value={proposal.text} disabled={!accepted || busy !== ""} onChange={(event) => onProposalChange(proposal.targetFactId, event.target.value)} /><small>Citations: {proposal.supportingFactIds.join(", ")}</small></label></div></article>;
        })}</div>
        <footer><div><strong>{selectedProposals.length} proposals ready</strong><span>Accept into a new immutable version using {templateName}.</span></div><button className="primary-button" disabled={!selectedProposals.length || busy !== "" || Boolean(run.resumeVersionId)} onClick={onAccept}>{run.resumeVersionId ? "Already accepted" : busy === "accept" ? "Validating & saving…" : "Accept new resume version"}</button></footer>
      </section>}
      {runs.length > 0 && <section className="ai-run-history"><header><strong>AI run history</strong><span>{runs.length}</span></header><div>{runs.slice(0, 8).map((item) => <article key={item.id}><div><strong>{item.model}</strong><span>{item.validationPassed ? item.resumeVersionId ? "Accepted" : "Validated" : `Blocked · ${item.failureCategory || "validation"}`}</span></div><small>{item.selectedFactIds.length} facts · {new Date(item.createdAt).toLocaleString()}</small></article>)}</div></section>}
    </section>
  );
}

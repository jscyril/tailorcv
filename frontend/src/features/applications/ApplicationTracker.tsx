import { FormEvent, useState } from "react";
import { BrowserOpenURL } from "../../../wailsjs/runtime/runtime";
import { type Application } from "../../lib/resume";
import { ApplicationStatusControl, type ApplicationStatus } from "./ApplicationStatusControl";

export type ApplicationDetails = { appliedOn: string; followUpOn: string; postingUrl: string; notes: string };

const today = () => new Date().toLocaleDateString("en-CA"); // YYYY-MM-DD in local time

// A follow-up is due once its date arrives while the application is still open.
export function followUpDue(application: Application): boolean {
  return Boolean(application.followUpOn) && application.followUpOn! <= today() && (application.status === "submitted" || application.status === "interviewing");
}

export function ApplicationTracker({ application, title, busy, onStatusChange, onSave }: {
  application: Application;
  title: string;
  busy: boolean;
  onStatusChange: (status: ApplicationStatus) => void;
  onSave: (details: ApplicationDetails) => Promise<boolean>;
}) {
  const saved: ApplicationDetails = { appliedOn: application.appliedOn ?? "", followUpOn: application.followUpOn ?? "", postingUrl: application.postingUrl ?? "", notes: application.notes ?? "" };
  // Only fields the user edited are kept here, so values the app changes
  // (like the applied date set on submit) still show through.
  const [edits, setEdits] = useState<Partial<ApplicationDetails>>({});
  const draft = { ...saved, ...edits };
  const dirty = (Object.keys(saved) as Array<keyof ApplicationDetails>).some((key) => draft[key] !== saved[key]);
  const field = (key: keyof ApplicationDetails) => ({ value: draft[key], onChange: (event: { target: { value: string } }) => setEdits({ ...edits, [key]: event.target.value }) });
  const submit = async (event: FormEvent) => {
    event.preventDefault();
    if (await onSave(draft)) setEdits({});
  };
  const id = application.id;
  return <form className="application-tracker" aria-label={`Tracking for ${title}`} onSubmit={submit}>
    <ApplicationStatusControl status={application.status} busy={busy} onChange={onStatusChange} />
    <div className="application-tracker-grid">
      <label htmlFor={`applied-${id}`}>Applied on<input id={`applied-${id}`} type="date" {...field("appliedOn")} /></label>
      <label htmlFor={`follow-up-${id}`}>Follow up on<input id={`follow-up-${id}`} type="date" {...field("followUpOn")} /></label>
      <label htmlFor={`posting-${id}`} className="wide">Job posting link
        <span className="application-tracker-link"><input id={`posting-${id}`} type="url" placeholder="https://…" {...field("postingUrl")} />{saved.postingUrl && <button type="button" className="text-button" onClick={() => BrowserOpenURL(saved.postingUrl)}>Open</button>}</span>
      </label>
      <label htmlFor={`notes-${id}`} className="wide">Notes<textarea id={`notes-${id}`} rows={3} maxLength={10000} placeholder="Recruiter, interview dates, salary range…" {...field("notes")} /></label>
    </div>
    <footer><button className="primary-button" type="submit" disabled={busy || !dirty}>Save tracking</button>{dirty && <button className="text-button" type="button" onClick={() => setEdits({})}>Undo changes</button>}</footer>
  </form>;
}

import { FormEvent, KeyboardEvent, useEffect, useRef } from "react";

export type OnboardingField = "name" | "email" | "headline" | "location" | "githubUsername" | "linkedInUrl";

export interface OnboardingProfile {
  name: string;
  email: string;
  headline: string;
  location: string;
  githubUsername: string;
  linkedInUrl: string;
}

interface OnboardingProps {
  profile: OnboardingProfile;
  skillsText: string;
  busy: boolean;
  onChange: (field: OnboardingField, value: string) => void;
  onSkillsChange: (value: string) => void;
  onSubmit: (event: FormEvent) => void;
  onSkip: () => void;
}

const focusableSelector = "button:not([disabled]), input:not([disabled]), textarea:not([disabled]), select:not([disabled]), [tabindex]:not([tabindex='-1'])";

export function Onboarding({ profile, skillsText, busy, onChange, onSkillsChange, onSubmit, onSkip }: OnboardingProps) {
  const dialogRef = useRef<HTMLElement>(null);

  useEffect(() => {
    const previousFocus = document.activeElement instanceof HTMLElement ? document.activeElement : null;
    dialogRef.current?.querySelector<HTMLInputElement>("input")?.focus();
    return () => previousFocus?.focus();
  }, []);

  const containFocus = (event: KeyboardEvent<HTMLElement>) => {
    if (event.key !== "Tab" || !dialogRef.current) return;
    const controls = Array.from(dialogRef.current.querySelectorAll<HTMLElement>(focusableSelector));
    if (controls.length === 0) return;
    const first = controls[0];
    const last = controls[controls.length - 1];
    if (event.shiftKey && document.activeElement === first) {
      event.preventDefault();
      last.focus();
    } else if (!event.shiftKey && document.activeElement === last) {
      event.preventDefault();
      first.focus();
    }
  };

  return (
    <div className="onboarding-overlay">
      <section ref={dialogRef} className="onboarding-card" role="dialog" aria-modal="true" aria-labelledby="onboarding-title" aria-describedby="onboarding-description" onKeyDown={containFocus}>
        <aside>
          <span className="brand-mark">T</span>
          <p className="eyebrow">Welcome to TailorCV</p>
          <h1 id="onboarding-title">Start with facts you control.</h1>
          <p id="onboarding-description">Create your local career profile. TailorCV uses it as the source of truth for every resume.</p>
          <div className="onboarding-promise"><span className="status-dot" /><div><strong>Stored on this device</strong><small>No account or cloud sync required.</small></div></div>
        </aside>
        <form onSubmit={onSubmit}>
          <header><div><span>01</span><p>Profile foundation</p></div><small>About 2 minutes</small></header>
          <div className="onboarding-fields">
            <div className="field-grid two">
              <Field label="Full name" value={profile.name} onChange={(value) => onChange("name", value)} placeholder="Ada Lovelace" required />
              <Field label="Email" type="email" value={profile.email} onChange={(value) => onChange("email", value)} placeholder="ada@example.com" required />
              <Field label="Professional headline" value={profile.headline} onChange={(value) => onChange("headline", value)} placeholder="Backend engineer building reliable systems" required />
              <Field label="Location" value={profile.location} onChange={(value) => onChange("location", value)} placeholder="Bengaluru, India" />
              <Field label="GitHub username" value={profile.githubUsername} onChange={(value) => onChange("githubUsername", value)} placeholder="octocat" prefix="github.com/" />
              <Field label="LinkedIn URL" type="url" value={profile.linkedInUrl} onChange={(value) => onChange("linkedInUrl", value)} placeholder="https://linkedin.com/in/..." />
            </div>
            <label className="field"><span>Core skills</span><textarea rows={3} value={skillsText} onChange={(event) => onSkillsChange(event.target.value)} placeholder="Go, TypeScript, PostgreSQL, Docker" /><small>Separate skills with commas.</small></label>
          </div>
          <footer><button className="text-button" type="button" onClick={onSkip}>Explore first</button><button className="primary-button" disabled={busy}>{busy ? "Creating profile…" : "Create local profile"}</button></footer>
        </form>
      </section>
    </div>
  );
}

function Field({ label, value, placeholder, type = "text", prefix, required = false, onChange }: {
  label: string;
  value: string;
  placeholder: string;
  type?: string;
  prefix?: string;
  required?: boolean;
  onChange: (value: string) => void;
}) {
  return <label className="field"><span>{label}</span><div className={prefix ? "prefixed-input" : ""}>{prefix && <em>{prefix}</em>}<input type={type} value={value} placeholder={placeholder} required={required} onChange={(event) => onChange(event.target.value)} /></div></label>;
}

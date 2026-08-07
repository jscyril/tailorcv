export type ApplicationStatus = "draft" | "submitted" | "archived";

const STATUS_OPTIONS: Array<{ value: ApplicationStatus; label: string }> = [
  { value: "draft", label: "Draft" },
  { value: "submitted", label: "Submitted" },
  { value: "archived", label: "Archived" },
];

export function ApplicationStatusControl({ status, busy = false, onChange }: {
  status: ApplicationStatus;
  busy?: boolean;
  onChange: (status: ApplicationStatus) => void;
}) {
  return (
    <div className="application-status-control" aria-label="Application status">
      <span>Status</span>
      <div role="group" aria-label="Set application status">
        {STATUS_OPTIONS.map((option) => (
          <button
            key={option.value}
            type="button"
            className={status === option.value ? "active" : ""}
            aria-pressed={status === option.value}
            disabled={busy || status === option.value}
            onClick={() => onChange(option.value)}
          >
            {option.label}
          </button>
        ))}
      </div>
    </div>
  );
}

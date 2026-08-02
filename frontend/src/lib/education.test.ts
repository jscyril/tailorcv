import { describe, expect, it } from "vitest";
import { educationToDraft, newEducationDraft, toEducationInput } from "./education";

describe("education drafts", () => {
  it("creates distinct local keys", () => {
    expect(newEducationDraft().key).not.toBe(newEducationDraft().key);
  });

  it("uses persisted IDs as keys and clears the end date for current study", () => {
    const draft = educationToDraft({
      id: "education-1",
      institution: "Example Institute",
      degree: "Master of Science",
      fieldOfStudy: "Computer Science",
      location: "Bengaluru",
      startDate: "2024-08",
      endDate: "2026-05",
      current: true,
      details: "Research assistant.",
    });
    expect(draft.key).toBe("education-1");
    expect(toEducationInput(draft).endDate).toBe("");
  });
});

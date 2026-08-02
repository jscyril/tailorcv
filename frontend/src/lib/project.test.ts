import { describe, expect, it } from "vitest";
import { newProjectDraft, toProjectInput } from "./project";

describe("toProjectInput", () => {
  it("normalizes skills and clears the end date for ongoing projects", () => {
    const draft = newProjectDraft();
    draft.skillsText = " Go, TypeScript, go ";
    draft.ongoing = true;
    draft.endDate = "2025-02";

    const input = toProjectInput(draft);
    expect(input.skills).toEqual(["Go", "TypeScript"]);
    expect(input.endDate).toBe("");
  });
});

import { describe, expect, it } from "vitest";
import { newProjectDraft, reconcileSelectedProjectKeys, removeSelectedProjectKey, toProjectInput } from "./project";

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

  it("removes deleted and ineligible projects from resume selection", () => {
    expect(removeSelectedProjectKey(["project-1", "project-2"], "project-1")).toEqual(["project-2"]);
    expect(reconcileSelectedProjectKeys(["new-project-1", "project-2"], "new-project-1", "saved-project", false)).toEqual(["project-2"]);
    expect(reconcileSelectedProjectKeys(["new-project-1"], "new-project-1", "saved-project", true)).toEqual(["saved-project"]);
  });
});

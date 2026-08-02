import { describe, expect, it } from "vitest";
import { filterProjects, isProjectSelectable, newProjectDraft, reconcileSelectedProjectKeys, removeSelectedProjectKey, toProjectInput, toggleProjectLanguage } from "./project";

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

  it("filters projects across names, roles, descriptions, and skills", () => {
    const compiler = { ...newProjectDraft(), name: "Atlas Compiler", role: "Maintainer", description: "Incremental builds", skills: ["Rust", "LLVM"], skillsText: "Rust, LLVM" };
    const dashboard = { ...newProjectDraft(), name: "Metrics Dashboard", role: "Frontend Engineer", description: "Operational analytics", skills: ["TypeScript", "React"], skillsText: "TypeScript, React" };
    const projects = [compiler, dashboard];

    expect(filterProjects(projects, "atlas")).toEqual([compiler]);
    expect(filterProjects(projects, "FRONTEND react")).toEqual([dashboard]);
    expect(filterProjects(projects, "incremental rust")).toEqual([compiler]);
    expect(filterProjects(projects, "   ")).toBe(projects);
    expect(filterProjects(projects, "python")).toEqual([]);
  });

  it("toggles a detected language without losing unrelated skills", () => {
    expect(toggleProjectLanguage("React, Go", "Rust")).toBe("React, Go, Rust");
    expect(toggleProjectLanguage("React, Go, Rust", "go")).toBe("React, Rust");
  });

  it("requires GitHub projects to be reviewed and eligible before selection", () => {
    expect(isProjectSelectable({ provenance: "github", verification: "unverified", resumeEligible: true })).toBe(false);
    expect(isProjectSelectable({ provenance: "github", verification: "verified", resumeEligible: true })).toBe(true);
    expect(isProjectSelectable({ provenance: "manual", verification: "unverified", resumeEligible: true })).toBe(true);
  });
});

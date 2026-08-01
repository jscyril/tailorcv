import { describe, expect, it } from "vitest";
import { emptyProfile, parseSkills, profileCompletion } from "./profile";

describe("parseSkills", () => {
  it("trims and removes case-insensitive duplicates", () => {
    expect(parseSkills(" Go, TypeScript, go,  PostgreSQL ")).toEqual([
      "Go",
      "TypeScript",
      "PostgreSQL",
    ]);
  });
});

describe("profileCompletion", () => {
  it("returns zero for a blank profile", () => {
    expect(profileCompletion(emptyProfile)).toBe(0);
  });
});

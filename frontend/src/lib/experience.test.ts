import { describe, expect, it } from "vitest";
import {
  moveItem,
  newExperienceDraft,
  toExperienceInput,
} from "./experience";

describe("moveItem", () => {
  it("moves evidence without mutating the original order", () => {
    const original = ["first", "second", "third"];
    expect(moveItem(original, 2, 0)).toEqual(["third", "first", "second"]);
    expect(original).toEqual(["first", "second", "third"]);
  });
});

describe("toExperienceInput", () => {
  it("clears the end date for a current role", () => {
    const draft = newExperienceDraft();
    draft.current = true;
    draft.endDate = "2025-01";
    expect(toExperienceInput(draft).endDate).toBe("");
  });
});

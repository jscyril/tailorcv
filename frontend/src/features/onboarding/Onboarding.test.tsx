// @vitest-environment jsdom

import { cleanup, render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach, describe, expect, it, vi } from "vitest";
import { Onboarding } from "./Onboarding";

afterEach(cleanup);

const profile = { name: "", email: "", headline: "", location: "", githubUsername: "", linkedInUrl: "" };

describe("Onboarding", () => {
  it("announces itself as a modal, focuses the first field, and contains keyboard focus", async () => {
    const user = userEvent.setup();
    render(<Onboarding profile={profile} skillsText="" busy={false} onChange={vi.fn()} onSkillsChange={vi.fn()} onSubmit={vi.fn()} onSkip={vi.fn()} />);

    expect(screen.getByRole("dialog", { name: "Start with facts you control." }).getAttribute("aria-modal")).toBe("true");
    const firstField = screen.getByLabelText("Full name");
    expect(document.activeElement).toBe(firstField);
    await user.tab({ shift: true });
    expect(document.activeElement).toBe(screen.getByRole("button", { name: "Create local profile" }));
    await user.tab();
    expect(document.activeElement).toBe(firstField);
  });

  it("supports keyboard activation of the explicit skip action", async () => {
    const user = userEvent.setup();
    const onSkip = vi.fn();
    render(<Onboarding profile={profile} skillsText="" busy={false} onChange={vi.fn()} onSkillsChange={vi.fn()} onSubmit={vi.fn()} onSkip={onSkip} />);
    screen.getByRole("button", { name: "Explore first" }).focus();
    await user.keyboard("{Enter}");
    expect(onSkip).toHaveBeenCalledOnce();
  });
});

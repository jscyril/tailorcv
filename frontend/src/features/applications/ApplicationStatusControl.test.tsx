// @vitest-environment jsdom

import { cleanup, render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach, describe, expect, it, vi } from "vitest";
import { ApplicationStatusControl } from "./ApplicationStatusControl";

afterEach(cleanup);

describe("ApplicationStatusControl", () => {
  it("shows the current lifecycle state and requests a transition", async () => {
    const onChange = vi.fn();
    render(<ApplicationStatusControl status="draft" onChange={onChange} />);

    expect(screen.getByRole("button", { name: "Draft" }).getAttribute("aria-pressed")).toBe("true");
    await userEvent.click(screen.getByRole("button", { name: "Submitted" }));

    expect(onChange).toHaveBeenCalledWith("submitted");
  });

  it("blocks transitions while a status update is in progress", async () => {
    const onChange = vi.fn();
    render(<ApplicationStatusControl status="submitted" busy onChange={onChange} />);

    await userEvent.click(screen.getByRole("button", { name: "Archived" }));
    expect(onChange).not.toHaveBeenCalled();
  });
});

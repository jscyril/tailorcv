// @vitest-environment jsdom

import { cleanup, render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach, describe, expect, it, vi } from "vitest";
import { PDFPreview, joinPDFText } from "./ResumePreview";

const pdf = vi.hoisted(() => ({ numPages: 1 }));
const runtime = vi.hoisted(() => ({ ClipboardSetText: vi.fn() }));
vi.mock("../../../wailsjs/runtime/runtime", () => runtime);
const pageText = [{ str: "ADA LOVELACE", hasEOL: true }, { str: "Go", hasEOL: false }, { str: " · ", hasEOL: false }, { str: "Rust", hasEOL: true }];

vi.mock("pdfjs-dist/build/pdf.worker.min.mjs?url", () => ({ default: "worker.js" }));
vi.mock("pdfjs-dist", () => ({
  GlobalWorkerOptions: {},
  getDocument: () => ({
    promise: Promise.resolve({ numPages: pdf.numPages, getPage: async () => ({ getViewport: () => ({ width: 10, height: 10 }), render: () => ({ promise: Promise.resolve(), cancel() {} }), getTextContent: async () => ({ items: pageText }) }) }),
    destroy: async () => {},
  }),
}));

afterEach(cleanup);
// jsdom has no canvas; drawing fails quietly and the page count still shows.
HTMLCanvasElement.prototype.getContext = () => null;

describe("PDFPreview", () => {
  it("warns when the compiled resume runs past one page", async () => {
    pdf.numPages = 2;
    render(<PDFPreview data={btoa("%PDF")} />);
    expect(await screen.findByText(/runs to 2 pages/)).toBeTruthy();
  });

  it("shows no warning for a one-page resume", async () => {
    pdf.numPages = 1;
    render(<PDFPreview data={btoa("%PDF")} />);
    expect(await screen.findByText("1 page")).toBeTruthy();
    expect(screen.queryByText(/runs to/)).toBeNull();
  });

  it("copies the compiled resume as plain text", async () => {
    pdf.numPages = 2;
    runtime.ClipboardSetText.mockResolvedValue(true);
    render(<PDFPreview data={btoa("%PDF")} />);
    await userEvent.click(screen.getByRole("button", { name: "Copy text" }));
    expect(await screen.findByRole("button", { name: "Copied" })).toBeTruthy();
    expect(runtime.ClipboardSetText).toHaveBeenCalledWith("ADA LOVELACE\nGo · Rust\n\nADA LOVELACE\nGo · Rust");
  });
});

describe("joinPDFText", () => {
  it("keeps line breaks, trims trailing spaces, and separates pages", () => {
    expect(joinPDFText([[{ str: "A  ", hasEOL: true }, { str: "B" }], [], [{ str: "C", hasEOL: true }, { str: "", hasEOL: true }, { str: "", hasEOL: true }, { str: "D" }]])).toBe("A\nB\n\nC\n\nD");
  });
});

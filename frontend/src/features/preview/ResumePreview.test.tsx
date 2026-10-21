// @vitest-environment jsdom

import { cleanup, render, screen } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";
import { PDFPreview } from "./ResumePreview";

const pdf = vi.hoisted(() => ({ numPages: 1 }));

vi.mock("pdfjs-dist/build/pdf.worker.min.mjs?url", () => ({ default: "worker.js" }));
vi.mock("pdfjs-dist", () => ({
  GlobalWorkerOptions: {},
  getDocument: () => ({
    promise: Promise.resolve({ numPages: pdf.numPages, getPage: async () => ({ getViewport: () => ({ width: 10, height: 10 }), render: () => ({ promise: Promise.resolve(), cancel() {} }) }) }),
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
});

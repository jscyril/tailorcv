import { domain } from "../../wailsjs/go/models";
import { type EvidenceMatch } from "../features/ai/AIWorkspace";
import { type ApplicationStatus } from "../features/applications/ApplicationStatusControl";

export type View = "overview" | "profile" | "experience" | "projects" | "education" | "credentials" | "skills" | "templates" | "latex" | "job" | "ai" | "data";

export type ResumeVersion = {
  id: string;
  applicationId: string;
  versionNumber: number;
  jobDescriptionSnapshot: string;
  selectedFactIds: string[];
  latexSource: string;
  templateId: string;
  rankingExplanations: EvidenceMatch[];
  contentHash: string;
  compileSuccess: boolean;
  compileEngine: string;
  compileDurationMs: number;
  compileDiagnostics: domain.CompileDiagnostic[];
  compiledAt: string;
  pdfAvailable: boolean;
  createdAt: string;
};

export type Application = {
  id: string;
  jobId: string;
  name: string;
  status: ApplicationStatus;
  selectedFactIds: string[];
  versions: ResumeVersion[];
  createdAt: string;
  updatedAt: string;
};

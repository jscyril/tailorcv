import { EvidenceBullet, Provenance, VerificationState } from "./experience";
import { parseSkills } from "./profile";

export type Project = {
  id: string;
  name: string;
  role: string;
  description: string;
  url: string;
  repositoryUrl: string;
  startDate: string;
  endDate: string;
  ongoing: boolean;
  provenance: Provenance;
  verification: VerificationState;
  resumeEligible: boolean;
  position?: number;
  skills: string[];
  bullets: EvidenceBullet[];
  createdAt?: string;
  updatedAt?: string;
};

export type ProjectDraft = Project & { key: string; skillsText: string };

let draftSequence = 0;

export function newProjectDraft(): ProjectDraft {
  draftSequence += 1;
  return {
    key: `new-project-${draftSequence}`,
    id: "",
    name: "",
    role: "",
    description: "",
    url: "",
    repositoryUrl: "",
    startDate: "",
    endDate: "",
    ongoing: false,
    provenance: "manual",
    verification: "unverified",
    resumeEligible: true,
    skills: [],
    skillsText: "",
    bullets: [],
  };
}

export function projectToDraft(project: Project): ProjectDraft {
  const skills = project.skills ?? [];
  return {
    ...project,
    key: project.id,
    skills,
    skillsText: skills.join(", "),
    bullets: project.bullets ?? [],
  };
}

export function toProjectInput(draft: ProjectDraft) {
  return {
    id: draft.id,
    name: draft.name,
    role: draft.role,
    description: draft.description,
    url: draft.url,
    repositoryUrl: draft.repositoryUrl,
    startDate: draft.startDate,
    endDate: draft.ongoing ? "" : draft.endDate,
    ongoing: draft.ongoing,
    provenance: draft.provenance,
    verification: draft.verification,
    resumeEligible: draft.resumeEligible,
    skills: parseSkills(draft.skillsText),
    bullets: draft.bullets.map(({ id, text, provenance, sourceUrl, verification }) => ({
      id,
      text,
      provenance,
      sourceUrl,
      verification,
    })),
  };
}

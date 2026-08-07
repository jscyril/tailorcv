import { EvidenceBullet, Provenance, VerificationState } from "./experience";
import { parseSkills } from "./profile";

export type Project = {
  id: string;
  name: string;
  role: string;
  description: string;
  url: string;
  repositoryUrl: string;
  repositoryId: number;
  repositoryReadme: string;
  repositoryVisibility: string;
  repositoryUpdatedAt: string;
  startDate: string;
  endDate: string;
  ongoing: boolean;
  provenance: Provenance;
  verification: VerificationState;
  resumeEligible: boolean;
  position?: number;
  skills: string[];
  detectedLanguages: RepositoryLanguage[];
  bullets: EvidenceBullet[];
  createdAt?: string;
  updatedAt?: string;
};

export type RepositoryLanguage = {
  name: string;
  bytes: number;
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
    repositoryId: 0,
    repositoryReadme: "",
    repositoryVisibility: "",
    repositoryUpdatedAt: "",
    startDate: "",
    endDate: "",
    ongoing: false,
    provenance: "manual",
    verification: "unverified",
    resumeEligible: true,
    skills: [],
    skillsText: "",
    detectedLanguages: [],
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
    detectedLanguages: project.detectedLanguages ?? [],
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
    repositoryId: draft.repositoryId,
    repositoryReadme: draft.repositoryReadme,
    repositoryVisibility: draft.repositoryVisibility,
    repositoryUpdatedAt: draft.repositoryUpdatedAt,
    startDate: draft.startDate,
    endDate: draft.ongoing ? "" : draft.endDate,
    ongoing: draft.ongoing,
    provenance: draft.provenance,
    verification: draft.verification,
    resumeEligible: draft.resumeEligible,
    skills: parseSkills(draft.skillsText),
    detectedLanguages: draft.detectedLanguages,
    bullets: draft.bullets.map(({ id, text, provenance, sourceUrl, verification, importance }) => ({
      id,
      text,
      provenance,
      sourceUrl,
      verification,
      importance,
    })),
  };
}

export function toggleProjectLanguage(skillsText: string, language: string): string {
  const skills = parseSkills(skillsText);
  const matchingIndex = skills.findIndex((skill) => skill.toLocaleLowerCase() === language.toLocaleLowerCase());
  if (matchingIndex >= 0) {
    return skills.filter((_, index) => index !== matchingIndex).join(", ");
  }
  return [...skills, language].join(", ");
}

export function reconcileSelectedProjectKeys(selectedKeys: string[], previousKey: string, savedKey: string, resumeEligible: boolean): string[] {
  const updated = selectedKeys.map((key) => key === previousKey ? savedKey : key);
  if (!resumeEligible) {
    return updated.filter((key) => key !== savedKey);
  }
  return [...new Set(updated)];
}

export function removeSelectedProjectKey(selectedKeys: string[], key: string): string[] {
  return selectedKeys.filter((selectedKey) => selectedKey !== key);
}

export function isProjectSelectable(project: Pick<Project, "provenance" | "verification" | "resumeEligible">): boolean {
  return project.resumeEligible && (project.provenance !== "github" || project.verification === "verified");
}

export function filterProjects(projects: ProjectDraft[], query: string): ProjectDraft[] {
  const terms = query.trim().toLocaleLowerCase().split(/\s+/).filter(Boolean);
  if (terms.length === 0) {
    return projects;
  }

  return projects.filter((project) => {
    const searchableText = [
      project.name,
      project.role,
      project.description,
      project.skillsText,
      ...project.skills,
    ].join(" ").toLocaleLowerCase();

    return terms.every((term) => searchableText.includes(term));
  });
}

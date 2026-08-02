export type Provenance = "manual" | "github" | "imported";
export type VerificationState = "unverified" | "verified";

export type EvidenceBullet = {
  id: string;
  text: string;
  provenance: Provenance;
  sourceUrl: string;
  verification: VerificationState;
  position?: number;
  createdAt?: string;
  updatedAt?: string;
};

export type Experience = {
  id: string;
  company: string;
  title: string;
  location: string;
  startDate: string;
  endDate: string;
  current: boolean;
  position?: number;
  bullets: EvidenceBullet[];
  createdAt?: string;
  updatedAt?: string;
};

export type ExperienceDraft = Experience & { key: string };

let draftSequence = 0;

export function newEvidenceBullet(): EvidenceBullet {
  return {
    id: "",
    text: "",
    provenance: "manual",
    sourceUrl: "",
    verification: "unverified",
  };
}

export function newExperienceDraft(): ExperienceDraft {
  draftSequence += 1;
  return {
    key: `new-experience-${draftSequence}`,
    id: "",
    company: "",
    title: "",
    location: "",
    startDate: "",
    endDate: "",
    current: false,
    bullets: [],
  };
}

export function experienceToDraft(experience: Experience): ExperienceDraft {
  return {
    ...experience,
    key: experience.id,
    bullets: experience.bullets ?? [],
  };
}

export function moveItem<T>(items: T[], from: number, to: number): T[] {
  if (from < 0 || from >= items.length || to < 0 || to >= items.length || from === to) {
    return items;
  }
  const result = [...items];
  const [item] = result.splice(from, 1);
  result.splice(to, 0, item);
  return result;
}

export function toExperienceInput(draft: ExperienceDraft) {
  return {
    id: draft.id,
    company: draft.company,
    title: draft.title,
    location: draft.location,
    startDate: draft.startDate,
    endDate: draft.current ? "" : draft.endDate,
    current: draft.current,
    bullets: draft.bullets.map(({ id, text, provenance, sourceUrl, verification }) => ({
      id,
      text,
      provenance,
      sourceUrl,
      verification,
    })),
  };
}

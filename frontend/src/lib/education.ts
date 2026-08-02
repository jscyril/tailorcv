export type Education = {
  id: string;
  institution: string;
  degree: string;
  fieldOfStudy: string;
  location: string;
  startDate: string;
  endDate: string;
  current: boolean;
  details: string;
  position?: number;
  createdAt?: string;
  updatedAt?: string;
};

export type EducationDraft = Education & { key: string };

let draftSequence = 0;

export function newEducationDraft(): EducationDraft {
  draftSequence += 1;
  return {
    key: `new-education-${draftSequence}`,
    id: "",
    institution: "",
    degree: "",
    fieldOfStudy: "",
    location: "",
    startDate: "",
    endDate: "",
    current: false,
    details: "",
  };
}

export function educationToDraft(education: Education): EducationDraft {
  return { ...education, key: education.id };
}

export function toEducationInput(draft: EducationDraft) {
  return {
    id: draft.id,
    institution: draft.institution,
    degree: draft.degree,
    fieldOfStudy: draft.fieldOfStudy,
    location: draft.location,
    startDate: draft.startDate,
    endDate: draft.current ? "" : draft.endDate,
    current: draft.current,
    details: draft.details,
  };
}

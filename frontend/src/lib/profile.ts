export type Profile = {
  name: string;
  headline: string;
  email: string;
  phone: string;
  location: string;
  website: string;
  githubUsername: string;
  linkedInUrl: string;
  summary: string;
  skills: string[];
  updatedAt?: string;
};

export const emptyProfile: Profile = {
  name: "",
  headline: "",
  email: "",
  phone: "",
  location: "",
  website: "",
  githubUsername: "",
  linkedInUrl: "",
  summary: "",
  skills: [],
};

export function parseSkills(value: string): string[] {
  const seen = new Set<string>();
  return value
    .split(",")
    .map((skill) => skill.trim().replace(/\s+/g, " "))
    .filter((skill) => {
      const key = skill.toLocaleLowerCase();
      if (!skill || seen.has(key)) return false;
      seen.add(key);
      return true;
    });
}

export function profileCompletion(profile: Profile): number {
  const values = [
    profile.name,
    profile.headline,
    profile.email,
    profile.location,
    profile.summary,
    profile.githubUsername,
    profile.skills.length ? "skills" : "",
  ];
  return Math.round((values.filter(Boolean).length / values.length) * 100);
}

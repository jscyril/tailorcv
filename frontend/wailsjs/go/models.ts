export namespace domain {
	export class CompileResult {
	    pdfBase64: string;
	    engine: string;
	    durationMs: number;
	    log: string;

	    static createFrom(source: any = {}) { return new CompileResult(source); }
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.pdfBase64 = source["pdfBase64"];
	        this.engine = source["engine"];
	        this.durationMs = source["durationMs"];
	        this.log = source["log"];
	    }
	}
	export class FileResult {
	    path: string;
	    cancelled: boolean;

	    static createFrom(source: any = {}) { return new FileResult(source); }
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.path = source["path"];
	        this.cancelled = source["cancelled"];
	    }
	}
	export class ResumeTemplate {
	    id: string;
	    name: string;
	    description: string;
	    source: string;
	    builtIn: boolean;
	    createdAt: string;
	    updatedAt: string;

	    static createFrom(source: any = {}) { return new ResumeTemplate(source); }
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.description = source["description"];
	        this.source = source["source"];
	        this.builtIn = source["builtIn"];
	        this.createdAt = source["createdAt"];
	        this.updatedAt = source["updatedAt"];
	    }
	}
	export class ResumeTemplateInput {
	    id: string;
	    name: string;
	    description: string;
	    source: string;

	    static createFrom(source: any = {}) { return new ResumeTemplateInput(source); }
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.description = source["description"];
	        this.source = source["source"];
	    }
	}
	export class BackupResult {
	    path: string;
	    cancelled: boolean;
	    experienceCount: number;
	    projectCount: number;
	    educationCount: number;

	    static createFrom(source: any = {}) {
	        return new BackupResult(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.path = source["path"];
	        this.cancelled = source["cancelled"];
	        this.experienceCount = source["experienceCount"];
	        this.projectCount = source["projectCount"];
	        this.educationCount = source["educationCount"];
	    }
	}
	export class Education {
	    id: string;
	    institution: string;
	    degree: string;
	    fieldOfStudy: string;
	    location: string;
	    startDate: string;
	    endDate: string;
	    current: boolean;
	    details: string;
	    position: number;
	    createdAt: string;
	    updatedAt: string;

	    static createFrom(source: any = {}) {
	        return new Education(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.institution = source["institution"];
	        this.degree = source["degree"];
	        this.fieldOfStudy = source["fieldOfStudy"];
	        this.location = source["location"];
	        this.startDate = source["startDate"];
	        this.endDate = source["endDate"];
	        this.current = source["current"];
	        this.details = source["details"];
	        this.position = source["position"];
	        this.createdAt = source["createdAt"];
	        this.updatedAt = source["updatedAt"];
	    }
	}
	export class EducationInput {
	    id: string;
	    institution: string;
	    degree: string;
	    fieldOfStudy: string;
	    location: string;
	    startDate: string;
	    endDate: string;
	    current: boolean;
	    details: string;

	    static createFrom(source: any = {}) {
	        return new EducationInput(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.institution = source["institution"];
	        this.degree = source["degree"];
	        this.fieldOfStudy = source["fieldOfStudy"];
	        this.location = source["location"];
	        this.startDate = source["startDate"];
	        this.endDate = source["endDate"];
	        this.current = source["current"];
	        this.details = source["details"];
	    }
	}
	export class GitHubImportResult {
	    fetched: number;
	    imported: number;
	    updated: number;
	    skipped: number;

	    static createFrom(source: any = {}) {
	        return new GitHubImportResult(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.fetched = source["fetched"];
	        this.imported = source["imported"];
	        this.updated = source["updated"];
	        this.skipped = source["skipped"];
	    }
	}

	export class EvidenceBullet {
	    id: string;
	    text: string;
	    provenance: string;
	    sourceUrl: string;
	    verification: string;
	    position: number;
	    createdAt: string;
	    updatedAt: string;

	    static createFrom(source: any = {}) {
	        return new EvidenceBullet(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.text = source["text"];
	        this.provenance = source["provenance"];
	        this.sourceUrl = source["sourceUrl"];
	        this.verification = source["verification"];
	        this.position = source["position"];
	        this.createdAt = source["createdAt"];
	        this.updatedAt = source["updatedAt"];
	    }
	}
	export class EvidenceBulletInput {
	    id: string;
	    text: string;
	    provenance: string;
	    sourceUrl: string;
	    verification: string;

	    static createFrom(source: any = {}) {
	        return new EvidenceBulletInput(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.text = source["text"];
	        this.provenance = source["provenance"];
	        this.sourceUrl = source["sourceUrl"];
	        this.verification = source["verification"];
	    }
	}
	export class Experience {
	    id: string;
	    company: string;
	    title: string;
	    location: string;
	    startDate: string;
	    endDate: string;
	    current: boolean;
	    position: number;
	    bullets: EvidenceBullet[];
	    createdAt: string;
	    updatedAt: string;

	    static createFrom(source: any = {}) {
	        return new Experience(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.company = source["company"];
	        this.title = source["title"];
	        this.location = source["location"];
	        this.startDate = source["startDate"];
	        this.endDate = source["endDate"];
	        this.current = source["current"];
	        this.position = source["position"];
	        this.bullets = this.convertValues(source["bullets"], EvidenceBullet);
	        this.createdAt = source["createdAt"];
	        this.updatedAt = source["updatedAt"];
	    }

	    convertValues(a: any, classs: any, asMap: boolean = false): any {
	        if (!a) return a;
	        if (a.slice && a.map) return (a as any[]).map(elem => this.convertValues(elem, classs));
	        if ("object" === typeof a) {
	            if (asMap) {
	                for (const key of Object.keys(a)) a[key] = new classs(a[key]);
	                return a;
	            }
	            return new classs(a);
	        }
	        return a;
	    }
	}
	export class ExperienceInput {
	    id: string;
	    company: string;
	    title: string;
	    location: string;
	    startDate: string;
	    endDate: string;
	    current: boolean;
	    bullets: EvidenceBulletInput[];

	    static createFrom(source: any = {}) {
	        return new ExperienceInput(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.company = source["company"];
	        this.title = source["title"];
	        this.location = source["location"];
	        this.startDate = source["startDate"];
	        this.endDate = source["endDate"];
	        this.current = source["current"];
	        this.bullets = this.convertValues(source["bullets"], EvidenceBulletInput);
	    }

	    convertValues(a: any, classs: any, asMap: boolean = false): any {
	        if (!a) return a;
	        if (a.slice && a.map) return (a as any[]).map(elem => this.convertValues(elem, classs));
	        if ("object" === typeof a) {
	            if (asMap) {
	                for (const key of Object.keys(a)) a[key] = new classs(a[key]);
	                return a;
	            }
	            return new classs(a);
	        }
	        return a;
	    }
	}
	export class Project {
	    id: string;
	    name: string;
	    role: string;
	    description: string;
	    url: string;
	    repositoryUrl: string;
	    startDate: string;
	    endDate: string;
	    ongoing: boolean;
	    provenance: string;
	    verification: string;
	    resumeEligible: boolean;
	    position: number;
	    skills: string[];
	    bullets: EvidenceBullet[];
	    createdAt: string;
	    updatedAt: string;

	    static createFrom(source: any = {}) {
	        return new Project(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.role = source["role"];
	        this.description = source["description"];
	        this.url = source["url"];
	        this.repositoryUrl = source["repositoryUrl"];
	        this.startDate = source["startDate"];
	        this.endDate = source["endDate"];
	        this.ongoing = source["ongoing"];
	        this.provenance = source["provenance"];
	        this.verification = source["verification"];
	        this.resumeEligible = source["resumeEligible"];
	        this.position = source["position"];
	        this.skills = source["skills"];
	        this.bullets = this.convertValues(source["bullets"], EvidenceBullet);
	        this.createdAt = source["createdAt"];
	        this.updatedAt = source["updatedAt"];
	    }

	    convertValues(a: any, classs: any, asMap: boolean = false): any {
	        if (!a) return a;
	        if (a.slice && a.map) return (a as any[]).map(elem => this.convertValues(elem, classs));
	        if ("object" === typeof a) {
	            if (asMap) {
	                for (const key of Object.keys(a)) a[key] = new classs(a[key]);
	                return a;
	            }
	            return new classs(a);
	        }
	        return a;
	    }
	}
	export class ProjectInput {
	    id: string;
	    name: string;
	    role: string;
	    description: string;
	    url: string;
	    repositoryUrl: string;
	    startDate: string;
	    endDate: string;
	    ongoing: boolean;
	    provenance: string;
	    verification: string;
	    resumeEligible: boolean;
	    skills: string[];
	    bullets: EvidenceBulletInput[];

	    static createFrom(source: any = {}) {
	        return new ProjectInput(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.role = source["role"];
	        this.description = source["description"];
	        this.url = source["url"];
	        this.repositoryUrl = source["repositoryUrl"];
	        this.startDate = source["startDate"];
	        this.endDate = source["endDate"];
	        this.ongoing = source["ongoing"];
	        this.provenance = source["provenance"];
	        this.verification = source["verification"];
	        this.resumeEligible = source["resumeEligible"];
	        this.skills = source["skills"];
	        this.bullets = this.convertValues(source["bullets"], EvidenceBulletInput);
	    }

	    convertValues(a: any, classs: any, asMap: boolean = false): any {
	        if (!a) return a;
	        if (a.slice && a.map) return (a as any[]).map(elem => this.convertValues(elem, classs));
	        if ("object" === typeof a) {
	            if (asMap) {
	                for (const key of Object.keys(a)) a[key] = new classs(a[key]);
	                return a;
	            }
	            return new classs(a);
	        }
	        return a;
	    }
	}

	export class JobAnalysis {
	    score: number;
	    matchedSkills: string[];
	    unmentionedSkills: string[];
	    explanation: string;

	    static createFrom(source: any = {}) {
	        return new JobAnalysis(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.score = source["score"];
	        this.matchedSkills = source["matchedSkills"];
	        this.unmentionedSkills = source["unmentionedSkills"];
	        this.explanation = source["explanation"];
	    }
	}
	export class JobAnalysisInput {
	    description: string;

	    static createFrom(source: any = {}) {
	        return new JobAnalysisInput(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.description = source["description"];
	    }
	}
	export class Profile {
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
	    updatedAt: string;

	    static createFrom(source: any = {}) {
	        return new Profile(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.headline = source["headline"];
	        this.email = source["email"];
	        this.phone = source["phone"];
	        this.location = source["location"];
	        this.website = source["website"];
	        this.githubUsername = source["githubUsername"];
	        this.linkedInUrl = source["linkedInUrl"];
	        this.summary = source["summary"];
	        this.skills = source["skills"];
	        this.updatedAt = source["updatedAt"];
	    }
	}
	export class ProfileInput {
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

	    static createFrom(source: any = {}) {
	        return new ProfileInput(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.headline = source["headline"];
	        this.email = source["email"];
	        this.phone = source["phone"];
	        this.location = source["location"];
	        this.website = source["website"];
	        this.githubUsername = source["githubUsername"];
	        this.linkedInUrl = source["linkedInUrl"];
	        this.summary = source["summary"];
	        this.skills = source["skills"];
	    }
	}

}

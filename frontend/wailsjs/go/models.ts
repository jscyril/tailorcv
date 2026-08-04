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
	    jobCount: number;
	    applicationCount: number;
	    resumeVersionCount: number;

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
	        this.jobCount = source["jobCount"];
	        this.applicationCount = source["applicationCount"];
	        this.resumeVersionCount = source["resumeVersionCount"];
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
	    languageFallbacks: number;

	    static createFrom(source: any = {}) {
	        return new GitHubImportResult(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.fetched = source["fetched"];
	        this.imported = source["imported"];
	        this.updated = source["updated"];
	        this.skipped = source["skipped"];
	        this.languageFallbacks = source["languageFallbacks"];
	    }
	}
	export class RepositoryLanguage {
	    name: string;
	    bytes: number;

	    static createFrom(source: any = {}) { return new RepositoryLanguage(source); }
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.bytes = source["bytes"];
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
	    detectedLanguages: RepositoryLanguage[];
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
	        this.detectedLanguages = this.convertValues(source["detectedLanguages"], RepositoryLanguage);
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
	    detectedLanguages: RepositoryLanguage[];
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
	        this.detectedLanguages = this.convertValues(source["detectedLanguages"], RepositoryLanguage);
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

	export class Job {
	    id: string;
	    company: string;
	    role: string;
	    description: string;
	    createdAt: string;
	    updatedAt: string;

	    static createFrom(source: any = {}) { return new Job(source); }
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.company = source["company"];
	        this.role = source["role"];
	        this.description = source["description"];
	        this.createdAt = source["createdAt"];
	        this.updatedAt = source["updatedAt"];
	    }
	}
	export class EvidenceMatch {
	    factId: string;
	    sourceId: string;
	    sourceType: string;
	    sourceLabel: string;
	    text: string;
	    score: number;
	    matchedSkills: string[];
	    reasons: string[];
	    verified: boolean;
	    selectable: boolean;

	    static createFrom(source: any = {}) { return new EvidenceMatch(source); }
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.factId = source["factId"];
	        this.sourceId = source["sourceId"];
	        this.sourceType = source["sourceType"];
	        this.sourceLabel = source["sourceLabel"];
	        this.text = source["text"];
	        this.score = source["score"];
	        this.matchedSkills = source["matchedSkills"];
	        this.reasons = source["reasons"];
	        this.verified = source["verified"];
	        this.selectable = source["selectable"];
	    }
	}
	export class ResumeVersion {
	    id: string;
	    applicationId: string;
	    versionNumber: number;
	    jobDescriptionSnapshot: string;
	    selectedFactIds: string[];
	    latexSource: string;
	    templateId: string;
	    createdAt: string;

	    static createFrom(source: any = {}) { return new ResumeVersion(source); }
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.applicationId = source["applicationId"];
	        this.versionNumber = source["versionNumber"];
	        this.jobDescriptionSnapshot = source["jobDescriptionSnapshot"];
	        this.selectedFactIds = source["selectedFactIds"];
	        this.latexSource = source["latexSource"];
	        this.templateId = source["templateId"];
	        this.createdAt = source["createdAt"];
	    }
	}
	export class Application {
	    id: string;
	    jobId: string;
	    status: string;
	    selectedFactIds: string[];
	    versions: ResumeVersion[];
	    createdAt: string;
	    updatedAt: string;

	    static createFrom(source: any = {}) { return new Application(source); }
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.jobId = source["jobId"];
	        this.status = source["status"];
	        this.selectedFactIds = source["selectedFactIds"];
	        this.versions = this.convertValues(source["versions"], ResumeVersion);
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
	export class CreateResumeVersionInput {
	    jobId: string;
	    selectedFactIds: string[];
	    templateId: string;

	    static createFrom(source: any = {}) { return new CreateResumeVersionInput(source); }
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.jobId = source["jobId"];
	        this.selectedFactIds = source["selectedFactIds"];
	        this.templateId = source["templateId"];
	    }
	}
	export class ApplicationResumeResult {
	    application: Application;
	    version: ResumeVersion;

	    static createFrom(source: any = {}) { return new ApplicationResumeResult(source); }
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.application = this.convertValues(source["application"], Application);
	        this.version = this.convertValues(source["version"], ResumeVersion);
	    }

	    convertValues(a: any, classs: any): any {
	        if (!a) return a;
	        if (a.slice && a.map) return (a as any[]).map(elem => this.convertValues(elem, classs));
	        if ("object" === typeof a) return new classs(a);
	        return a;
	    }
	}
	export class JobAnalysis {
	    job: Job;
	    score: number;
	    matchedSkills: string[];
	    unmentionedSkills: string[];
	    detectedSkills: string[];
	    requiredSkills: string[];
	    preferredSkills: string[];
	    responsibilities: string[];
	    keywords: string[];
	    searchTerms: string[];
	    rankedEvidence: EvidenceMatch[];
	    explanation: string;

	    static createFrom(source: any = {}) {
	        return new JobAnalysis(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.job = this.convertValues(source["job"], Job);
	        this.score = source["score"];
	        this.matchedSkills = source["matchedSkills"];
	        this.unmentionedSkills = source["unmentionedSkills"];
	        this.detectedSkills = source["detectedSkills"];
	        this.requiredSkills = source["requiredSkills"];
	        this.preferredSkills = source["preferredSkills"];
	        this.responsibilities = source["responsibilities"];
	        this.keywords = source["keywords"];
	        this.searchTerms = source["searchTerms"];
	        this.rankedEvidence = this.convertValues(source["rankedEvidence"], EvidenceMatch);
	        this.explanation = source["explanation"];
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
	export class JobAnalysisInput {
	    id: string;
	    company: string;
	    role: string;
	    description: string;

	    static createFrom(source: any = {}) {
	        return new JobAnalysisInput(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.company = source["company"];
	        this.role = source["role"];
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

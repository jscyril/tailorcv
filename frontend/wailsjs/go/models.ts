export namespace domain {

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

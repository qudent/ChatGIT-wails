export namespace models {
	
	export class Commit {
	    sha: string;
	    parents: string[];
	    author: string;
	    subject: string;
	    body: string;
	    refs: string[];
	    role: string;
	    type: string;
	    relatesTo: string;
	    // Go type: time
	    timestamp: any;
	
	    static createFrom(source: any = {}) {
	        return new Commit(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.sha = source["sha"];
	        this.parents = source["parents"];
	        this.author = source["author"];
	        this.subject = source["subject"];
	        this.body = source["body"];
	        this.refs = source["refs"];
	        this.role = source["role"];
	        this.type = source["type"];
	        this.relatesTo = source["relatesTo"];
	        this.timestamp = this.convertValues(source["timestamp"], null);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class CommitList {
	    commits: Commit[];
	    total: number;
	    hasMore: boolean;
	
	    static createFrom(source: any = {}) {
	        return new CommitList(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.commits = this.convertValues(source["commits"], Commit);
	        this.total = source["total"];
	        this.hasMore = source["hasMore"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class Hunk {
	    oldStart: number;
	    oldLines: number;
	    newStart: number;
	    newLines: number;
	    content: string;
	
	    static createFrom(source: any = {}) {
	        return new Hunk(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.oldStart = source["oldStart"];
	        this.oldLines = source["oldLines"];
	        this.newStart = source["newStart"];
	        this.newLines = source["newLines"];
	        this.content = source["content"];
	    }
	}
	export class Diff {
	    from: string;
	    to: string;
	    content: string;
	    hunks: Hunk[];
	    files: string[];
	
	    static createFrom(source: any = {}) {
	        return new Diff(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.from = source["from"];
	        this.to = source["to"];
	        this.content = source["content"];
	        this.hunks = this.convertValues(source["hunks"], Hunk);
	        this.files = source["files"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	export class PatchProposal {
	    id: string;
	    description: string;
	    diff: string;
	    files: string[];
	    approved: boolean;
	
	    static createFrom(source: any = {}) {
	        return new PatchProposal(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.description = source["description"];
	        this.diff = source["diff"];
	        this.files = source["files"];
	        this.approved = source["approved"];
	    }
	}
	export class RepoMetadata {
	    path: string;
	    name: string;
	    lastUsed: number;
	    favorite: boolean;
	    tags: string[];
	
	    static createFrom(source: any = {}) {
	        return new RepoMetadata(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.path = source["path"];
	        this.name = source["name"];
	        this.lastUsed = source["lastUsed"];
	        this.favorite = source["favorite"];
	        this.tags = source["tags"];
	    }
	}
	export class Settings {
	    provider: string;
	    model: string;
	    maxTokens: number;
	    temperature: number;
	
	    static createFrom(source: any = {}) {
	        return new Settings(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.provider = source["provider"];
	        this.model = source["model"];
	        this.maxTokens = source["maxTokens"];
	        this.temperature = source["temperature"];
	    }
	}
	export class Status {
	    branch: string;
	    head: string;
	    ahead: number;
	    behind: number;
	    staged: number;
	    unstaged: number;
	    untracked: number;
	    isClean: boolean;
	
	    static createFrom(source: any = {}) {
	        return new Status(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.branch = source["branch"];
	        this.head = source["head"];
	        this.ahead = source["ahead"];
	        this.behind = source["behind"];
	        this.staged = source["staged"];
	        this.unstaged = source["unstaged"];
	        this.untracked = source["untracked"];
	        this.isClean = source["isClean"];
	    }
	}
	export class UIState {
	    currentRepo: string;
	    openTabs: string[];
	    splitLayout: boolean;
	    theme: string;
	
	    static createFrom(source: any = {}) {
	        return new UIState(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.currentRepo = source["currentRepo"];
	        this.openTabs = source["openTabs"];
	        this.splitLayout = source["splitLayout"];
	        this.theme = source["theme"];
	    }
	}

}


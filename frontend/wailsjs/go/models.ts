export namespace domain {
	
	export class Video {
	    path: string;
	    name: string;
	    size: number;
	    duration: number;
	    audioBitrate: number;
	
	    static createFrom(source: any = {}) {
	        return new Video(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.path = source["path"];
	        this.name = source["name"];
	        this.size = source["size"];
	        this.duration = source["duration"];
	        this.audioBitrate = source["audioBitrate"];
	    }
	}

}


export namespace main {
	
	export class AudioResponse {
	    text: string;
	    audio: string;
	
	    static createFrom(source: any = {}) {
	        return new AudioResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.text = source["text"];
	        this.audio = source["audio"];
	    }
	}
	export class ScreenSize {
	    width: number;
	    height: number;
	
	    static createFrom(source: any = {}) {
	        return new ScreenSize(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.width = source["width"];
	        this.height = source["height"];
	    }
	}
	export class WindowPos {
	    x: number;
	    y: number;
	
	    static createFrom(source: any = {}) {
	        return new WindowPos(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.x = source["x"];
	        this.y = source["y"];
	    }
	}

}


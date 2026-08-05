export namespace audio {
	
	export class Songdata {
	    Filetype: string;
	    Title: string;
	    Album: string;
	    Artist: string;
	    Year: number;
	    Track: number;
	    TrackTotal: number;
	    Duration: number;
	
	    static createFrom(source: any = {}) {
	        return new Songdata(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Filetype = source["Filetype"];
	        this.Title = source["Title"];
	        this.Album = source["Album"];
	        this.Artist = source["Artist"];
	        this.Year = source["Year"];
	        this.Track = source["Track"];
	        this.TrackTotal = source["TrackTotal"];
	        this.Duration = source["Duration"];
	    }
	}

}


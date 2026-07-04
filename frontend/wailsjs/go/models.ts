export namespace main {
	
	export class KeyRecord {
	    id: string;
	    provider: string;
	    name: string;
	    value: string;
	    createdAt: string;
	
	    static createFrom(source: any = {}) {
	        return new KeyRecord(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.provider = source["provider"];
	        this.name = source["name"];
	        this.value = source["value"];
	        this.createdAt = source["createdAt"];
	    }
	}

}


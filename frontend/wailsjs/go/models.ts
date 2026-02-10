export namespace main {
	
	export class DownloadProgress {
	    id: string;
	    filename: string;
	    status: string;
	    statusText: string;
	    progress: number;
	    totalBytes: number;
	    downloadedBytes: number;
	    error?: string;
	    outputPath: string;
	
	    static createFrom(source: any = {}) {
	        return new DownloadProgress(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.filename = source["filename"];
	        this.status = source["status"];
	        this.statusText = source["statusText"];
	        this.progress = source["progress"];
	        this.totalBytes = source["totalBytes"];
	        this.downloadedBytes = source["downloadedBytes"];
	        this.error = source["error"];
	        this.outputPath = source["outputPath"];
	    }
	}
	export class LibraryItem {
	    id: string;
	    dramaId: string;
	    dramaName: string;
	    dramaCover: string;
	    episodeIndex: number;
	    episodeName: string;
	    platform: string;
	    filePath: string;
	    fileSize: number;
	    duration: number;
	    hasSubtitles: boolean;
	    // Go type: time
	    downloadedAt: any;
	    // Go type: time
	    lastPlayedAt?: any;
	    watchProgress: number;
	
	    static createFrom(source: any = {}) {
	        return new LibraryItem(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.dramaId = source["dramaId"];
	        this.dramaName = source["dramaName"];
	        this.dramaCover = source["dramaCover"];
	        this.episodeIndex = source["episodeIndex"];
	        this.episodeName = source["episodeName"];
	        this.platform = source["platform"];
	        this.filePath = source["filePath"];
	        this.fileSize = source["fileSize"];
	        this.duration = source["duration"];
	        this.hasSubtitles = source["hasSubtitles"];
	        this.downloadedAt = this.convertValues(source["downloadedAt"], null);
	        this.lastPlayedAt = this.convertValues(source["lastPlayedAt"], null);
	        this.watchProgress = source["watchProgress"];
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
	export class PlatformInfo {
	    id: string;
	    name: string;
	
	    static createFrom(source: any = {}) {
	        return new PlatformInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	    }
	}
	export class SubtitleInfo {
	    url: string;
	    language: string;
	    display_name: string;
	
	    static createFrom(source: any = {}) {
	        return new SubtitleInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.url = source["url"];
	        this.language = source["language"];
	        this.display_name = source["display_name"];
	    }
	}

}

export namespace models {
	
	export class DramaItem {
	    id: any;
	    name: string;
	    cover: any;
	    badge?: any;
	    chapterCount: any;
	    introduction: any;
	    tags: any;
	    playCount: any;
	
	    static createFrom(source: any = {}) {
	        return new DramaItem(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.cover = source["cover"];
	        this.badge = source["badge"];
	        this.chapterCount = source["chapterCount"];
	        this.introduction = source["introduction"];
	        this.tags = source["tags"];
	        this.playCount = source["playCount"];
	    }
	}
	export class BookList {
	    list: DramaItem[];
	    recommend?: DramaItem[];
	    preview?: DramaItem[];
	
	    static createFrom(source: any = {}) {
	        return new BookList(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.list = this.convertValues(source["list"], DramaItem);
	        this.recommend = this.convertValues(source["recommend"], DramaItem);
	        this.preview = this.convertValues(source["preview"], DramaItem);
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
	export class Subtitle {
	    subtitle: string;
	    display_name: string;
	    language: string;
	    type: string;
	
	    static createFrom(source: any = {}) {
	        return new Subtitle(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.subtitle = source["subtitle"];
	        this.display_name = source["display_name"];
	        this.language = source["language"];
	        this.type = source["type"];
	    }
	}
	export class Chapter {
	    eps: string;
	    index: number;
	    videoPath: string;
	    subtitle: Subtitle[];
	
	    static createFrom(source: any = {}) {
	        return new Chapter(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.eps = source["eps"];
	        this.index = source["index"];
	        this.videoPath = source["videoPath"];
	        this.subtitle = this.convertValues(source["subtitle"], Subtitle);
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
	
	export class PageInfo {
	    has_more: boolean;
	    pageNo?: number;
	    pageSize?: number;
	    tabKey?: any;
	    positionIndex?: any;
	    next?: any;
	    current?: number;
	    total?: number;
	    module_key?: string;
	    businessId?: any;
	    type?: string;
	    key?: any;
	    keyword?: string;
	
	    static createFrom(source: any = {}) {
	        return new PageInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.has_more = source["has_more"];
	        this.pageNo = source["pageNo"];
	        this.pageSize = source["pageSize"];
	        this.tabKey = source["tabKey"];
	        this.positionIndex = source["positionIndex"];
	        this.next = source["next"];
	        this.current = source["current"];
	        this.total = source["total"];
	        this.module_key = source["module_key"];
	        this.businessId = source["businessId"];
	        this.type = source["type"];
	        this.key = source["key"];
	        this.keyword = source["keyword"];
	    }
	}
	export class SearchResponse {
	    isMore?: boolean;
	    book: DramaItem[];
	    page_info?: PageInfo;
	
	    static createFrom(source: any = {}) {
	        return new SearchResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.isMore = source["isMore"];
	        this.book = this.convertValues(source["book"], DramaItem);
	        this.page_info = this.convertValues(source["page_info"], PageInfo);
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
	export class SeriesDetailResponse {
	    book?: DramaItem;
	    chapters: Chapter[];
	
	    static createFrom(source: any = {}) {
	        return new SeriesDetailResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.book = this.convertValues(source["book"], DramaItem);
	        this.chapters = this.convertValues(source["chapters"], Chapter);
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
	
	export class TabDataResponse {
	    isMore: boolean;
	    book: BookList;
	    page_info: PageInfo;
	
	    static createFrom(source: any = {}) {
	        return new TabDataResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.isMore = source["isMore"];
	        this.book = this.convertValues(source["book"], BookList);
	        this.page_info = this.convertValues(source["page_info"], PageInfo);
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
	export class TabFeedResponse {
	    isMore: boolean;
	    book: DramaItem[];
	    page_info: PageInfo;
	
	    static createFrom(source: any = {}) {
	        return new TabFeedResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.isMore = source["isMore"];
	        this.book = this.convertValues(source["book"], DramaItem);
	        this.page_info = this.convertValues(source["page_info"], PageInfo);
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
	export class TabItem {
	    type: string;
	    name: string;
	    tab_key: any;
	    position_index: any;
	
	    static createFrom(source: any = {}) {
	        return new TabItem(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.type = source["type"];
	        this.name = source["name"];
	        this.tab_key = source["tab_key"];
	        this.position_index = source["position_index"];
	    }
	}

}


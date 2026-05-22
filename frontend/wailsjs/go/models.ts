export namespace migrator {
	
	export class MigrationResult {
	    success: boolean;
	    message: string;
	    bytesMoved: number;
	    durationMs: number;
	
	    static createFrom(source: any = {}) {
	        return new MigrationResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.message = source["message"];
	        this.bytesMoved = source["bytesMoved"];
	        this.durationMs = source["durationMs"];
	    }
	}

}

export namespace models {
	
	export class AppSettings {
	    AutoScanOnStartup: boolean;
	    ConfirmBeforeMigration: boolean;
	    Theme: string;
	    CustomScanPaths: string[];
	
	    static createFrom(source: any = {}) {
	        return new AppSettings(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.AutoScanOnStartup = source["AutoScanOnStartup"];
	        this.ConfirmBeforeMigration = source["ConfirmBeforeMigration"];
	        this.Theme = source["Theme"];
	        this.CustomScanPaths = source["CustomScanPaths"];
	    }
	}
	export class CleanableItem {
	    Name: string;
	    Path: string;
	    Description: string;
	    SizeBytes: number;
	    Selected: boolean;
	    EnvKey: string;
	    Category: string;
	    RiskLevel: string;
	
	    static createFrom(source: any = {}) {
	        return new CleanableItem(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Name = source["Name"];
	        this.Path = source["Path"];
	        this.Description = source["Description"];
	        this.SizeBytes = source["SizeBytes"];
	        this.Selected = source["Selected"];
	        this.EnvKey = source["EnvKey"];
	        this.Category = source["Category"];
	        this.RiskLevel = source["RiskLevel"];
	    }
	}
	export class DiskInfo {
	    Letter: string;
	    TotalBytes: number;
	    FreeBytes: number;
	    UsedBytes: number;
	    UsedPercent: number;
	
	    static createFrom(source: any = {}) {
	        return new DiskInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Letter = source["Letter"];
	        this.TotalBytes = source["TotalBytes"];
	        this.FreeBytes = source["FreeBytes"];
	        this.UsedBytes = source["UsedBytes"];
	        this.UsedPercent = source["UsedPercent"];
	    }
	}
	export class Env {
	    Id: number;
	    Name: string;
	    Key: string;
	    Category: string;
	    Icon: string;
	    Description: string;
	    Website: string;
	    IsManaged: boolean;
	    // Go type: time
	    CreatedAt: any;
	    // Go type: time
	    UpdatedAt: any;
	
	    static createFrom(source: any = {}) {
	        return new Env(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Id = source["Id"];
	        this.Name = source["Name"];
	        this.Key = source["Key"];
	        this.Category = source["Category"];
	        this.Icon = source["Icon"];
	        this.Description = source["Description"];
	        this.Website = source["Website"];
	        this.IsManaged = source["IsManaged"];
	        this.CreatedAt = this.convertValues(source["CreatedAt"], null);
	        this.UpdatedAt = this.convertValues(source["UpdatedAt"], null);
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
	export class EnvInstance {
	    Id: number;
	    EnvId: number;
	    Version: string;
	    InstallPath: string;
	    IsDefault: boolean;
	    IsActive: boolean;
	    Source: string;
	    // Go type: time
	    DetectedAt: any;
	
	    static createFrom(source: any = {}) {
	        return new EnvInstance(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Id = source["Id"];
	        this.EnvId = source["EnvId"];
	        this.Version = source["Version"];
	        this.InstallPath = source["InstallPath"];
	        this.IsDefault = source["IsDefault"];
	        this.IsActive = source["IsActive"];
	        this.Source = source["Source"];
	        this.DetectedAt = this.convertValues(source["DetectedAt"], null);
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
	export class EnvPath {
	    Id: number;
	    EnvId: number;
	    InstanceId?: number;
	    Type: string;
	    Path: string;
	    SizeBytes: number;
	    IsMovable: boolean;
	    // Go type: time
	    LastSized: any;
	
	    static createFrom(source: any = {}) {
	        return new EnvPath(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Id = source["Id"];
	        this.EnvId = source["EnvId"];
	        this.InstanceId = source["InstanceId"];
	        this.Type = source["Type"];
	        this.Path = source["Path"];
	        this.SizeBytes = source["SizeBytes"];
	        this.IsMovable = source["IsMovable"];
	        this.LastSized = this.convertValues(source["LastSized"], null);
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
	export class EnvSummary {
	    Env: Env;
	    Instances: EnvInstance[];
	    Paths: EnvPath[];
	    TotalSize: number;
	    Health: string;
	
	    static createFrom(source: any = {}) {
	        return new EnvSummary(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Env = this.convertValues(source["Env"], Env);
	        this.Instances = this.convertValues(source["Instances"], EnvInstance);
	        this.Paths = this.convertValues(source["Paths"], EnvPath);
	        this.TotalSize = source["TotalSize"];
	        this.Health = source["Health"];
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
	export class HistoryEntry {
	    Id: number;
	    Action: string;
	    TargetEnv: string;
	    DetailsJson: string;
	    Success: boolean;
	    ErrorMessage: string;
	    // Go type: time
	    CreatedAt: any;
	
	    static createFrom(source: any = {}) {
	        return new HistoryEntry(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Id = source["Id"];
	        this.Action = source["Action"];
	        this.TargetEnv = source["TargetEnv"];
	        this.DetailsJson = source["DetailsJson"];
	        this.Success = source["Success"];
	        this.ErrorMessage = source["ErrorMessage"];
	        this.CreatedAt = this.convertValues(source["CreatedAt"], null);
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

}


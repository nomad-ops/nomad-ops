export interface Source {
    name: string,
    url: string,
    path: string,
    id?: string,
    branch: string,
    dataCenter: string,
    namespace?: string,
    region?: string,
    force?: boolean,
    paused?: boolean,
    created?: string,
    updated?: string,
    teams?: string[],
    deployKey?: string | string[],
    vaultToken?: string | string[],
    status?: SourceStatus | null
}
export interface SourceStatus {
    jobs?: {[jobID: string]: any}
    status: string,
    message?: string,
    lastCheckTime?: string
}
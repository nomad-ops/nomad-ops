import { Team } from "./Team";

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

export function userIsSourceMember(src: Source, teams: Team[], userID: string) : boolean {
    if (src.teams === undefined) {
        return true; // nobody owns this source => everybody is considered part of this source
    }
    if (src.teams.length === 0) {
        return true; // nobody owns this source => everybody is considered part of this source
    }

    for (let i = 0; i < src.teams.length; i++) {
        const element = src.teams[i];
        const team = teams.filter((t) => {
            return t.id === element;
        });
        if (team.length === 0) {
            continue;
        }
        if (team[0].members === undefined) {
            continue;
        }
        if (team[0].members.includes(userID) === true) {
            return true;
        }
    }

    return false;
}

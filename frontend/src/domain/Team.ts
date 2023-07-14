
export interface Team {
    id?: string,
    name: string,
    members?: string[],
    created?: string
}


export function userIsTeamMember(team: Team, userID: string): boolean {

    if (team.members === undefined) {
        return true;
    }
    if (team.members.length === 0) {
        return true;
    }
    if (team.members.includes(userID) === true) {
        return true;
    }

    return false;
}
import { Team } from "../domain/Team";
import pb from "./PocketBase";

const TeamService = {
    deleteTeam: (id: string) => {
        return pb.collection("teams").delete(id);
    },
    updateAssignedMembers: (id: string, members?: string[]) => {
        return pb.collection("teams").update(id, {
            members: members
        });
    },
    createTeam: (src: Team) => {
        return pb.collection("teams").create<Team>(src);
    },
}

export default TeamService;
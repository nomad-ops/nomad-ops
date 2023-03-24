import { Source } from "../domain/Source";
import pb from "./PocketBase";

const SourceService = {
    deleteSource: (id: string) => {
        return pb.collection("sources").delete(id);
    },
    createSource: (src: Source) => {
        src.status = {
            message: "init",
            status: "init"
        };
        return pb.collection("sources").create<Source>(src);
    },
    updateAssignedTeams: (id: string, teams?: string[]) => {
        return pb.collection("sources").update(id, {
            teams: teams
        });
    },
    syncSource: (id: string) => {
        return pb.send("/api/actions/sources/sync", {
            method: "POST",
            params: {
                id: id
            }
        });
    },
    pauseSource: (id: string, paused: boolean) => {
        return pb.collection("sources").update(id, {
            paused: paused
        });
    },
}

export default SourceService;
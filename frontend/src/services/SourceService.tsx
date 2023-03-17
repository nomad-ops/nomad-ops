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
    syncSource: (id: string) => {
        return pb.send("/api/actions/sources/sync", {
            method: "POST",
            params: {
                id: id
            }
        });
    },
    pauseSource: (id: string, pause: boolean) => {
        return pb.send("/api/actions/sources/pause", {
            method: "POST",
            params: {
                id: id,
                pause: pause
            }
        });
    },
}

export default SourceService;
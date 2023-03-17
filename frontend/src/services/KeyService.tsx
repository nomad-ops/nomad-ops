import { Key } from "../domain/Key";
import pb from "./PocketBase";

const KeyService = {
    deleteKey: (id: string) => {
        return pb.collection("keys").delete(id);
    },
    createKey: (key: Key) => {
        return pb.collection("keys").create<Key>(key);
    },
}

export default KeyService;
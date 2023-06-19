import { VaultToken } from "../domain/VaultToken";
import pb from "./PocketBase";

const VaultTokenService = {
    deleteVaultToken: (id: string) => {
        return pb.collection("vault_tokens").delete(id);
    },
    createVaultToken: (t: VaultToken) => {
        return pb.collection("vault_tokens").create<VaultToken>(t);
    },
}

export default VaultTokenService;
import { NomadURLs } from "../domain/NomadURLs";
import pb from "./PocketBase";

const NomadService = {
    getNomadURLs: async () => {
        const resp = await fetch(pb.buildUrl("/api/nomad/urls"), {
            method: "GET",
            headers: {
                "Authorization": pb.authStore.token
            },
        });
        return (await resp.json()) as NomadURLs;
    },
    getJobSummary: async (jobID: string, namespace: string) => {
        const resp = await fetch(pb.buildUrl("/api/nomad/proxy/v1/job/" + jobID + "/summary?namespace=" + namespace), {
            method: "GET",
            headers: {
                "Authorization": pb.authStore.token
            },
        });
        return await resp.json();
    },
    listAllocations: async (jobID: string, namespace: string) => {
        const resp = await fetch(pb.buildUrl("/api/nomad/proxy/v1/job/" + jobID + "/allocations?namespace=" + namespace), {
            method: "GET",
            headers: {
                "Authorization": pb.authStore.token
            },
        });
        if (resp.status !== 200 && resp.status !== 204) {
            return Promise.reject(resp);
        }
        return await resp.json();
    },
}

export default NomadService;
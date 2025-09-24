import { api } from "$lib/api/api";

export const getDNSSettings = () => {
    return api.get("/dns/config")
}

export const getLocalDomains = () => {
    return api.get("/dns/local-domains")
}
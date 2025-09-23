import { api } from "$lib/api/api";

export const getDNSSettings = () => {
    return api.get("/dns/config")
}
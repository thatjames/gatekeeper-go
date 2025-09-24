import { api } from "$lib/api/api";

export const getDNSSettings = () => {
    return api.get("/dns/config")
}

export const getLocalDomains = () => {
    return api.get("/dns/local-domains")
}

export const createLocalDomain = (localDomainRecord) => {
    return api.post("/dns/local-domains", localDomainRecord)
}

export const deleteLocalDomain = (domain) => {
    return api.delete(`/dns/local-domains/${domain}`)
}

export const updateLocalDomain = (originalName, localDomainRecord) => {
  return api.put(`/dns/local-domains/${originalName}`, { 
    domain: localDomainRecord.domain, 
    ip: localDomainRecord.ip 
  });
}
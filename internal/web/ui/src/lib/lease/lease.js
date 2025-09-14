import { api } from "$lib/api/api";

export const getLeases = () => {
  return api.get("/dhcp/leases");
};

export const deleteLease = (clientId) => {
  return api.delete(`/dhcp/leases/${clientId}`);
};

export const reserveLease = (clientId, ipAddress) => {
  return api.post(`/dhcp/leases/reserve`, {
    clientId: clientId,
    ip: ipAddress,
  });
};

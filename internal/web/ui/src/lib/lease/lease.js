import { api } from "$lib/api/api";

export const getLeases = () => {
  return api.get("/leases");
};

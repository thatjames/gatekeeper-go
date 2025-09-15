import { api } from "$lib/api/api";

export const getSettings = () => {
  return api.get("/dhcp/options");
};

export const saveSettings = (options) => {
  return api.put("/dhcp/options", options);
};

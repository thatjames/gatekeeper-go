import { api } from "$lib/api/api";

export const getSystemInfo = () => {
  return api.get("/system/info");
};

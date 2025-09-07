import { api } from "$lib/api/api";

export const getSettings = () => {
  return api.get("/options");
};

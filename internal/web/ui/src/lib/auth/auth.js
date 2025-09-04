import { api } from "$lib/api/api";

export const login = (username, password) => {
  return api.post("/login", { username: username, password: password });
};

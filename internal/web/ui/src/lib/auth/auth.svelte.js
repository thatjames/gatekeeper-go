import { api } from "$lib/api/api";

export const auth = $state({
  user: {},
  token: "",
});

export const login = (username, password) => {
  api
    .post("/login", { username: username, password: password })
    .then((resp) => {
      auth.token = resp.token;
    });
};

export const verify = () => {
  return api.get("/verify");
};

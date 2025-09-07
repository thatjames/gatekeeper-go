import { api } from "$lib/api/api";
import { jwtDecode } from "jwt-decode";

export const auth = $state({
  user: {},
  token: "",
});

let authHolder = JSON.parse(sessionStorage.getItem("auth"));
if (authHolder) {
  auth.user = authHolder.user;
  auth.token = authHolder.token;
}

export const login = async ({ username, password }) => {
  return api
    .post("/login", { username: username, password: password })
    .then((resp) => {
      auth.token = resp.token;
      auth.user = jwtDecode(resp.token);
      sessionStorage.setItem("auth", JSON.stringify(resp));
      return resp;
    })
    .catch((err) => {
      throw err;
    });
};

export const logout = () => {
  auth.token = "";
  auth.user = {};
  sessionStorage.removeItem("auth");
};

export const verify = () => {
  return api.get("/verify");
};

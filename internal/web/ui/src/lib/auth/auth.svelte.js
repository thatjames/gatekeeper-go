import { api, env } from "$lib/api/api";
import { Routes } from "$lib/common/routes";
import { jwtDecode } from "jwt-decode";
import { push } from "svelte-spa-router";

export const auth = $state({
  user: {},
  token: "",
});

let authHolder = JSON.parse(sessionStorage.getItem("auth"));
if (authHolder?.token) {
  auth.token = authHolder.token;
  auth.user = jwtDecode(authHolder.token);
}

// Viewers may only read; the server refuses their changes with 403.
export const isAdmin = () => auth.user?.Role === "admin";

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

export const logout = async () => {
  auth.token = "";
  auth.user = {};
  sessionStorage.removeItem("auth");
  await api.post("/logout");
  window.location.reload();
};

// The oauth_token cookie is HttpOnly, so ask the server who it belongs to.
// Plain fetch: a 401 here means "not logged in", not "reload to login".
export const restoreSession = async () => {
  try {
    const resp = await fetch(env.url + "me", { credentials: "include" });
    if (!resp.ok) {
      return false;
    }
    auth.user = await resp.json();
    return true;
  } catch {
    return false;
  }
};

export const loginFromSession = async () => {
  if (!(await restoreSession())) {
    return false;
  }
  push(Routes.Home);
  return true;
};

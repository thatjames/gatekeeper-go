import { auth } from "$lib/auth/auth.svelte";

const environments = {
  dev: {
    url: "http://localhost:8085/api/",
  },
  live: {
    url: "/api/",
  },
};

let env = import.meta.env.PROD ? environments.live : environments.dev;

export class API {
  async networkRequest(method, path, data) {
    const url = env.url.replace(/\/$/, "") + "/" + path.replace(/^\//, "");
    return fetch(url, {
      method: method,
      body: data ? JSON.stringify(data) : null,
      headers: auth.token ? { Authorization: "Bearer " + auth.token } : {},
    }).then((resp) => {
      if (resp.headers.get("Content-Type")?.includes("application/json")) {
        return resp.json();
      }
      return resp;
    });
  }

  get(path) {
    return this.networkRequest("get", path);
  }

  post(path, data) {
    return this.networkRequest("post", path, data);
  }

  put(path, data) {
    return this.networkRequest("put", path, data);
  }

  delete(path, data) {
    return this.networkRequest("delete", path, data);
  }
}

export const api = new API();

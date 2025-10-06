import { auth } from "$lib/auth/auth.svelte";
import { push } from "svelte-spa-router";
import { Routes } from "$lib/common/routes";
import { writable } from "svelte/store";

const environments = {
  dev: {
    url: "http://localhost:8085/api/v1/",
  },
  live: {
    url: "/api/v1/",
  },
};

let env = import.meta.env.PROD ? environments.live : environments.dev;

export let systemError = writable(null);

export class API {
  async networkRequest(method, path, data) {
    const url = env.url.replace(/\/$/, "") + "/" + path.replace(/^\//, "");
    try {
      const resp = await fetch(url, {
        method: method,
        body: data ? JSON.stringify(data) : null,
        headers: auth.token ? { Authorization: "Bearer " + auth.token } : {},
      });
      if (resp.status >= 200 && resp.status < 399) {
        // Clear any previous errors on successful request
        systemError.set(null);
        if (resp.headers.get("Content-Type")?.includes("application/json")) {
          return resp.json();
        }
        return resp;
      } else {
        // HTTP error (401, 404, 500, etc.)
        let error = await resp.json();
        throw error;
      }
    } catch (error) {
      // Check if it's a network error (TypeError is thrown for network failures)
      if (error instanceof TypeError && error.message.includes("fetch")) {
        console.error("Network error: Server unavailable", error);
        const errorMessage =
          "Server unavailable. Please check your connection.";
        systemError.set(errorMessage);
        push(Routes.Error);
      }
      // Re-throw HTTP errors or other errors
      throw error;
    }
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

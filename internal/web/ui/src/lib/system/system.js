import { api } from "$lib/api/api";
import { writable } from "svelte/store";

export let networkInterfaces = writable([]);

const getNetworkInterfaces = () => {
  api.get(`/system/interfaces`).then((resp) => {
    networkInterfaces.set(resp);
  });
};

export let version = writable("1.0.0");

export const getVersion = () => {
  api.get("/version").then((resp) => {
    version.set(resp.version);
  });
};

export const getSystemInfo = () => {
  return api.get("/system/info");
};

export let modules = writable([]);

export const loadModules = () => {
  api.get("/system/modules").then((resp) => {
    modules.set(resp);
    if (resp.length > 1) {
      getNetworkInterfaces();
    }
  });
};

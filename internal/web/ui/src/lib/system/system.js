import { api } from "$lib/api/api";
import { writable } from "svelte/store";

export let dhcpInterfaces = writable([]);

export const getInterface = (moduleName) => {
  api.get(`/system/interfaces/${moduleName}`).then((resp) => {
    dhcpInterfaces.set(resp);
  });
};

export let version = writable("1.0.0");

api.get("/version").then((resp) => {
  version.set(resp.version);
});

export const getSystemInfo = () => {
  return api.get("/system/info");
};

export let modules = writable([]);

export const loadModules = () => {
  api.get("/system/modules").then((resp) => {
    modules.set(resp);
  });
};

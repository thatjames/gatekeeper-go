import { api } from "$lib/api/api";
import { writable } from "svelte/store";

export let dhcpInterfaces = writable([]);

api.get("/system/interfaces/dhcp").then((resp) => {
  dhcpInterfaces.set(resp);
});

export let version = writable("1.0.0");

api.get("/system/version").then((resp) => {
  version.set(resp.version);
});

export const getSystemInfo = () => {
  return api.get("/system/info");
};

export let modules = writable([]);

api.get("/system/modules").then((resp) => {
  modules.set(resp)
})

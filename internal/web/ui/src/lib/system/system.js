import { api } from "$lib/api/api";
import { writable } from "svelte/store";

export let dhcpInterfaces = writable([]);

api.get("/system/interfaces/dhcp").then((resp) => {
  dhcpInterfaces.set(resp);
});

export const getSystemInfo = () => {
  return api.get("/system/info");
};

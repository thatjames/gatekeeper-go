import { writable } from "svelte/store";

export const isMenuOpen = writable(true);

export const isDropDownClicked = writable(false);

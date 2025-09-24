// In $lib/stores/menu.js
import { writable } from 'svelte/store';

// Check if screen is larger than mobile on initialization
const isLargeScreen = typeof window !== 'undefined' ? window.innerWidth >= 768 : false;

export const isMenuOpen = writable(isLargeScreen);
export const dropdownStates = writable({});
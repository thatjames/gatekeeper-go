<script>
  import { Button, DarkMode } from "flowbite-svelte";
  import { BarsFromLeftOutline, HomeOutline } from "flowbite-svelte-icons";
  import { onDestroy, onMount } from "svelte";
  import { push } from "svelte-spa-router";
  import { location } from "svelte-spa-router";
  import { isMenuOpen } from "$lib/stores/menu.js"; // Import the store
  import { logout } from "$lib/auth/auth.svelte";
  import { Routes } from "$lib/common/routes";

  let { menuOptions = [] } = $props();
  let componentId = Math.random().toString(36).substr(2, 9);
  let isLargeScreen = $state(false);
  const checkScreenSize = () => {
    isLargeScreen = window.innerWidth >= 1536;
  };

  const toggleMenu = (e) => {
    isMenuOpen.update((open) => !open);
  };

  const doLogout = () => {
    logout();
    window.location.reload();
  };

  onMount(() => {
    checkScreenSize();
    window.addEventListener("resize", checkScreenSize);
  });
</script>

<div>
  <div class="flex items-center relative 2xl:hidden top-1 left-1 z-50">
    <Button
      onclick={toggleMenu}
      class="menu-toggle p-2 m-2"
      aria-label="Toggle menu"
    >
      <BarsFromLeftOutline class="w-6 h-6" />
    </Button>
  </div>

  <div
    class="side-menu fixed top-0 left-0 h-full dark:bg-gray-800 bg-gray-200 text-white transition-transform duration-300 ease-in-out w-64 z-40 {$isMenuOpen ||
    isLargeScreen
      ? 'translate-x-0'
      : '-translate-x-full'}"
  >
    <div class="pt-16 p-4">
      <div class="flex">
        <img
          src="/logo.png"
          class="me-3 h-6 sm:h-9"
          alt="Gate keeper Logo of a padlock"
        />
        <span
          class="self-center text-xl font-semibold whitespace-nowrap dark:text-white text-black"
        >
          Gate<span class="text-primary-500">Keeper</span>
        </span>
      </div>
      <div>
        <div>
          <DarkMode class="w-full" />
        </div>
        {#each menuOptions as option}
          <div
            class="flex items-center justify-between rounded dark:hover:bg-gray-500 hover:bg-gray-100 hover:cursor-pointer transition-colors py-2 pl-3 my-1 {$location ===
            option.location
              ? 'bg-primary-500'
              : ''}"
            onclick={(e) => {
              e.stopPropagation();
              push(option.location);
            }}
          >
            <span class="text-black dark:text-gray-100">
              {option.label}
            </span>
            <svelte:component
              this={option.icon}
              class="w-5 h-5 mr-3 text-black dark:text-gray-100"
            />
          </div>
        {/each}
        <Button class="w-full" outline onclick={doLogout}>Logout</Button>
      </div>
    </div>
  </div>

  <!-- Fixed: Remove duplicate div -->
  <div
    class="transition-all duration-300 ease-in-out min-h-screen {$isMenuOpen ||
    isLargeScreen
      ? 'ml-64'
      : 'ml-0'}"
  >
    <div class="2xl:pt-20">
      <slot></slot>
    </div>
  </div>
</div>

<style>
  :global(body) {
    overflow-x: hidden;
  }
</style>

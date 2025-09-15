<script>
  import { Button, DarkMode, Dropdown } from "flowbite-svelte";
  import {
    BarsFromLeftOutline,
    ChevronDownOutline,
    ChevronRightOutline,
    HomeOutline,
  } from "flowbite-svelte-icons";
  import { onDestroy, onMount } from "svelte";
  import { push } from "svelte-spa-router";
  import { location } from "svelte-spa-router";
  import { isMenuOpen, isDropDownClicked } from "$lib/stores/menu.js"; // Import the store
  import { logout } from "$lib/auth/auth.svelte";
  import { Routes } from "$lib/common/routes";
  import { MenuComponent } from "$lib/common/menu-types";
  import { version } from "$lib/system/system";

  let { menuOptions = [], children } = $props();
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
    class="side-menu fixed top-0 left-0 h-full dark:bg-gray-800 bg-gray-200 text-white transition-transform duration-300 ease-in-out w-64 z-40 flex flex-col {$isMenuOpen ||
    isLargeScreen
      ? 'translate-x-0'
      : '-translate-x-full'}"
  >
    <!-- Header Section -->
    <div class="pt-16 p-4 flex-shrink-0">
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
        <DarkMode class="w-full" />
      </div>
    </div>

    <!-- Main Menu Content - Scrollable -->
    <div class="flex-grow overflow-y-auto p-4 pt-0">
      {#each menuOptions as option}
        <div>
          <div
            class="flex items-center justify-between rounded dark:hover:bg-gray-500 hover:bg-gray-100 hover:cursor-pointer transition-colors py-2 pl-3 my-1 {$location ===
            option.location
              ? 'bg-primary-500'
              : ''}"
            onclick={option.type !== MenuComponent.Dropdown
              ? (e) => {
                  push(option?.location);
                }
              : (e) => {
                  $isDropDownClicked = !$isDropDownClicked;
                }}
            role="button"
            tabindex="0"
            onkeydown={option.type !== MenuComponent.Dropdown
              ? (e) => e.key === "Enter" && push(option?.location)
              : (e) => {
                  if (e.key === "Enter") {
                    $isDropDownClicked = !$isDropDownClicked;
                  }
                }}
          >
            <span class="text-black dark:text-gray-100">
              {option?.label}
            </span>
            <span class="w-5 h-5 mr-2 text-black dark:text-gray-100">
              {#if option.type === MenuComponent.Dropdown}
                <div
                  class="transform transition-transform duration-300 ease-in-out {$isDropDownClicked
                    ? 'rotate-90'
                    : 'rotate-0'}"
                >
                  <ChevronRightOutline class="w-5 h-5" />
                </div>
              {:else}
                {@render option?.icon({})}
              {/if}
            </span>
          </div>

          <div
            class="transition-all duration-300 ease-in-out overflow-hidden ml-6 {option.type ===
              MenuComponent.Dropdown && $isDropDownClicked
              ? 'max-h-96 opacity-100'
              : 'max-h-0 opacity-0'}"
          >
            {#if option.type === MenuComponent.Dropdown}
              {#each option.items as item}
                <div
                  class="py-2 px-3 my-1 hover:bg-gray-200 dark:hover:bg-gray-600 rounded cursor-pointer text-black dark:text-gray-100 text-sm transition-colors duration-200 {$location ===
                  item.location
                    ? 'bg-primary-500'
                    : ''}"
                  onclick={() => push(item.location)}
                  role="button"
                  tabindex="0"
                  onkeydown={(e) => e.key === "Enter" && push(item.location)}
                >
                  <div class="flex items-center justify-between">
                    <span class="text-black dark:text-gray-100">
                      {item.label}
                    </span>
                    <span class="w-5 h-5 mr-2 text-black dark:text-gray-100">
                      {@render item.icon({})}
                    </span>
                  </div>
                </div>
              {/each}
            {/if}
          </div>
        </div>
      {/each}
    </div>

    <!-- Footer Section -->
    <div
      class="flex-shrink-0 p-4 border-t border-gray-300 dark:border-gray-600"
    >
      <Button class="w-full" outline onclick={doLogout}>Logout</Button>

      <!-- Additional footer content - customize as needed -->
      <div class="mt-3 text-center">
        <p class="text-xs text-gray-500 dark:text-gray-400">
          © 2025 GateKeeper
        </p>
        <p class="text-xs text-gray-500 dark:text-gray-400">
          Version {$version}
        </p>
      </div>
    </div>
  </div>

  <div
    class="transition-all duration-300 ease-in-out min-h-screen {$isMenuOpen ||
    isLargeScreen
      ? 'ml-64'
      : 'ml-0'}"
  >
    <div class="2xl:pt-20">
      {@render children?.()}
    </div>
  </div>
</div>

<style>
  :global(body) {
    overflow-x: hidden;
  }
</style>

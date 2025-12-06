<script>
  import { Button, DarkMode, Dropdown, Tooltip } from "flowbite-svelte";
  import {
    BarsFromLeftOutline,
    ChevronDownOutline,
    ChevronRightOutline,
    HomeOutline,
  } from "flowbite-svelte-icons";
  import { onDestroy, onMount } from "svelte";
  import { push } from "svelte-spa-router";
  import { location } from "svelte-spa-router";
  import { isMenuOpen, dropdownStates } from "$lib/stores/menu.js"; // Import dropdownStates from store
  import { logout } from "$lib/auth/auth.svelte";
  import { Routes } from "$lib/common/routes";
  import { MenuComponent } from "$lib/common/menu-types";
  import { version } from "$lib/system/system";

  let { menuOptions = [], children } = $props();
  let componentId = Math.random().toString(36).substr(2, 9);
  let isLargeScreen = $state(false);

  const checkScreenSize = () => {
    isLargeScreen = window.innerWidth >= 768;
  };

  const toggleMenu = (e) => {
    isMenuOpen.update((open) => !open);
  };

  const toggleDropdown = (optionLabel) => {
    dropdownStates.update((states) => ({
      ...states,
      [optionLabel]: !states[optionLabel],
    }));
  };

  // Check if current location matches any submenu item in a dropdown
  const getActiveDropdown = () => {
    return menuOptions.find(
      (option) =>
        option.type === MenuComponent.Dropdown &&
        option.items?.some((item) => item.location === $location),
    )?.label;
  };

  const doLogout = () => {
    logout();
    window.location.reload();
  };

  // Reactive statement to keep dropdown open if submenu is active
  // but don't close dropdowns when navigating to other pages
  $effect(() => {
    const activeDropdown = getActiveDropdown();
    if (activeDropdown && !dropdownStates[activeDropdown]) {
      dropdownStates[activeDropdown] = true;
    }
  });

  onMount(() => {
    checkScreenSize();
    window.addEventListener("resize", checkScreenSize);

    // Initialize dropdown states in store only if they haven't been set yet
    dropdownStates.update((states) => {
      const newStates = { ...states };

      menuOptions.forEach((option) => {
        if (
          option.type === MenuComponent.Dropdown &&
          newStates[option.label] === undefined
        ) {
          // Keep dropdown open if one of its subitems is currently active
          const hasActiveSubitem = option.items?.some(
            (item) => item.location === $location,
          );
          newStates[option.label] = hasActiveSubitem || false;
        }
      });

      return newStates;
    });
  });
</script>

<div>
  <div class="flex items-center relative md:fixed 2xl:hidden top-1 left-1 z-50">
    <Button
      onclick={toggleMenu}
      class="menu-toggle p-2 m-2"
      aria-label="Toggle menu"
    >
      <BarsFromLeftOutline class="w-6 h-6" />
    </Button>
    <Tooltip>Show/Hide the menu</Tooltip>
  </div>

  <div
    class="side-menu fixed z-40 dark:bg-gray-800 bg-gray-200 text-white transition-transform duration-300 ease-in-out flex flex-col {!isLargeScreen
      ? 'top-0 left-0 right-0 bottom-0 w-full h-full ' +
        ($isMenuOpen ? 'translate-y-0' : '-translate-y-full')
      : 'top-0 left-0 h-full w-64 ' +
        ($isMenuOpen ? 'translate-x-0' : '-translate-x-full')}"
  >
    <!-- Header Section -->
    <div class="{!isLargeScreen ? 'pt-20' : 'pt-16'} p-4 flex-shrink-0">
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
    <div
      class="flex-grow overflow-y-auto p-4 pt-0 {!isLargeScreen
        ? 'max-h-screen'
        : ''}"
    >
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
                  if (!isLargeScreen) {
                    isMenuOpen.set(false);
                  }
                }
              : (e) => {
                  toggleDropdown(option.label);
                }}
            role="button"
            tabindex="0"
            onkeydown={option.type !== MenuComponent.Dropdown
              ? (e) => {
                  if (e.key === "Enter") {
                    push(option?.location);
                    if (!isLargeScreen) {
                      isMenuOpen.set(false);
                    }
                  }
                }
              : (e) => {
                  if (e.key === "Enter") {
                    toggleDropdown(option.label);
                  }
                }}
          >
            <span class="text-black dark:text-gray-100">
              {option?.label}
            </span>
            <span class="w-5 h-5 mr-2 text-black dark:text-gray-100">
              {#if option.type === MenuComponent.Dropdown}
                <div
                  class="transform transition-transform duration-300 ease-in-out {$dropdownStates[
                    option.label
                  ]
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
              MenuComponent.Dropdown && $dropdownStates[option.label]
              ? 'max-h-96 opacity-100'
              : 'max-h-0 opacity-0'}"
          >
            {#if option.type === MenuComponent.Dropdown}
              {#each option.items as item}
                <div
                  id={item.label}
                  class="py-2 px-3 my-1 hover:bg-gray-200 dark:hover:bg-gray-600 rounded cursor-pointer text-black dark:text-gray-100 text-sm transition-colors duration-200 {$location ===
                  item.location
                    ? 'bg-primary-500'
                    : ''}"
                  onclick={() => {
                    push(item.location);
                    if (!isLargeScreen) {
                      isMenuOpen.set(false);
                    }
                  }}
                  role="button"
                  tabindex="0"
                  onkeydown={(e) => {
                    if (e.key === "Enter") {
                      push(item.location);
                      if (!isLargeScreen) {
                        isMenuOpen.set(false);
                      }
                    }
                  }}
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
    class="transition-all duration-300 ease-in-out min-h-screen {isLargeScreen &&
    $isMenuOpen
      ? 'ml-64'
      : 'ml-0'}"
  >
    <div class="md:pt-20">
      {@render children?.()}
    </div>
  </div>
</div>

<style>
  :global(body) {
    overflow-x: hidden;
  }
</style>

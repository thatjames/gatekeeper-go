<script>
  import { location } from "svelte-spa-router";
  import {
    Navbar,
    NavBrand,
    NavLi,
    NavUl,
    NavHamburger,
    Button,
    DarkMode,
  } from "flowbite-svelte";
  import { logout } from "$lib/auth/auth.svelte";

  $: activeUrl = "#" + $location;

  let activeClass =
    "text-white bg-green-700 md:bg-transparent md:text-green-700 md:dark:text-white dark:bg-green-600 md:dark:bg-transparent";
  let nonActiveClass =
    "text-gray-700 hover:bg-gray-100 md:hover:bg-transparent md:border-0 md:hover:text-green-700 dark:text-gray-400 md:dark:hover:text-white dark:hover:bg-gray-700 dark:hover:text-white md:dark:hover:bg-transparent";

  const doLogout = () => {
    logout();
    window.location.reload();
  };
</script>

<Navbar>
  <NavBrand href="#/">
    <img
      src="/logo.png"
      class="me-3 h-6 sm:h-9"
      alt="Gate keeper Logo of a padlock"
    />
    <span
      class="self-center text-xl font-semibold whitespace-nowrap dark:text-white"
    >
      Gate<span class="text-primary-500">Keeper</span>
    </span>
  </NavBrand>
  <NavHamburger />
  <NavUl
    {activeUrl}
    classes={{ active: activeClass, nonActive: nonActiveClass }}
  >
    <NavLi href="#/">Home</NavLi>
    <NavLi href="#/leases">Leases</NavLi>
    <NavLi href="#/settings">Settings</NavLi>
    <NavLi href="#/alert">Alert</NavLi>
    <NavLi href="#/avatar">Avatar</NavLi>
    <DarkMode class="mr-2" />
    <Button onclick={doLogout}>Logout</Button>
  </NavUl>
</Navbar>

<script>
  import Nav from "$components/Nav.svelte";
  import SideMenu from "$components/SideMenu.svelte";
  import { Routes } from "$lib/common/routes";
  import { getLeases } from "$lib/lease/lease";
  import { Button, P } from "flowbite-svelte";
  import {
    CogOutline,
    HomeOutline,
    LinkOutline,
    ServerOutline,
  } from "flowbite-svelte-icons";
  import { jwtDecode } from "jwt-decode";
  import Router from "svelte-spa-router";
  import wrap from "svelte-spa-router/wrap";

  const routes = {
    "/": wrap({
      asyncComponent: () => import("$scenes/home/HomeScreen.svelte"),
    }),
    "/leases": wrap({
      asyncComponent: () => import("$scenes/leases/LeaseLayout.svelte"),
    }),
    "/settings": wrap({
      asyncComponent: () => import("$scenes/settings/SettingsLayout.svelte"),
    }),
  };

  const menuOptions = [
    { label: "Home", location: Routes.Home, icon: HomeOutline },
    { label: "Leases", location: Routes.Leases, icon: ServerOutline },
    { label: "Settings", location: Routes.Settings, icon: CogOutline },
    // { label: "DNS", location: Routes.DNS, icon: LinkOutline },
  ];
</script>

<div class="flex flex-col gap-5">
  <SideMenu {menuOptions}>
    <div class="pl-6 pr-8">
      <Router {routes} />
    </div>
  </SideMenu>
</div>

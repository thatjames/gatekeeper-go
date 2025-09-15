<script>
  import Nav from "$components/Nav.svelte";
  import SideMenu from "$components/SideMenu.svelte";
  import { MenuComponent } from "$lib/common/menu-types";
  import { Routes } from "$lib/common/routes";
  import { getLeases } from "$lib/dhcp/lease";
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
    "/dhcp/leases": wrap({
      asyncComponent: () => import("$scenes/dhcp/LeaseLayout.svelte"),
    }),
    "/dhcp/settings": wrap({
      asyncComponent: () => import("$scenes/dhcp/OptionsLayout.svelte"),
    }),
  };

  const menuOptions = [
    { label: "Home", location: Routes.Home, icon: HomeOutline },
    {
      label: "DHCP",
      type: MenuComponent.Dropdown,
      items: [
        {
          label: "Leases",
          location: Routes.Leases,
          icon: ServerOutline,
        },
        { label: "Settings", location: Routes.Settings, icon: CogOutline },
      ],
    },
  ];
</script>

<div class="flex flex-col gap-5">
  <SideMenu {menuOptions}>
    <div class="pl-6 pr-8">
      <Router {routes} />
    </div>
  </SideMenu>
</div>

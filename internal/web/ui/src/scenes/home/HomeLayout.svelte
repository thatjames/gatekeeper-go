<script>
  import Nav from "$components/Nav.svelte";
  import SideMenu from "$components/SideMenu.svelte";
  import { MenuComponent } from "$lib/common/menu-types";
  import { Routes } from "$lib/common/routes";
  import { getLeases } from "$lib/dhcp/lease";
  import { modules } from "$lib/system/system";
  import { Button, P } from "flowbite-svelte";
  import {
    CogOutline,
    HomeOutline,
    LinkOutline,
    ServerOutline,
    BookOutline,
    ChartLineUpOutline,
  } from "flowbite-svelte-icons";
  import { jwtDecode } from "jwt-decode";
  import Router from "svelte-spa-router";
  import wrap from "svelte-spa-router/wrap";

  const routes = {
    "/": wrap({
      asyncComponent: () => import("$scenes/home/HomeScreen.svelte"),
    }),
    "/dhcp": wrap({
      asyncComponent: () => import("$scenes/dhcp/DHCPLayout.svelte"),
    }),
    "/dhcp/*": wrap({
      asyncComponent: () => import("$scenes/dhcp/DHCPLayout.svelte"),
    }),
    "/dns": wrap({
      asyncComponent: () => import("$scenes/dns/DNSLayout.svelte"),
    }),
    "/dns/*": wrap({
      asyncComponent: () => import("$scenes/dns/DNSLayout.svelte"),
    }),
  };

  const serviceConfig = {
    dhcp: {
      label: "DHCP",
      type: MenuComponent.Dropdown,
      items: [
        {
          label: "Leases",
          location: Routes.Leases,
          icon: ServerOutline,
        },
        { label: "Settings", location: Routes.DHCPSettings, icon: CogOutline },
      ],
    },
    dns: {
      type: MenuComponent.Dropdown,
      label: "DNS",
      items: [
        {
          label: "Local Domains",
          location: Routes.DNSLocalDomains,
          icon: BookOutline,
        },
        {
          label: "Statistics",
          location: Routes.DNSStatistics,
          icon: ChartLineUpOutline,
        },
        {
          label: "Settings",
          location: Routes.DNSSettings,
          icon: CogOutline,
        },
      ],
    },
  };

  const generateMenuOptions = (modulesList) => {
    const menuOptions = [
      { label: "Home", location: Routes.Home, icon: HomeOutline },
    ];

    if (modulesList && Array.isArray(modulesList)) {
      modulesList.forEach((module) => {
        const moduleKey = module.toLowerCase();
        if (serviceConfig[moduleKey]) {
          menuOptions.push(serviceConfig[moduleKey]);
        }
      });
    }
    return menuOptions;
  };

  let menuOptions = $state([
    { label: "Home", location: Routes.Home, icon: HomeOutline },
  ]);

  $effect(() => {
    menuOptions = generateMenuOptions($modules);
  });
</script>

<div class="flex flex-col gap-5">
  <SideMenu {menuOptions}>
    <div class="pl-6 pr-8">
      <Router {routes} />
    </div>
  </SideMenu>
</div>

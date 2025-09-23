<script>
  import Router from "svelte-spa-router";
  import wrap from "svelte-spa-router/wrap";

  const routes = {
    "/": wrap({
        asyncComponent: () => import("$scenes/dhcp/LeaseLayout.svelte")
    }),
     "/leases": wrap({
        asyncComponent: () => import("$scenes/dhcp/LeaseLayout.svelte")
    }),
    "/settings": wrap({
        asyncComponent: () => import("$scenes/dhcp/OptionsLayout.svelte")
    })
  }

</script>

<Router {routes} prefix="/dhcp" />
<script>
  import Router from "svelte-spa-router";
  import wrap from "svelte-spa-router/wrap";

  const routes = {
    "/settings": wrap({
      asyncComponent: () => import("./DNSSettings.svelte"),
    }),
    "/local-domains": wrap({
      asyncComponent: () => import("./DNSLocalDomains.svelte"),
    }),
    "/statistics": wrap({
      asyncComponent: () => import("./DNSStatistics.svelte"),
    }),
  };
</script>

<Router {routes} prefix="/dns" />

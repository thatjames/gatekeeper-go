<script>
  import Router from "svelte-spa-router";
  import wrap from "svelte-spa-router/wrap";

  const routes = {
    "/settings": wrap({
      asyncComponent: () => import("$scenes/dns/DNSSettings.svelte"),
    }),
    "/local-domains": wrap({
      asyncComponent: () => import("$scenes/dns/DNSLocalDomains.svelte"),
    }),
  };
</script>

<Router {routes} prefix="/dns" />

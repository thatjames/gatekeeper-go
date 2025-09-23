<script>
  import Router from "svelte-spa-router";
  import wrap from "svelte-spa-router/wrap";


    const routes = {
        "/settings": wrap({
            asyncComponent: () => import("$scenes/dns/DNSSettings.svelte")
        })
    }
</script>

<Router {routes} prefix="/dns" />
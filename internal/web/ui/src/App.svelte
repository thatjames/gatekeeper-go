<script>

import Router from "svelte-spa-router"
import Home from "$scenes/Home.svelte";

const routes = {
    "/": Home,
}
</script>

<Router {routes} />
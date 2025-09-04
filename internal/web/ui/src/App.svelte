<script>

import Router from "svelte-spa-router"
import Home from "$scenes/Home.svelte";
import Test from "$scenes/Test.svelte";

const routes = {
    "/": Home,
    "/test": Test
}
</script>

<Router {routes} />
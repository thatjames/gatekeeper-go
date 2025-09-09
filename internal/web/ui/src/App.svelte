<script>
  import { auth } from "$lib/auth/auth.svelte";
  import Router, { push } from "svelte-spa-router";
  import wrap from "svelte-spa-router/wrap";

  const routes = {
    "/auth/login": wrap({
      asyncComponent: () => import("$scenes/login/Login.svelte"),
    }),
    "/": wrap({
      asyncComponent: () => import("$scenes/home/HomeLayout.svelte"),
    }),
    "/*": wrap({
      asyncComponent: () => import("$scenes/home/HomeLayout.svelte"),
    }),
  };

  if (!auth.token) {
    push("/auth/login");
  }
</script>

<Router {routes} />

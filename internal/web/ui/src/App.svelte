<script>
  import { auth, restoreSession } from "$lib/auth/auth.svelte";
  import Router, { location, push } from "svelte-spa-router";
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

  // A password login keeps its token in sessionStorage; an OIDC login has
  // only the HttpOnly cookie, so ask the server before sending anyone away.
  if (!auth.token && $location !== "/auth/login") {
    restoreSession().then((ok) => {
      if (!ok) {
        push("/auth/login");
      }
    });
  }
</script>

<div class="h-full 2xl:w-4/5 m-auto">
  <Router {routes} />
</div>

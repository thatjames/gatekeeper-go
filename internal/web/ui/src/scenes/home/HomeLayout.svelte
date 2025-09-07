<script>
  import Nav from "$components/Nav.svelte";
  import { auth, logout } from "$lib/auth/auth.svelte";
  import { getLeases } from "$lib/lease/lease";
  import { Button, P } from "flowbite-svelte";
  import { jwtDecode } from "jwt-decode";
  import Router from "svelte-spa-router";
  import wrap from "svelte-spa-router/wrap";

  let leaseData = "";

  const routes = {
    "/leases": wrap({
      asyncComponent: () => import("$scenes/leases/LeaseLayout.svelte"),
    }),
  };

  getLeases().then((resp) => {
    leaseData = JSON.stringify(resp);
  });

  const doLogout = () => {
    logout();
    window.location.reload();
  };
</script>

<div class="flex flex-col gap-5">
  <Nav />
  <Router {routes} />
</div>

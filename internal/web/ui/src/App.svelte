<script>
  import { auth } from "$lib/auth/auth.svelte";
  import HomeLayout from "$scenes/home/HomeLayout.svelte";
  import Login from "$scenes/login/Login.svelte";
  import Router, { push } from "svelte-spa-router";

  const routes = {
    "/": HomeLayout,
    "/auth/login": Login,
  };

  if (!auth.token) {
    push("/auth/login");
  }
</script>

<div class="p-8 flex flex-col">
  <Router {routes} />
</div>

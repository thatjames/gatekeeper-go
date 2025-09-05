<script>
  import HomeLayout from "$scenes/home/HomeLayout.svelte";
  import Router from "svelte-spa-router";

  const routes = {
    "/": HomeLayout
  }
</script>

<div class="p-8 flex flex-col">
  <Router {routes}/>
</div>
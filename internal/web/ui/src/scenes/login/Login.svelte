<script>
  import { login } from "$lib/auth/auth.svelte";
  import { Routes } from "$lib/common/routes";
  import { getSystemInfo, loadModules } from "$lib/system/system";
  import { Button, Heading, Input, Label, P } from "flowbite-svelte";
  import { push } from "svelte-spa-router";
  let loginData = {};
  let errorText = $state("");
  const doLogin = (e) => {
    e.preventDefault();
    login(loginData)
      .then(() => {
        loadModules();
        push(Routes.Home);
      })
      .catch((err) => {
        errorText = err;
      });
  };
</script>

<div
  class="flex flex-col gap-5 p-5 md:h-full md:justify-center md:items-middle"
>
  <div class="flex gap-5 justify-center items-center">
    <img
      src="/logo.png"
      class="me-3 h-6 sm:h-9 logo"
      alt="Gate keeper Logo of a padlock"
    />
    <Heading tag="h1" class="text-center"
      >Gate<span class="text-primary-500">Keeper</span></Heading
    >
  </div>
  <form class="md:w-1/2 mx-auto" onsubmit={doLogin}>
    <div class="flex flex-col gap-2">
      <Label for="username" class="mb-2">Username</Label>
      <Input
        type="text"
        id="username"
        placeholder="username"
        required
        bind:value={loginData.username}
      />
      <Label for="password" class="mb-2">Password</Label>
      <Input
        type="password"
        id="password"
        placeholder="password"
        required
        bind:value={loginData.password}
      />
      <Button type="submit">Login</Button>
    </div>
  </form>
  <P class="text-center text-primary-600">{errorText}</P>
</div>

<style>
  .logo {
    width: 4rem;
    height: 4rem;
  }

  @media (max-width: 640px) {
    .logo {
      display: none;
    }
  }
</style>

<script>
  import { env } from "$lib/api/api";
  import { login, loginFromOIDCToken } from "$lib/auth/auth.svelte";
  import { Routes } from "$lib/common/routes";
  import { getExtraFeatures, features } from "$lib/system/system";
  import {
    Button,
    FloatingLabelInput,
    Heading,
    Input,
    Label,
    P,
  } from "flowbite-svelte";
  import { push } from "svelte-spa-router";
  let loginData = {};
  let errorText = $state("");
  const doLogin = (e) => {
    e.preventDefault();
    login(loginData)
      .then(() => {
        push(Routes.Home);
      })
      .catch((err) => {
        errorText = err.error;
      });
  };
  getExtraFeatures();

  const handleOIDCRedirect = () => {
    console.log("do redirect");
    window.location = env.url + "oidc-login";
  };

  if (document.cookie.includes("oauth_token")) {
    loginFromOIDCToken();
  }
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
  <div
    id="exampleWrapper"
    class="grid w-full items-end gap-6 md:grid-cols-3"
  ></div>

  <form class="md:w-1/2 mx-auto" onsubmit={doLogin}>
    <div class="flex flex-col gap-2">
      <FloatingLabelInput
        variant="outlined"
        id="username"
        name="username"
        type="text"
        bind:value={loginData.username}
        >Username
      </FloatingLabelInput>
      <FloatingLabelInput
        variant="outlined"
        id="password"
        name="password"
        type="password"
        bind:value={loginData.password}>Password</FloatingLabelInput
      >
      <Button type="submit">Login</Button>
      {#if $features.includes("oidc")}
        <Button outline color="dark" onclick={handleOIDCRedirect}
          >OIDC Login</Button
        >
      {/if}
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

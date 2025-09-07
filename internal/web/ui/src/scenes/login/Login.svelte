<script>
  import { login } from "$lib/auth/auth.svelte";
  import { Button, Heading, Input, Label, P } from "flowbite-svelte";
  import { push } from "svelte-spa-router";
  let loginData = {};
  let errorText = $state("");
  const doLogin = (e) => {
    e.preventDefault();
    login(loginData)
      .then(() => {
        push("/");
      })
      .catch((err) => {
        errorText = err;
      });
  };
</script>

<div class="flex flex-col gap-5 p-5 h-100 justify-center align-middle">
  <Heading tag="h1" class="text-center">GateKeeper Login</Heading>
  <form class="w-1/2 m-auto" onsubmit={doLogin}>
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

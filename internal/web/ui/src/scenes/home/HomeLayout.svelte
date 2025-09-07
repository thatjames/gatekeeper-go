<script>
  import { auth, logout } from "$lib/auth/auth.svelte";
  import { Button, P } from "flowbite-svelte";
  import { jwtDecode } from "jwt-decode";

  let userData = JSON.stringify(auth.user);
  console.log(jwtDecode(auth.token));

  const doLogout = () => {
    logout();
    window.location.reload();
  };
</script>

<div class="flex flex-col gap-5">
  <P>{userData}</P>
  <Button onclick={doLogout}>Logout</Button>
</div>

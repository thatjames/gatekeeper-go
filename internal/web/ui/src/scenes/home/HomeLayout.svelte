<script>
  import { auth } from "$lib/auth/auth.svelte";
  import { P } from "flowbite-svelte";
  import { jwtDecode } from "jwt-decode";

  let userData = JSON.stringify(auth.user);
  console.log(jwtDecode(auth.token));
</script>

<div class="flex flex-col gap-5">
  <P>{userData}</P>
</div>

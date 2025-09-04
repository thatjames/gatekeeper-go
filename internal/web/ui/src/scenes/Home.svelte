<script>
  import { login } from "$lib/auth/auth";

  const doLogin = () => {
    login("username", "password")
  };
</script>

<button on:click={doLogin}>Click Me</button>
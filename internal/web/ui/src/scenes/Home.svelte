<script>
  import { login } from "$lib/auth/auth";

  const doLogin = () => {
    login("test", "test")
  };
</script>

<button on:click={doLogin}>Click Me</button>
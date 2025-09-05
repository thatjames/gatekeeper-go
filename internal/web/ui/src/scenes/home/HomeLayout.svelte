<script>
  import { login, auth, verify } from "$lib/auth/auth.svelte";
  import {
    Button,
    P,
    Table,
    TableBodyCell,
    TableBodyRow,
    TableHead,
    TableHeadCell,
  } from "flowbite-svelte";

  let verifyVal = null;

  const doLogin = () => {
    login("admin", "admin");
  };

  const doVerify = () => {
    verify().then((resp) => {
      verifyVal = resp;
    });
  };
</script>

<div class="flex flex-col gap-2">
  <span>
    <Button onclick={doLogin}>Click me</Button>
  </span>
  {#if auth.token}
    <span><Button onclick={doVerify}>Verify</Button></span>
    {#if verifyVal}
      <Table>
        <TableHead>
          <TableHeadCell>Username</TableHeadCell>
          <TableHeadCell>Verified</TableHeadCell>
        </TableHead>
        <TableBodyRow>
          <TableBodyCell>{verifyVal.username}</TableBodyCell>
          <TableBodyCell>{verifyVal.valid}</TableBodyCell>
        </TableBodyRow>
      </Table>
    {/if}
  {/if}
</div>

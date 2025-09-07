<script>
  import {
    Heading,
    Table,
    TableBody,
    TableBodyCell,
    TableBodyRow,
    TableHead,
    TableHeadCell,
  } from "flowbite-svelte";

  export let leases = [];
</script>

<div class="flex flex-col gap-5">
  <Table hoverable>
    <TableHead>
      <TableHeadCell>Client Id</TableHeadCell>
      <TableHeadCell>Expiry</TableHeadCell>
      <TableHeadCell>Hostname</TableHeadCell>
      <TableHeadCell>IP Address</TableHeadCell>
      <TableHeadCell>State</TableHeadCell>
    </TableHead>
    {#if leases.length === 0}
      <TableBodyRow>
        <TableBodyCell colspan="5">No leases found</TableBodyCell>
      </TableBodyRow>
    {:else}
      <TableBody>
        {#each leases as lease}
          <TableBodyRow>
            <TableBodyCell>{lease.clientId}</TableBodyCell>
            <TableBodyCell>{lease.expiry}</TableBodyCell>
            <TableBodyCell>{lease.hostname || " - "}</TableBodyCell>
            <TableBodyCell>{lease.ip}</TableBodyCell>
            <TableBodyCell>{lease.state}</TableBodyCell>
          </TableBodyRow>
        {/each}
      </TableBody>
    {/if}
  </Table>
</div>

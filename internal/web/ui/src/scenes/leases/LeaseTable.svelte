<script>
  import {
    Heading,
    Modal,
    Table,
    TableBody,
    TableBodyCell,
    TableBodyRow,
    TableHead,
    TableHeadCell,
    Button,
  } from "flowbite-svelte";
  import { TrashBinOutline } from "flowbite-svelte-icons";

  let { leases = [] } = $props();
  let showModal = $state(false);
</script>

<Modal bind:open={showModal} title="Delete Lease">
  <div class="flex flex-col gap-5">
    <p>Are you sure you want to delete this lease?</p>
    <div class="flex gap-5 items-center justify-center">
      <Button
        color="alternative"
        onclick={() => (showModal = false)}
        class="focus:ring-0 focus:ring-offset-0">Cancel</Button
      >
      <Button
        color="red"
        onclick={() => (showModal = false)}
        class="focus:ring-0 focus:ring-offset-0">Delete</Button
      >
    </div>
  </div>
</Modal>

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
          <TableBodyRow class="group">
            <TableBodyCell
              ><div class="flex items-center justify-between">
                <span>{lease.clientId}</span>
                <button
                  class="opacity-0 group-hover:opacity-100 transition-opacity duration-200 ml-2 p-1 hover:bg-gray-100 rounded"
                  onclick={() => (showModal = true)}
                >
                  <TrashBinOutline class="shrink-0 h-6 w-6" />
                </button>
              </div></TableBodyCell
            >
            <TableBodyCell>{lease.expiry}</TableBodyCell>
            <TableBodyCell>{lease.hostname || " - "}</TableBodyCell>
            <TableBodyCell>
              {lease.ip}
            </TableBodyCell>
            <TableBodyCell>{lease.state}</TableBodyCell>
          </TableBodyRow>
        {/each}
      </TableBody>
    {/if}
  </Table>
</div>

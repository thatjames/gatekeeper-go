<script>
  import { deleteLease, getLeases, reserveLease } from "$lib/lease/lease";
  import {
    Button,
    Heading,
    Input,
    Label,
    Modal,
    P,
    Tooltip,
  } from "flowbite-svelte";
  import LeaseTable from "./LeaseTable.svelte";
  import { Tabs, TabItem } from "flowbite-svelte";

  let leases = $state([]);
  getLeases().then((resp) => {
    leases = resp;
  });

  let showModal = $state(false);
  let activeLease = $state(null);

  const onAddLeaseClick = () => {
    showModal = true;
  };

  const onDeleteLeaseClick = (clientId) => {
    deleteLease(clientId)
      .then((resp) => {
        activeLease = null;
        leases = resp;
      })
      .catch((err) => {
        console.log(err);
      });
  };

  const onLeaseClick = (lease) => {
    activeLease = $state.snapshot(lease);
  };

  const onLeaseReserveClick = (lease) => {
    reserveLease(lease.clientId, lease.ip)
      .then((resp) => {
        activeLease = null;
        leases = resp;
      })
      .catch((err) => {
        console.log(err);
      });
  };
</script>

{#if activeLease}
  <Modal open title="Lease Details" oncancel={() => (activeLease = null)}>
    <div class="flex flex-col gap-5 items-center justify-center">
      <form class="w-1/2 gap-3 grid grid-cols-2 items-center">
        <Label for="clientId" class="mb-2">Client Id:</Label>
        <Input
          id="clientId"
          placeholder="Client Id"
          disabled
          bind:value={activeLease.clientId}
        />
        <Label for="ipAddress" class="mb-2">IP Address</Label>
        <Input
          id="ipAddress"
          placeholder="IP Address"
          bind:value={activeLease.ip}
        />
        {#if activeLease.state === "Active"}
          <Button outline onclick={() => onLeaseReserveClick(activeLease)}
            >Reserve</Button
          >
          <Tooltip>Reserves the captured IP for the client</Tooltip>
        {/if}
        <Button
          outline
          color="red"
          class={activeLease.state === "Active" ? "" : "col-span-2"}
          onclick={() => onDeleteLeaseClick(activeLease.clientId)}
          >Release</Button
        >
        <Tooltip>Releases the IP</Tooltip>
      </form>
    </div>
  </Modal>
{/if}
<Modal bind:open={showModal} title="Add Lease">
  <div class="flex flex-col gap-5">
    <form class="flex flex-col gap-1">
      <Label for="clientId" class="mb-2">Client Id</Label>
      <Input id="clientId" placeholder="Client Id" required />
      <Label for="ipAddress" class="mb-2">IP Address</Label>
      <Input id="ipAddress" placeholder="IP Address" required />
    </form>
    <Button on:click={onAddLeaseClick}>Save</Button>
  </div>
</Modal>
<div class="flex flex-col gap-5">
  <Heading tag="h4">Leases</Heading>
  <Tabs>
    <TabItem title="Active Leases" open>
      <LeaseTable leases={leases.active} {onLeaseClick} />
    </TabItem>
    <TabItem title="Reserved Leases">
      <LeaseTable leases={leases.reserved} {onLeaseClick} />
      <Button outline class="mt-5" onclick={onAddLeaseClick}>Add Lease</Button>
    </TabItem>
  </Tabs>
</div>

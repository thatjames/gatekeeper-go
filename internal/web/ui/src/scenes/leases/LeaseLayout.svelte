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
    Helper,
  } from "flowbite-svelte";
  import LeaseTable from "./LeaseTable.svelte";
  import { Tabs, TabItem } from "flowbite-svelte";

  let leases = $state([]);
  getLeases().then((resp) => {
    leases = resp;
  });

  let showModal = $state(false);
  let activeLease = $state(null);
  let newLease = $state({});

  let fieldErrors = $state({});
  let generalError = $state("");

  const onAddLeaseClick = () => {
    showModal = true;
    fieldErrors = {};
    generalError = "";
    newLease = {};
  };

  const reserveNewLease = async () => {
    fieldErrors = {};
    generalError = "";

    reserveLease(newLease.clientId, newLease.ip)
      .then((resp) => {
        showModal = false;
        leases = resp;
        newLease = { clientId: "", ip: "" };
      })
      .catch((error) => {
        handleError(error);
      });
  };

  const onDeleteLeaseClick = async (clientId) => {
    try {
      const resp = await deleteLease(clientId);
      activeLease = null;
      leases = resp;
    } catch (err) {
      console.log(err);
    }
  };

  const onLeaseClick = (lease) => {
    activeLease = $state.snapshot(lease);
  };

  const onLeaseReserveClick = async (lease) => {
    try {
      const resp = await reserveLease(lease.clientId, lease.ip);
      activeLease = null;
      leases = resp;
    } catch (err) {
      console.log(err);
    }
  };

  const handleError = (error) => {
    if (error.fields && Array.isArray(error.fields)) {
      const newFieldErrors = {};
      error.fields.forEach((fieldError) => {
        newFieldErrors[fieldError.field] = fieldError.message;
      });
      fieldErrors = newFieldErrors;
    }

    if (error.error) {
      generalError = error.error;
    }
  };

  const clearFieldError = (fieldName) => {
    if (fieldErrors[fieldName]) {
      const newFieldErrors = { ...fieldErrors };
      delete newFieldErrors[fieldName];
      fieldErrors = newFieldErrors;
    }

    if (generalError) {
      generalError = "";
    }
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
          <Button outline onclick={() => onLeaseReserveClick(activeLease)}>
            Reserve
          </Button>
          <Tooltip>Reserves the captured IP for the client</Tooltip>
        {:else}
          <Button outline onclick={() => updateLease(activeLease)}>Save</Button>
          <Tooltip>Reserves the captured IP for the client</Tooltip>
        {/if}
        <Button
          outline
          color="red"
          onclick={() => onDeleteLeaseClick(activeLease.clientId)}
        >
          Release
        </Button>
        <Tooltip>Releases the IP</Tooltip>
      </form>
    </div>
  </Modal>
{/if}

<Modal bind:open={showModal} title="Add Reserved Lease">
  <div class="flex flex-col gap-5">
    <form class="flex flex-col gap-4">
      <div>
        <Label for="clientId" class="mb-2">Client Id</Label>
        <Input
          placeholder="Client Id"
          required
          bind:value={newLease.clientId}
          oninput={() => clearFieldError("clientId")}
          color={fieldErrors.clientId ? "red" : "default"}
        />
        {#if fieldErrors.clientId}
          <Helper class="mt-2" color="red">
            {fieldErrors.clientId}
          </Helper>
        {/if}
      </div>

      <div>
        <Label for="ipAddress" class="mb-2">IP Address</Label>
        <Input
          placeholder="IP Address"
          required
          bind:value={newLease.ip}
          oninput={() => clearFieldError("ip")}
          color={fieldErrors.ip ? "red" : "default"}
        />
        {#if fieldErrors.ip}
          <Helper class="mt-2" color="red">
            {fieldErrors.ip}
          </Helper>
        {/if}
      </div>
    </form>

    <Button
      disabled={fieldErrors.clientId || fieldErrors.ip}
      onclick={reserveNewLease}>Save</Button
    >
    {#if generalError}
      <Helper class="font-medium text-center" color="red">
        {generalError}
      </Helper>
    {/if}
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

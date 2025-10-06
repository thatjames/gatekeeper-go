<script>
  import {
    deleteLease,
    getLeases,
    reserveLease,
    updateLease,
  } from "$lib/dhcp/lease";
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

  $effect(() => {
    if (activeLease === null) {
      fieldErrors = {};
      generalError = "";
    }
  });

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

  const onDeleteLeaseClick = (clientId) => {
    deleteLease(clientId).then((resp) => {
      activeLease = null;
      leases = resp;
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
        handleError(err);
      });
  };

  const onUpdateLeaseClick = () => {
    updateLease(activeLease)
      .then((resp) => {
        activeLease = null;
        leases = resp;
      })
      .catch((err) => {
        handleError(err);
      });
  };

  const handleError = (error) => {
    if (error.fields && Array.isArray(error.fields)) {
      fieldErrors = error.fields.reduce((acc, fieldError) => {
        acc[fieldError.field] = fieldError.message;
        return acc;
      }, {});

      if (error.error) {
        generalError = error.error;
      }
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
      <form class="w-1/2 flex flex-col gap-4">
        <!-- Client ID Field -->
        <div>
          <Label for="clientId" class="mb-2">Client Id:</Label>
          <Input
            id="clientId"
            placeholder="Client Id"
            disabled
            bind:value={activeLease.clientId}
            color={fieldErrors.clientId ? "red" : "default"}
          />
          {#if fieldErrors.clientId}
            <Helper class="mt-2" color="red">
              {fieldErrors.clientId}
            </Helper>
          {/if}
        </div>

        <!-- IP Address Field -->
        <div>
          <Label for="ipAddress" class="mb-2">IP Address</Label>
          <Input
            id="ipAddress"
            placeholder="IP Address"
            bind:value={activeLease.ip}
            oninput={() => clearFieldError("ip")}
            color={fieldErrors.ip ? "red" : "default"}
          />
          {#if fieldErrors.ip}
            <Helper class="mt-2" color="red">
              {fieldErrors.ip}
            </Helper>
          {/if}
        </div>

        <!-- Action Buttons -->
        <div class="flex gap-3 mt-4 items-center justify-center">
          {#if activeLease.state === "Active"}
            <Button
              outline
              disabled={fieldErrors.ip}
              onclick={() => onLeaseReserveClick(activeLease)}
            >
              Reserve
            </Button>
            <Tooltip>Reserves the captured IP for the client</Tooltip>
          {:else}
            <Button
              disabled={fieldErrors.ip}
              outline
              onclick={onUpdateLeaseClick}>Save</Button
            >
            <Tooltip>Updates the reserved IP</Tooltip>
          {/if}

          <Button
            outline
            color="red"
            onclick={() => onDeleteLeaseClick(activeLease.clientId)}
          >
            Release
          </Button>
          <Tooltip>Releases the IP</Tooltip>
        </div>

        <!-- General Error -->
        {#if generalError}
          <Helper class="mt-2 font-medium text-center" color="red">
            {generalError}
          </Helper>
        {/if}
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
  <Heading tag="h3">Leases</Heading>
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

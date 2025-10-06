<script>
  import { networkInterfaces } from "$lib/system/system";
  import {
    Button,
    ButtonGroup,
    Helper,
    Input,
    Label,
    Modal,
    Select,
    Tooltip,
  } from "flowbite-svelte";
  import { EditOutline, TrashBinOutline } from "flowbite-svelte-icons";
  import { createEventDispatcher } from "svelte";

  const dispatch = createEventDispatcher();
  let { settings, externalErrors } = $props();
  let fieldErrors = $state({});
  let interfaceItems = $state(
    $networkInterfaces.map((interfaceItem) => {
      return {
        value: interfaceItem,
        name: interfaceItem,
      };
    }),
  );

  let showModal = $state(false);
  let newDomainNameServer = $state("");

  $effect(() => {
    fieldErrors =
      externalErrors?.fields?.reduce((acc, error) => {
        acc[error.field] = error.message;
        return acc;
      }, {}) || {};
  });

  const handleFormInput = (event) => {
    const fieldName = event.target.id;
    if (fieldName && fieldErrors[fieldName]) {
      fieldErrors = { ...fieldErrors, [fieldName]: null };

      dispatch("errorsCleared", { fieldName });
    }
  };

  const onEditNameServerClick = () => {
    showModal = true;
  };

  const addDomainNameServer = () => {
    settings.nameServers.push(newDomainNameServer);
    newDomainNameServer = "";
  };

  const closeModal = () => {
    showModal = false;
    dispatch("errorsCleared", "nameServers");
  };
</script>

<Modal bind:open={showModal} title="Add Upstream">
  <div class="flex flex-col justify-center gap-5">
    <div class="flex flex-col gap-2">
      <Label>Existing Domain Name Servers</Label>
      {#each settings?.nameServers as nameserver, index}
        <ButtonGroup>
          <Input
            type="text"
            bind:value={settings.nameServers[index]}
            tabindex="-1"
            autofocus={false}
          />
          <Button outline onclick={() => settings.nameServers.splice(index, 1)}>
            <TrashBinOutline />
          </Button>
        </ButtonGroup>
      {/each}
    </div>
    <div class="flex flex-col gap-2">
      <Label for="upstream">New Domain Name Server</Label>
      <Input
        type="text"
        id="upstream"
        placeholder="1.1.1.1"
        bind:value={newDomainNameServer}
        autofocus
        tabindex="0"
      />
    </div>
    <div class="grid grid-cols-2 gap-2">
      <Button
        outline
        disabled={!newDomainNameServer}
        onclick={addDomainNameServer}>Add</Button
      >
      <Button outline color="dark" onclick={closeModal}>Close</Button>
    </div>
  </div>
</Modal>
<form class="m:w-1/2" oninput={handleFormInput}>
  <div class="flex flex-col gap-2 md:grid md:grid-cols-2 md:gap-4">
    <Label for="interface" class="mb-2">Interface</Label>
    <div class="flex flex-col">
      <Select
        type="text"
        id="interface"
        placeholder="eth0"
        required
        items={interfaceItems}
        bind:value={settings.interface}
      />
      {#if fieldErrors.interface}
        <Helper class="mt-2 text-primary-500 dark:text-primary-500"
          >{fieldErrors.interface}</Helper
        >
      {/if}
    </div>

    <Label for="startAddr" class="mb-2">Start Address</Label>
    <div class="flex flex-col">
      <Input
        type="text"
        id="startAddr"
        placeholder="10.0.0.2"
        required
        bind:value={settings.startAddr}
      />
      {#if fieldErrors.startAddr}
        <Helper class="mt-2 text-primary-500 dark:text-primary-500"
          >{fieldErrors.startAddr}</Helper
        >
      {/if}
    </div>

    <Label for="endAddr" class="mb-2">End Address</Label>
    <div class="flex flex-col">
      <Input
        type="text"
        id="endAddr"
        placeholder="10.0.0.99"
        required
        bind:value={settings.endAddr}
      />
      {#if fieldErrors.endAddr}
        <Helper class="mt-2 text-primary-500 dark:text-primary-500"
          >{fieldErrors.endAddr}</Helper
        >
      {/if}
    </div>

    <Label for="leaseTTL" class="mb-2">Lease TTL</Label>
    <div class="flex flex-col">
      <Input
        type="text"
        id="leaseTTL"
        placeholder="300"
        required
        bind:value={settings.leaseTTL}
      />
      {#if fieldErrors.leaseTTL}
        <Helper class="mt-2 text-primary-500 dark:text-primary-500"
          >{fieldErrors.leaseTTL}</Helper
        >
      {/if}
    </div>

    <Label for="gateway" class="mb-2">Gateway</Label>
    <div class="flex flex-col">
      <Input
        type="text"
        id="gateway"
        placeholder="10.0.0.1"
        required
        bind:value={settings.gateway}
      />
      {#if fieldErrors.gateway}
        <Helper class="mt-2 text-primary-500 dark:text-primary-500"
          >{fieldErrors.gateway}</Helper
        >
      {/if}
    </div>

    <Label for="subnetMask" class="mb-2">Subnet Mask</Label>
    <div class="flex flex-col">
      <Input
        type="text"
        id="subnetMask"
        placeholder="255.255.255.0"
        required
        bind:value={settings.subnetMask}
      />
      {#if fieldErrors.subnetMask}
        <Helper class="mt-2 text-primary-500 dark:text-primary-500"
          >{fieldErrors.subnetMask}</Helper
        >
      {/if}
    </div>

    <Label for="domainName" class="mb-2">Domain Name</Label>
    <div class="flex flex-col">
      <Input
        type="text"
        id="domainName"
        placeholder="international-space-station"
        required
        bind:value={settings.domainName}
      />
      {#if fieldErrors.domainName}
        <Helper class="mt-2 text-primary-500 dark:text-primary-500"
          >{fieldErrors.domainName}</Helper
        >
      {/if}
    </div>

    <Label for="leaseFile" class="mb-2">Lease File</Label>
    <div class="flex flex-col">
      <Input
        type="text"
        id="leaseFile"
        placeholder="/etc/gatekeeper/leases.json"
        bind:value={settings.leaseFile}
      />
      {#if fieldErrors.leaseFile}
        <Helper class="mt-2 text-primary-500 dark:text-primary-500"
          >{fieldErrors.leaseFile}</Helper
        >
      {/if}
    </div>
    <Label for="nameServers" class="mb-2">Domain Name Servers</Label>
    <div class="flex flex-col">
      <ButtonGroup>
        <Input
          type="text"
          id="nameServers"
          placeholder="1.1.1.1,9.9.9.9"
          required
          disabled
          bind:value={settings.nameServers}
        />
        <Button
          color="primary"
          class="!rounded-r-lg"
          onclick={onEditNameServerClick}><EditOutline /></Button
        >
        <Tooltip>Edit Upstreams</Tooltip>
      </ButtonGroup>
      {#if fieldErrors.nameServers}
        <Helper class="mt-2 text-primary-500 dark:text-primary-500"
          >{fieldErrors.nameServers}</Helper
        >
      {/if}
    </div>
  </div>
</form>

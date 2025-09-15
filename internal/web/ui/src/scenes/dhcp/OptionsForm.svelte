<script>
  import { dhcpInterfaces } from "$lib/system/system";
  import { Helper, Input, Label, Select } from "flowbite-svelte";
  import { createEventDispatcher } from "svelte";

  const dispatch = createEventDispatcher();
  let { settings, externalErrors } = $props();
  let fieldErrors = $state({});
  let interfaceItems = $state(
    $dhcpInterfaces.map((interfaceItem) => {
      return {
        value: interfaceItem,
        name: interfaceItem,
      };
    }),
  );

  $effect(() => {
    fieldErrors =
      externalErrors?.fields?.reduce((acc, error) => {
        acc[error.field] = error.message;
        return acc;
      }, {}) || {};
  });

  function handleFormInput(event) {
    const fieldName = event.target.id;
    if (fieldName && fieldErrors[fieldName]) {
      fieldErrors = { ...fieldErrors, [fieldName]: null };

      dispatch("errorsCleared", { fieldName });
    }
  }
</script>

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
  </div>
</form>

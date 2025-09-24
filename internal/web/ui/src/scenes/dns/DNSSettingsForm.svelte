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

    <Label for="upstreams" class="mb-2">Upstreams</Label>
    <div class="flex flex-col">
      <Input
        type="text"
        id="upstreams"
        placeholder="1.1.1.1,9.9.9.9"
        required
        bind:value={settings.upstreams}
      />
      {#if fieldErrors.startAddr}
        <Helper class="mt-2 text-primary-500 dark:text-primary-500"
          >{fieldErrors.startAddr}</Helper
        >
      {/if}
    </div>
  </div>
</form>

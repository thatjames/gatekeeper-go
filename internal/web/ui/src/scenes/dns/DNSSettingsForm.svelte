<script>
  import { networkInterfaces } from "$lib/system/system";
  import { Helper, Input, Label, Select } from "flowbite-svelte";
  import { createEventDispatcher } from "svelte";

  const dispatch = createEventDispatcher();

  let { settings, externalErrors, generalError } = $props();

  let interfaceItems = $state(
    $networkInterfaces.map((interfaceItem) => {
      return {
        value: interfaceItem,
        name: interfaceItem,
      };
    }),
  );

  function handleFormInput(event) {
    const fieldName = event.target.id;
    if (fieldName && externalErrors[fieldName]) {
      dispatch("errorsCleared", fieldName); // Just pass the field name
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
      {#if externalErrors.interface}
        <Helper class="mt-2 text-primary-500 dark:text-primary-500">
          {externalErrors.interface}
        </Helper>
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
      {#if externalErrors.upstreams}
        <Helper class="mt-2 text-primary-500 dark:text-primary-500">
          {externalErrors.upstreams}
        </Helper>
      {/if}
    </div>
  </div>
</form>

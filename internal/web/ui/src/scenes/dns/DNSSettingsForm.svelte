<script>
  import { networkInterfaces } from "$lib/system/system";
  import {
    Button,
    ButtonGroup,
    FloatingLabelInput,
    Heading,
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
  let showModal = $state(false);
  let newUpstreamIP = $state("");

  let { settings, externalErrors, generalError } = $props();
  let fieldErrors = $state({});

  let interfaceItems = $state(
    $networkInterfaces.map((interfaceItem) => {
      return {
        value: interfaceItem,
        name: interfaceItem,
      };
    }),
  );

  const handleFormInput = (event) => {
    const fieldName = event.target.id;
    if (fieldName && externalErrors[fieldName]) {
      dispatch("errorsCleared", fieldName); // Just pass the field name
    }
  };

  const onAddButtonClick = () => {
    showModal = true;
  };

  const addUpstream = () => {
    settings.upstreams.push(newUpstreamIP);
  };

  const closeModal = () => {
    showModal = false;
    dispatch("errorsCleared", "upstreams");
  };

  $effect(() => {
    fieldErrors =
      externalErrors?.fields?.reduce((acc, error) => {
        acc[error.field] = error.message;
        return acc;
      }, {}) || {};
  });
</script>

<Modal bind:open={showModal} title="Add Upstream">
  <div class="flex flex-col justify-center gap-5">
    <div class="flex flex-col gap-2">
      <Label>Existing Upstreams</Label>
      {#each settings?.upstreams as upstream, index}
        <ButtonGroup>
          <Input
            type="text"
            bind:value={settings.upstreams[index]}
            tabindex="-1"
            autofocus={false}
          />
          <Button outline onclick={() => settings.upstreams.splice(index, 1)}>
            <TrashBinOutline />
          </Button>
        </ButtonGroup>
      {/each}
    </div>
    <div class="flex flex-col gap-2">
      <Label for="upstream">New Upstream</Label>
      <Input
        type="text"
        id="upstream"
        placeholder="1.1.1.1"
        bind:value={newUpstreamIP}
        autofocus
        tabindex="0"
      />
    </div>
    <div class="grid grid-cols-2 gap-2">
      <Button outline disabled={!newUpstreamIP} onclick={addUpstream}
        >Add</Button
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
        <Helper class="mt-2 text-primary-500 dark:text-primary-500">
          {fieldErrors.interface}
        </Helper>
      {/if}
    </div>

    <Label for="upstreams" class="mb-2">Upstreams</Label>
    <div class="flex flex-col">
      <ButtonGroup>
        <Input
          type="text"
          id="upstreams"
          placeholder="1.1.1.1,9.9.9.9"
          required
          disabled
          bind:value={settings.upstreams}
        />
        <Button color="primary" class="!rounded-r-lg" onclick={onAddButtonClick}
          ><EditOutline /></Button
        >
        <Tooltip>Edit Upstreams</Tooltip>
      </ButtonGroup>
      {#if fieldErrors.upstreams}
        <Helper class="mt-2 text-primary-500 dark:text-primary-500">
          {fieldErrors.upstreams}
        </Helper>
      {/if}
    </div>
  </div>
</form>

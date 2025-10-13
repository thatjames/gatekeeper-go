<script>
  import SimpleCard from "$components/SimpleCard.svelte";
  import {
    addDNSBlocklist,
    deleteDNSBlocklist,
    getDNSSettings,
    updateDNSSettings,
  } from "$lib/dns/dns";
  import {
    Button,
    Heading,
    P,
    Tooltip,
    Modal,
    Label,
    Input,
    Card,
    Badge,
    Spinner,
    Table,
    TableHead,
    TableHeadCell,
    TableBody,
    TableBodyRow,
    TableBodyCell,
    Listgroup,
    ListgroupItem,
    FloatingLabelInput,
    Helper,
    ButtonGroup,
    InputAddon,
  } from "flowbite-svelte";
  import {
    EditOutline,
    InfoCircleOutline,
    TrashBinOutline,
  } from "flowbite-svelte-icons";
  import DNSSettingsForm from "./DNSSettingsForm.svelte";

  let settings = $state({});
  let edit = $state(false);
  let errors = $state(null);
  let showLocalDomain = $state(false);
  let generalError = $state("");
  let activeDomain = $state({});
  let loading = $state(false);
  let showAddBlocklistModal = $state(false);
  let showDeleteBlocklistModal = $state(false);
  let deleteModalIndex = $state(0);
  let blockListError = $state("");
  let blocklistUrl = $state("");

  getDNSSettings().then((resp) => {
    settings = resp;
  });

  const onEditClick = () => {
    edit = true;
  };

  const onSaveClick = () => {
    loading = true;
    updateDNSSettings(settings)
      .then((resp) => {
        settings = resp;
        edit = false;
        // Clear errors on successful save
        generalError = "";
      })
      .catch((err) => {
        errors = err;
      })
      .finally(() => (loading = false));
  };

  const clearErrors = () => {
    errors = null;
  };

  const submitBlocklist = () => {
    addDNSBlocklist(blocklistUrl)
      .then((resp) => {
        settings.blocklist.push(blocklistUrl);
        showAddBlocklistModal = false;
      })
      .catch((error) => (blockListError = error.error));
  };

  const deleteBlocklistByIndex = (index) => {
    deleteDNSBlocklist(index).then(() => {
      showDeleteBlocklistModal = false;
      const newArrVal = settings.blocklist.filter((_, i) => i !== index);
      settings.blocklist = newArrVal;
    });
  };

  const clearModalValues = () => {
    blocklistUrl = "";
    blockListError = "";
  };
</script>

<div class="flex flex-col gap-5">
  {#if edit}
    <Heading tag="h3">Edit DNS Settings</Heading>
    <div class="w-4/5 mx-auto flex flex-col gap-5">
      <DNSSettingsForm
        bind:settings
        externalErrors={errors}
        on:errorsCleared={clearErrors}
      />
      <div class="grid grid-cols-2 gap-5">
        <Button
          outline
          onclick={onSaveClick}
          class="w-full mx-auto"
          disabled={errors !== null || loading}
          >{#if loading}<Spinner class="me-3" size="4" /> Saving{:else}Save{/if}</Button
        >
        <Button
          outline
          color="alternative"
          disabled={loading}
          onclick={() => {
            edit = false;
            fieldErrors = {};
            generalError = "";
          }}
          class="w-full mx-auto">Cancel</Button
        >
        <P
          class="text-primary-600 dark:text-primary-600 col-span-2 text-center"
        >
          {generalError}
        </P>
      </div>
    </div>
  {:else}
    <div class="flex gap-5 items-center">
      <Heading tag="h3">DNS Settings</Heading>
      <EditOutline
        class="shrink-0 h-6 w-6 text-primary-600 hover:text-primary-500 hover:cursor-pointer"
        onclick={onEditClick}
      />
      <Tooltip>Edit Settings</Tooltip>
    </div>
    <div class="flex flex-col gap-5 md:grid md:grid-cols-2">
      <Card class="p-5 min-w-full">
        <h5
          class="mb-2 text-2xl font-bold tracking-tight text-gray-900 dark:text-white"
        >
          Interface
        </h5>
        <p class="leading-tight font-normal text-gray-700 dark:text-gray-400">
          {settings?.interface}
        </p>
      </Card>

      <Card class="p-5 min-w-full">
        <h5
          class="mb-2 text-2xl font-bold tracking-tight text-gray-900 dark:text-white"
        >
          Upstreams
        </h5>
        <div class="flex flex-wrap gap-2">
          {#if settings?.upstreams?.length}
            {#each settings.upstreams as upstream}
              <Badge
                color="none"
                class="bg-primary-700 text-gray-900 dark:text-white"
                large>{upstream}</Badge
              >
            {/each}
          {:else}
            <p
              class="leading-tight font-normal text-gray-500 dark:text-gray-500 italic"
            >
              None configured
            </p>
          {/if}
        </div>
      </Card>
    </div>
    {#if settings.blocklist && settings.blocklist.length > 0}
      <Modal
        bind:open={showAddBlocklistModal}
        title="Add Blocklist"
        onclose={clearModalValues}
      >
        <div class="flex flex-col gap-5">
          <Label for="url">URL</Label>
          <div>
            <ButtonGroup class="w-full">
              <Input
                type="text"
                id="url"
                name="url"
                placeholder="/path/to/blocklist or http://host.com/blocklist"
                bind:value={blocklistUrl}
                color={blockListError ? "red" : "default"}
              />
              <InputAddon
                ><InfoCircleOutline /><Tooltip
                  >The target file must be a standard hosts file format</Tooltip
                ></InputAddon
              >
            </ButtonGroup>
            <Helper class="mt-2 !text-primary-500">{blockListError}</Helper>
          </div>
          <div class="grid-cols-2 gap-5">
            <Button outline onclick={submitBlocklist}>Add</Button>
            <Button
              outline
              color="dark"
              onclick={() => (showAddBlocklistModal = false)}>Cancel</Button
            >
          </div>
        </div>
      </Modal>
      <Modal
        bind:open={showDeleteBlocklistModal}
        title="Delete blocklist entry?"
      >
        <div class="flex flex-col gap-5">
          <P
            >Are you sure you want to delete the blocklist entries from <span
              class="font-mono text-primary-700"
              >{settings.blocklist[deleteModalIndex]}</span
            > ?</P
          >
          <div class="grid grid-cols-2 gap-2">
            <Button
              outline
              color="red"
              class="focus:outline-none focus:ring-0"
              onclick={() => deleteBlocklistByIndex(deleteModalIndex)}
              >Delete</Button
            >
            <Button
              outline
              color="dark"
              onclick={() => (showDeleteBlocklistModal = false)}>Cancel</Button
            >
          </div>
        </div>
      </Modal>
      <div class="flex flex-col gap-5">
        <Heading tag="h4">Blocklists</Heading>
        <P class="text-xs">Click on an entry below to delete it</P>
        <div>
          <Listgroup active>
            {#each settings.blocklist as listItem, index}
              <ListgroupItem
                class="hover:cursor-pointer"
                onclick={() => {
                  showDeleteBlocklistModal = true;
                  deleteModalIndex = index;
                }}>{listItem}</ListgroupItem
              >
            {/each}
          </Listgroup>
          <div>
            <Button
              outline
              class="mt-3"
              onclick={() => (showAddBlocklistModal = true)}>Add</Button
            >
            <Tooltip>Add a blocklist</Tooltip>
          </div>
        </div>
      </div>
    {/if}
  {/if}
</div>

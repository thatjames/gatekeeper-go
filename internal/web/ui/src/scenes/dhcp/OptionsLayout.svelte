<script>
  import SettingsDisplay from "$scenes/dhcp/OptionsDisplay.svelte";
  import SettingsForm from "./OptionsForm.svelte";
  import { getSettings, saveSettings } from "$lib/dhcp/options";
  import {
    Button,
    Heading,
    Input,
    Label,
    P,
    Textarea,
    Tooltip,
    Spinner,
  } from "flowbite-svelte";
  import { EditOutline } from "flowbite-svelte-icons";

  let settings = $state({});
  let settingsFormData = $state({});
  let edit = $state(false);
  let errors = $state(null);
  let interfaces = $state([]);
  let loading = $state(false);

  getSettings().then((resp) => {
    settings = resp;
  });

  const editSettings = () => {
    edit = true;
    settingsFormData = $state.snapshot(settings);
  };

  const onSaveClick = () => {
    loading = true;
    saveSettings(settingsFormData)
      .then((resp) => {
        settings = resp;
        edit = false;
      })
      .catch((err) => {
        errors = err;
      })
      .finally(() => (loading = false));
  };
</script>

<div class="flex flex-col gap-5 justify-center">
  {#if edit}
    <Heading tag="h3">Edit Settings</Heading>
    <div class="w-4/5 mx-auto flex flex-col gap-5">
      <SettingsForm
        externalErrors={errors}
        bind:settings={settingsFormData}
        on:errorsCleared={() => (errors = null)}
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
          onclick={() => (edit = false)}
          disabled={loading}
          class="w-full mx-auto">Cancel</Button
        >
        <P class="text-primary-600 dark:text-primary-600 col-span-2 text-center"
          >{errors?.error}</P
        >
      </div>
    </div>
  {:else}
    <div class="flex gap-5 items-center">
      <Heading tag="h3">Settings</Heading>
      <EditOutline
        class="shrink-0 h-6 w-6 text-primary-600 hover:text-primary-500 hover:cursor-pointer"
        onclick={editSettings}
      />
      <Tooltip>Edit Settings</Tooltip>
    </div>
    <SettingsDisplay {settings} />
  {/if}
</div>

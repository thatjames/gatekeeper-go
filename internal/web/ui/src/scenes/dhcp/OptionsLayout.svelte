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
  } from "flowbite-svelte";
  import { EditOutline } from "flowbite-svelte-icons";

  let settings = $state({});
  let settingsFormData = $state({});
  let edit = $state(false);
  let errors = $state(null);
  let interfaces = $state([]);

  getSettings().then((resp) => {
    settings = resp;
  });

  const editSettings = () => {
    edit = true;
    settingsFormData = $state.snapshot(settings);
  };

  const onSaveClick = () => {
    saveSettings(settingsFormData)
      .then((resp) => {
        settings = resp;
        edit = false;
      })
      .catch((err) => {
        console.log("Error?");
        errors = err;
      });
  };
</script>

<div class="flex flex-col gap-5 justify-center">
  {#if edit}
    <Heading tag="h4">Edit Settings</Heading>
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
          disabled={errors !== null}>Save</Button
        >
        <Button
          outline
          color="alternative"
          onclick={() => (edit = false)}
          class="w-full mx-auto">Cancel</Button
        >
        <P class="text-primary-600 dark:text-primary-600 col-span-2 text-center"
          >{errors?.error}</P
        >
      </div>
    </div>
  {:else}
    <div class="flex gap-5 items-center">
      <Heading tag="h4">Settings</Heading>
      <EditOutline
        class="shrink-0 h-6 w-6 text-primary-600 hover:text-primary-500 hover:cursor-pointer"
        onclick={editSettings}
      />
      <Tooltip>Edit Settings</Tooltip>
    </div>
    <SettingsDisplay {settings} />
  {/if}
</div>

<script>
  import SettingsDisplay from "$components/SettingsDisplay.svelte";
  import SettingsForm from "$components/SettingsForm.svelte";
  import {
    Button,
    Heading,
    Input,
    Label,
    Textarea,
    Tooltip,
  } from "flowbite-svelte";
  import { EditOutline } from "flowbite-svelte-icons";

  export let settings = {};
  let edit = false;

  const editSettings = () => {
    edit = true;
  };

  const saveSettings = () => {
    edit = false;
  };
</script>

<div class="flex flex-col gap-5 w-full">
  {#if edit}
    <Heading tag="h4">Settings</Heading>
    <SettingsForm bind:settings />
    <Button onclick={saveSettings} class="w-1/3 mx-auto">Save</Button>
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

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

<div class="flex flex-col gap-5 justify-center">
  {#if edit}
    <Heading tag="h4">Edit Settings</Heading>
    <div class="w-4/5 mx-auto flex flex-col gap-5">
      <SettingsForm bind:settings />
      <div class="grid grid-cols-2 gap-5">
        <Button outline onclick={saveSettings} class="w-full mx-auto"
          >Save</Button
        >
        <Button
          outline
          color="alternative"
          onclick={() => (edit = false)}
          class="w-full mx-auto">Cancel</Button
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

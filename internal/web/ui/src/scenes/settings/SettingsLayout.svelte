<script>
  import { Heading } from "flowbite-svelte";
  import SettingsScreen from "./SettingsScreen.svelte";
  import { getSettings } from "$lib/settings/settings";

  let settings = {};
  getSettings().then((resp) => {
    settings = resp;
  });
</script>

<div class="flex flex-col gap-10">
  <Heading tag="h4">Settings</Heading>
  <SettingsScreen {settings} />
</div>

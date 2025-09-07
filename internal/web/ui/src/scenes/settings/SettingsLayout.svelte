<script>
  import { Heading } from "flowbite-svelte";
  import SettingsScreen from "./SettingsScreen.svelte";
  import { getSettings } from "$lib/settings/settings";

  let settings = {};
  getSettings().then((resp) => {
    settings = resp;
  });
</script>

<SettingsScreen {settings} />

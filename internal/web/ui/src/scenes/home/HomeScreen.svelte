<script>
  import SimpleCard from "$components/SimpleCard.svelte";
  import { getSystemInfo } from "$lib/system/system";
  import { Heading } from "flowbite-svelte";

  export let systemInfo = {};
  getSystemInfo().then((resp) => {
    systemInfo = resp;
  });
</script>

<div class="flex flex-col gap-5">
  <Heading tag="h4">System Info</Heading>
  <div class="flex flex-col gap-5 md:grid md:grid-cols-2">
    {#each Object.entries(systemInfo) as [key, value]}
      <SimpleCard title={key} description={value} />
    {/each}
  </div>
</div>

<script>
  import SimpleCard from "$components/SimpleCard.svelte";
  import { getSystemInfo } from "$lib/system/system";
  import { Card, Heading } from "flowbite-svelte";

  export let systemInfo = {};
  getSystemInfo().then((resp) => {
    systemInfo = resp;
  });
</script>

<div class="flex flex-col gap-5">
  <Heading tag="h3">System Info</Heading>
  <div class="flex flex-col gap-5 md:grid md:grid-cols-2">
    {#each Object.entries(systemInfo) as [key, value]}
      <Card class="p-5 min-w-full">
        <h5
          class="mb-2 text-2xl font-bold tracking-tight text-gray-900 dark:text-white"
        >
          {key}
        </h5>
        <p class="leading-tight font-normal text-gray-700 dark:text-gray-400">
          {value}
        </p>
      </Card>
    {/each}
  </div>
</div>

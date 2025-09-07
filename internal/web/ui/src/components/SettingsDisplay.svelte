<script>
  import { Heading } from "flowbite-svelte";

  export let settings = {};
</script>

<div class="flex flex-col gap-5">
  <div class="flex flex-col gap-5 md:grid md:grid-cols-2">
    {#each Object.entries(settings) as [key, value]}
      <div
        class="flex flex-col gap-2 p-4 bg-gray-50 rounded-lg dark:bg-gray-700"
      >
        <div class="text-lg font-bold dark:text-white">{key}</div>
        <div class="text-sm">{value}</div>
      </div>
    {/each}
  </div>
</div>

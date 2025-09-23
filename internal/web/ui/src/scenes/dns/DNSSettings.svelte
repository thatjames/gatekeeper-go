<script>
  import SimpleCard from "$components/SimpleCard.svelte";
  import { getDNSSettings } from "$lib/dns/dns";
  import { Heading, Tooltip } from "flowbite-svelte";
  import { EditOutline } from "flowbite-svelte-icons";

  let settings = $state({});

  getDNSSettings().then((resp) => {
    settings = resp;
  });
</script>

<div class="flex flex-col gap-5">
  <div class="flex gap-5 items-center">
    <Heading tag="h4">DNS Settings</Heading>
    <EditOutline
      class="shrink-0 h-6 w-6 text-primary-600 hover:text-primary-500 hover:cursor-pointer"
    />
    <Tooltip>Edit Settings</Tooltip>
  </div>
  <div class="flex flex-col gap-5 md:grid md:grid-cols-2">
    <SimpleCard title="Interface" description={settings?.interface} />
    <SimpleCard
      title="Upstreams"
      description={settings?.upstreams?.replaceAll(",", " ")}
    />
  </div>
  {#if settings.localDomains}
    <Heading tag="h5">Local Domains</Heading>
    <div class="flex flex-col gap-5 md:grid md:grid-cols-2">
      {#each Object.entries(settings.localDomains) as [key, value]}
        <SimpleCard title={key} description={value} />
      {/each}
    </div>
  {/if}
</div>

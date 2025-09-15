<script>
  import { Heading } from "flowbite-svelte";
  import SimpleCard from "../../components/SimpleCard.svelte";

  export let settings = {};
</script>

<div class="flex flex-col gap-5">
  <div class="flex flex-col gap-5 md:grid md:grid-cols-2">
    <SimpleCard title="Interface" description={settings.interface} />
    <SimpleCard title="Start Address" description={settings.startAddr} />
    <SimpleCard title="End Address" description={settings.endAddr} />
    <SimpleCard title="Lease TTL" description={settings.leaseTTL} />
    <SimpleCard title="Gateway" description={settings.gateway} />
    <SimpleCard title="Subnet Mask" description={settings.subnetMask} />
    <SimpleCard title="Domain Name" description={settings.domainName} />
    <SimpleCard title="Lease File" description={settings.leaseFile} />
    <SimpleCard
      title="Domain Name Servers"
      description={settings.nameServers?.replaceAll(",", " ")}
    />
  </div>
</div>

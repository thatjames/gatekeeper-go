<script>
  import { Card, Heading } from "flowbite-svelte";
  import SimpleCard from "../../components/SimpleCard.svelte";

  export let settings = {};
</script>

<div class="flex flex-col gap-5">
  <div class="flex flex-col gap-5 md:grid md:grid-cols-2">
    <Card class="min-w-full p-5">
      <h5
        class="mb-2 text-2xl font-bold tracking-tight text-gray-900 dark:text-white"
      >
        Interface
      </h5>
      <p class="leading-tight font-normal text-gray-700 dark:text-gray-400">
        {settings.interface}
      </p>
    </Card>
    <Card class="min-w-full p-5">
      <h5
        class="mb-2 text-2xl font-bold tracking-tight text-gray-900 dark:text-white"
      >
        Start Address
      </h5>
      <p class="leading-tight font-normal text-gray-700 dark:text-gray-400">
        {settings.startAddr}
      </p>
    </Card>
    <Card class="min-w-full p-5">
      <h5
        class="mb-2 text-2xl font-bold tracking-tight text-gray-900 dark:text-white"
      >
        End Address
      </h5>
      <p class="leading-tight font-normal text-gray-700 dark:text-gray-400">
        {settings.endAddr}
      </p>
    </Card>
    <Card class="min-w-full p-5">
      <h5
        class="mb-2 text-2xl font-bold tracking-tight text-gray-900 dark:text-white"
      >
        Lease TTL
      </h5>
      <p class="leading-tight font-normal text-gray-700 dark:text-gray-400">
        {settings.leaseTTL}
      </p>
    </Card>
    <Card class="min-w-full p-5">
      <h5
        class="mb-2 text-2xl font-bold tracking-tight text-gray-900 dark:text-white"
      >
        Gateway
      </h5>
      <p class="leading-tight font-normal text-gray-700 dark:text-gray-400">
        {settings.gateway}
      </p>
    </Card>
    <Card class="min-w-full p-5">
      <h5
        class="mb-2 text-2xl font-bold tracking-tight text-gray-900 dark:text-white"
      >
        Subnet Mask
      </h5>
      <p class="leading-tight font-normal text-gray-700 dark:text-gray-400">
        {settings.subnetMask}
      </p>
    </Card>
    <Card class="min-w-full p-5">
      <h5
        class="mb-2 text-2xl font-bold tracking-tight text-gray-900 dark:text-white"
      >
        Domain Name
      </h5>
      <p class="leading-tight font-normal text-gray-700 dark:text-gray-400">
        {settings.domainName}
      </p>
    </Card>
    <Card class="min-w-full p-5">
      <h5
        class="mb-2 text-2xl font-bold tracking-tight text-gray-900 dark:text-white"
      >
        Lease File
      </h5>
      <p class="leading-tight font-normal text-gray-700 dark:text-gray-400">
        {settings.leaseFile}
      </p>
    </Card>
    <Card class="min-w-full p-5">
      <h5
        class="mb-2 text-2xl font-bold tracking-tight text-gray-900 dark:text-white"
      >
        Domain Name Servers
      </h5>
      <p class="leading-tight font-normal text-gray-700 dark:text-gray-400">
        {settings.nameServers?.replaceAll(",", " ")}
      </p>
    </Card>
  </div>
</div>

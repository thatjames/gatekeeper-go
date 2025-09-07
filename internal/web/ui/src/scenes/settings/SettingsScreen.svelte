<script>
  import { Button, Input, Label, Textarea } from "flowbite-svelte";

  export let settings = {};
</script>

<div class="flex flex-col gap-5">
  <form class="m:w-1/2 m-auto">
    <div class="flex flex-col gap-2 md:grid md:grid-cols-2 md:gap-4">
      <Label for="interface" class="mb-2">Interface</Label>
      <Input
        type="text"
        id="interface"
        placeholder="eth0"
        required
        bind:value={settings.interface}
      />
      <Label for="startAddr" class="mb-2">Start Address</Label>
      <Input
        type="text"
        id="startAddr"
        placeholder="10.0.0.2"
        required
        bind:value={settings.startAddr}
      />
      <Label for="endAddr" class="mb-2">End Address</Label>
      <Input
        type="text"
        id="endAddr"
        placeholder="10.0.0.99"
        required
        bind:value={settings.endAddr}
      />
      <Label for="leaseTTL" class="mb-2">Lease TTL</Label>
      <Input
        type="text"
        id="leaseTTL"
        placeholder="300"
        required
        bind:value={settings.leaseTTL}
      />
      <Label for="router" class="mb-2">Router</Label>
      <Input
        type="text"
        id="router"
        placeholder="10.0.0.1"
        required
        bind:value={settings.router}
      />
      <Label for="subnetMask" class="mb-2">Subnet Mask</Label>
      <Input
        type="text"
        id="subnetMask"
        placeholder="255.255.255.0"
        required
        bind:value={settings.subnetMask}
      />
      <Label for="domainName" class="mb-2">Domain Name</Label>
      <Input
        type="text"
        id="domainName"
        placeholder="international-space-station"
        required
        bind:value={settings.domainName}
      />
      <!-- <Label for="reservedLeases" class="mb-2">Reserved Leases</Label>
      <Textarea id="reservedLeases" required class="w-full"></Textarea> -->
      <Label for="leaseFile" class="mb-2">Lease File</Label>
      <Input
        type="text"
        id="leaseFile"
        placeholder="/etc/gatekeeper/leases.json"
        bind:value={settings.leaseFile}
      />
      <Button class="col-span-2" type="submit">Save</Button>
    </div>
  </form>
</div>

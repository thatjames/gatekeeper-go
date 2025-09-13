<script>
  import { getLeases } from "$lib/lease/lease";
  import { Button, Heading, Input, Label, Modal, P } from "flowbite-svelte";
  import LeaseTable from "./LeaseTable.svelte";
  import { Tabs, TabItem } from "flowbite-svelte";

  let leases = $state([]);
  getLeases().then((resp) => {
    leases = resp;
  });

  let showModal = $state(false);

  const onAddLeaseClick = () => {
    showModal = true;
  };
</script>

<Modal bind:open={showModal} title="Add Lease">
  <div class="flex flex-col gap-5">
    <form class="flex flex-col gap-1">
      <Label for="clientId" class="mb-2">Client Id</Label>
      <Input id="clientId" placeholder="Client Id" required />
      <Label for="ipAddress" class="mb-2">IP Address</Label>
      <Input id="ipAddress" placeholder="IP Address" required />
    </form>
    <Button on:click={onAddLeaseClick}>Save</Button>
  </div>
</Modal>
<div class="flex flex-col gap-5">
  <Heading tag="h4">Leases</Heading>
  <Tabs>
    <TabItem title="Active Leases" open>
      <LeaseTable leases={leases.active} />
    </TabItem>
    <TabItem title="Reserved Leases">
      <LeaseTable leases={leases.reserved} />
      <Button outline class="mt-5" onclick={onAddLeaseClick}>Add Lease</Button>
    </TabItem>
  </Tabs>
</div>

<script>
  import { getLeases } from "$lib/lease/lease";
  import { Heading, P } from "flowbite-svelte";
  import LeaseTable from "./LeaseTable.svelte";
  import { Tabs, TabItem } from "flowbite-svelte";

  let leases = [];
  getLeases().then((resp) => {
    leases = resp;
  });
</script>

<div class="flex flex-col gap-5">
  <Heading tag="h4">Leases</Heading>
  <Tabs>
    <TabItem title="Active Leases" open>
      <LeaseTable leases={leases.active} />
    </TabItem>
    <TabItem title="Reserved Leases">
      <LeaseTable leases={leases.reserved} />
    </TabItem>
  </Tabs>
</div>

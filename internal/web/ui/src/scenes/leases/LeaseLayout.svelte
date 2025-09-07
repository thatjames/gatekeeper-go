<script>
  import { getLeases } from "$lib/lease/lease";
  import { P } from "flowbite-svelte";
  import LeaseTable from "./LeaseTable.svelte";
  import { Tabs, TabItem } from "flowbite-svelte";

  let leases = [];
  getLeases().then((resp) => {
    leases = resp;
  });
</script>

<Tabs>
  <TabItem title="Active Leases" open>
    <LeaseTable leases={leases.active} />
  </TabItem>
  <TabItem title="Reserved Leases">
    <LeaseTable leases={leases.reserved} />
  </TabItem>
</Tabs>

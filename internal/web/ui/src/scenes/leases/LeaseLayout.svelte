<script>
  import { getLeases } from "$lib/lease/lease";
  import { P } from "flowbite-svelte";
  import LeaseTable from "./LeaseTable.svelte";
  import { Tabs, TabItem } from "flowbite-svelte";

  let leases = [];
  getLeases().then((resp) => {
    leases = resp;
    console.log(leases);
  });
</script>

<Tabs>
  <TabItem title="Active Leases" open>
    <LeaseTable
      title="Active Leases"
      open={leases.active}
      leases={leases.active}
    />
  </TabItem>
  <TabItem title="Reserved Leases">
    <LeaseTable title="Reserved Leases" leases={leases.reserved} />
  </TabItem>
</Tabs>

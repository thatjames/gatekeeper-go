<script>
  import { getLocalDomains } from "$lib/dns/dns";
  import {
    Heading,
    Table,
    TableHead,
    TableHeadCell,
    TableBody,
    TableBodyRow,
    TableBodyCell,
    Modal,
    Label,
    Input,
    Button,
    Tooltip,
  } from "flowbite-svelte";

  let localDomains = $state({});
  let activeDomain = $state({});
  let showLocalDomain = $state(false);
  let fieldErrors = $state({});
  let generalError = $state("");

  const onAddClick = () => {
    activeDomain = {};
    showLocalDomain = true;
  };

  $effect(() => {
    getLocalDomains().then((resp) => {
      localDomains = resp;
      console.log(localDomains);
    });
  });
</script>

<Modal
  bind:open={showLocalDomain}
  title={activeDomain.domain ? "Local Domain Details" : "Add New Local Domain"}
  oncancel={() => (activeDomain = {})}
>
  <div class="flex flex-col gap-5 items-center justify-center">
    <form class="w-1/2 flex flex-col gap-4">
      <div>
        <Label for="domain" class="mb-2">Domain:</Label>
        <Input
          id="clientId"
          placeholder="example.com"
          color={fieldErrors.clientId ? "red" : "default"}
          bind:value={activeDomain.domain}
        />
        {#if fieldErrors.clientId}
          <Helper class="mt-2" color="red">
            {fieldErrors.clientId}
          </Helper>
        {/if}
      </div>

      <div>
        <Label for="domain" class="mb-2">Address</Label>
        <Input
          id="domain"
          placeholder="127.0.0.1"
          oninput={() => clearFieldError("ip")}
          color={fieldErrors.ip ? "red" : "default"}
          bind:value={activeDomain.address}
        />
        {#if fieldErrors.ip}
          <Helper class="mt-2" color="red">
            {fieldErrors.ip}
          </Helper>
        {/if}
      </div>

      {#if activeDomain.domain}
        <div class="flex gap-3 mt-4 items-center justify-center">
          <Button disabled={fieldErrors.ip} outline>Save</Button>
          <Tooltip>Updates the local domain</Tooltip>

          <Button outline color="red">Delete</Button>
          <Tooltip>Deletes the local domain</Tooltip>
        </div>
      {:else}
        <div class="flex gap-3 mt-4 items-center justify-center">
          <Button disabled={fieldErrors.ip} outline>Save</Button>
          <Tooltip>Updates the local domain</Tooltip>

          <Button outline color="dark" onclick={() => (showLocalDomain = false)}
            >Cancel</Button
          >
          <Tooltip>Close this popup</Tooltip>
        </div>
      {/if}

      {#if generalError}
        <Helper class="mt-2 font-medium text-center" color="red">
          {generalError}
        </Helper>
      {/if}
    </form>
  </div>
</Modal>
<div class="flex flex-col gap-5">
  <Heading tag="h5">Local Domains</Heading>
  <Table hoverable>
    <TableHead>
      <TableHeadCell>Domain Name</TableHeadCell>
      <TableHeadCell>Address</TableHeadCell>
    </TableHead>
    <TableBody>
      {#if Object.keys(localDomains).length > 0}
        {#each Object.entries(localDomains) as [domain, address]}
          <TableBodyRow
            class="group hover:cursor-pointer"
            onclick={() => {
              showLocalDomain = true;
              activeDomain = { domain, address };
            }}
          >
            <TableBodyCell>{domain}</TableBodyCell>
            <TableBodyCell>{address}</TableBodyCell>
          </TableBodyRow>
        {/each}
      {:else}
        <TableBodyRow>
          <TableBodyCell colspan="2">No local domains defined</TableBodyCell>
        </TableBodyRow>
      {/if}
    </TableBody>
  </Table>
  <div>
    <Button outline onclick={onAddClick}>Add</Button>
  </div>
</div>

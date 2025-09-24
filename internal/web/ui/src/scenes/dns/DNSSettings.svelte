<script>
  import SimpleCard from "$components/SimpleCard.svelte";
  import { getDNSSettings } from "$lib/dns/dns";
  import {
    Button,
    Heading,
    P,
    Tooltip,
    Table,
    TableHead,
    TableHeadCell,
    TableBody,
    TableBodyRow,
    TableBodyCell,
    Modal,
    Label,
    Input,
  } from "flowbite-svelte";
  import { EditOutline } from "flowbite-svelte-icons";
  import DNSSettingsForm from "./DNSSettingsForm.svelte";

  let settings = $state({});
  let edit = $state(false);
  let errors = $state(null);
  let fieldErrors = $state({});
  let showLocalDomain = $state(false);
  let generalError = $state("");
  let activeDomain = $state({});

  getDNSSettings().then((resp) => {
    settings = resp;
  });

  const onEditClick = () => {
    edit = true;
  };

  const onSaveClick = () => {
    edit = false;
  };
</script>

<Modal
  bind:open={showLocalDomain}
  title="Local Domain Details"
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

      <div class="flex gap-3 mt-4 items-center justify-center">
        <Button disabled={fieldErrors.ip} outline>Save</Button>
        <Tooltip>Updates the reserved IP</Tooltip>

        <Button outline color="red">Delete</Button>
        <Tooltip>Releases the IP</Tooltip>
      </div>

      {#if generalError}
        <Helper class="mt-2 font-medium text-center" color="red">
          {generalError}
        </Helper>
      {/if}
    </form>
  </div>
</Modal>

<div class="flex flex-col gap-5">
  {#if edit}
    <Heading tag="h4">Edit DNS Settings</Heading>
    <div class="w-4/5 mx-auto flex flex-col gap-5">
      <DNSSettingsForm {settings} />
      <div class="grid grid-cols-2 gap-5">
        <Button
          outline
          onclick={onSaveClick}
          class="w-full mx-auto"
          disabled={errors !== null}>Save</Button
        >
        <Button
          outline
          color="alternative"
          onclick={() => (edit = false)}
          class="w-full mx-auto">Cancel</Button
        >
        <P class="text-primary-600 dark:text-primary-600 col-span-2 text-center"
          >{errors?.error}</P
        >
      </div>
    </div>
  {:else}
    <div class="flex gap-5 items-center">
      <Heading tag="h4">DNS Settings</Heading>
      <EditOutline
        class="shrink-0 h-6 w-6 text-primary-600 hover:text-primary-500 hover:cursor-pointer"
        onclick={onEditClick}
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
      <Table hoverable>
        <TableHead>
          <TableHeadCell>Domain Name</TableHeadCell>
          <TableHeadCell>Address</TableHeadCell>
        </TableHead>
        <TableBody>
          {#each Object.entries(settings.localDomains) as [domain, address]}
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
        </TableBody>
      </Table>
    {/if}
  {/if}
</div>

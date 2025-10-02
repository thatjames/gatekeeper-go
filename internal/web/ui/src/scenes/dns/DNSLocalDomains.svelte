<script>
  import {
    createLocalDomain,
    deleteLocalDomain,
    getLocalDomains,
    updateLocalDomain,
  } from "$lib/dns/dns";
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
    Helper,
    Tooltip,
  } from "flowbite-svelte";

  let localDomains = $state({});
  let activeDomain = $state({});
  let originalDomain = $state("");
  let showLocalDomain = $state(false);
  let fieldErrors = $state({});
  let generalError = $state("");

  const onAddClick = () => {
    activeDomain = {};
    originalDomain = "";
    showLocalDomain = true;
  };

  const createDomainClick = () => {
    createLocalDomain(activeDomain)
      .then((resp) => {
        localDomains = resp;
        showLocalDomain = false;
      })
      .catch((err) => handleError(err));
  };

  const deleteDomainClick = () => {
    const domainToDelete = originalDomain || activeDomain.domain;
    deleteLocalDomain(domainToDelete).then((resp) => {
      localDomains = resp;
      showLocalDomain = false;
    });
  };

  const updateLocalDomainClick = () => {
    const updatePayload = {
      ...activeDomain,
      originalDomain: originalDomain,
    };
    updateLocalDomain(originalDomain, activeDomain).then((resp) => {
      localDomains = resp;
      showLocalDomain = false;
    });
  };

  const handleError = (error) => {
    if (error.fields && Array.isArray(error.fields)) {
      fieldErrors = error.fields.reduce((acc, fieldError) => {
        acc[fieldError.field] = fieldError.message;
        return acc;
      }, {});

      if (error.error) {
        generalError = error.error;
      }
    }
  };

  const clearFieldError = (fieldName) => {
    if (fieldErrors[fieldName]) {
      const newFieldErrors = { ...fieldErrors };
      delete newFieldErrors[fieldName];
      fieldErrors = newFieldErrors;
    }

    if (generalError) {
      generalError = "";
    }
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
  oncancel={() => {
    activeDomain = {};
    originalDomain = "";
    fieldErrors = {};
    generalError = "";
  }}
>
  <div class="flex flex-col gap-5 items-center justify-center">
    <form class="w-1/2 flex flex-col gap-4">
      <div>
        <Label for="domain" class="mb-2">Domain:</Label>
        <Input
          id="domain"
          placeholder="example.com"
          color={fieldErrors.domain ? "red" : "default"}
          oninput={() => clearFieldError("domain")}
          bind:value={activeDomain.domain}
        />
        {#if fieldErrors.domain}
          <Helper class="mt-2" color="red">
            {fieldErrors.domain}
          </Helper>
        {/if}
      </div>

      <div>
        <Label for="domain" class="mb-2">Address</Label>
        <Input
          id="domain"
          placeholder="127.0.0.1"
          oninput={() => clearFieldError("address")}
          color={fieldErrors.ip ? "red" : "default"}
          bind:value={activeDomain.ip}
        />
        {#if fieldErrors.ip}
          <Helper class="mt-2" color="red">
            {fieldErrors.ip}
          </Helper>
        {/if}
      </div>

      {#if originalDomain}
        <div class="flex gap-3 mt-4 items-center justify-center">
          <Button
            disabled={fieldErrors.domain}
            outline
            onclick={updateLocalDomainClick}>Save</Button
          >
          <Tooltip>Updates the local domain</Tooltip>

          <Button outline color="red" onclick={deleteDomainClick}>Delete</Button
          >
          <Tooltip>Deletes the local domain</Tooltip>
        </div>
      {:else}
        <div class="flex gap-3 mt-4 items-center justify-center">
          <Button
            disabled={fieldErrors.domain}
            outline
            onclick={createDomainClick}>Create</Button
          >
          <Tooltip>Creates a new local domain</Tooltip>

          <Button
            outline
            color="dark"
            onclick={() => {
              showLocalDomain = false;
            }}>Cancel</Button
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
        {#each Object.entries(localDomains) as [domain, ip]}
          <TableBodyRow
            class="group hover:cursor-pointer"
            onclick={() => {
              showLocalDomain = true;
              activeDomain = { domain, ip };
              originalDomain = domain;
            }}
          >
            <TableBodyCell>{domain}</TableBodyCell>
            <TableBodyCell>{ip}</TableBodyCell>
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

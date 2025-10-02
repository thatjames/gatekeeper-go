<script>
  import SimpleCard from "$components/SimpleCard.svelte";
  import { getDNSSettings, updateDNSSettings } from "$lib/dns/dns";
  import {
    Button,
    Heading,
    P,
    Tooltip,
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
    updateDNSSettings(settings)
      .then((resp) => {
        settings = resp;
        edit = false;
        // Clear errors on successful save
        fieldErrors = {};
        generalError = "";
      })
      .catch((err) => handleError(err));
  };

  const handleError = (error) => {
    if (error.fields && Array.isArray(error.fields)) {
      fieldErrors = error.fields.reduce((acc, fieldError) => {
        acc[fieldError.field] = fieldError.message;
        return acc;
      }, {});
    }
    if (error.error) {
      generalError = error.error;
    }
  };

  // Simplified error clearing - just clear the specific field
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
</script>

<div class="flex flex-col gap-5">
  {#if edit}
    <Heading tag="h4">Edit DNS Settings</Heading>
    <div class="w-4/5 mx-auto flex flex-col gap-5">
      <DNSSettingsForm
        {settings}
        externalErrors={fieldErrors}
        {generalError}
        onerrorscleared={clearFieldError}
      />
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
          onclick={() => {
            edit = false;
            fieldErrors = {};
            generalError = "";
          }}
          class="w-full mx-auto">Cancel</Button
        >
        <P
          class="text-primary-600 dark:text-primary-600 col-span-2 text-center"
        >
          {errors?.error}
        </P>
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
  {/if}
</div>

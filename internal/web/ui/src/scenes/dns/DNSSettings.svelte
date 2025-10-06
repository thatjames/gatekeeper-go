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
    Card,
    Badge,
    Spinner,
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
  let loading = $state(false);

  getDNSSettings().then((resp) => {
    settings = resp;
  });

  const onEditClick = () => {
    edit = true;
  };

  const onSaveClick = () => {
    loading = true;
    updateDNSSettings(settings)
      .then((resp) => {
        settings = resp;
        edit = false;
        // Clear errors on successful save
        fieldErrors = {};
        generalError = "";
      })
      .catch((err) => handleError(err))
      .finally(() => (loading = false));
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
    <Heading tag="h3">Edit DNS Settings</Heading>
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
          disabled={errors !== null || loading}
          >{#if loading}<Spinner class="me-3" size="4" /> Saving{:else}Save{/if}</Button
        >
        <Button
          outline
          color="alternative"
          disabled={loading}
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
      <Heading tag="h3">DNS Settings</Heading>
      <EditOutline
        class="shrink-0 h-6 w-6 text-primary-600 hover:text-primary-500 hover:cursor-pointer"
        onclick={onEditClick}
      />
      <Tooltip>Edit Settings</Tooltip>
    </div>
    <div class="flex flex-col gap-5 md:grid md:grid-cols-2">
      <Card class="p-5 min-w-full">
        <h5
          class="mb-2 text-2xl font-bold tracking-tight text-gray-900 dark:text-white"
        >
          Interface
        </h5>
        <p class="leading-tight font-normal text-gray-700 dark:text-gray-400">
          {settings?.interface}
        </p>
      </Card>

      <Card class="p-5 min-w-full">
        <h5
          class="mb-2 text-2xl font-bold tracking-tight text-gray-900 dark:text-white"
        >
          Upstreams
        </h5>
        <div class="flex flex-wrap gap-2">
          {#if settings?.upstreams?.length}
            {#each settings.upstreams as upstream}
              <Badge
                color="none"
                class="bg-primary-700 text-gray-900 dark:text-white"
                large>{upstream}</Badge
              >
            {/each}
          {:else}
            <p
              class="leading-tight font-normal text-gray-500 dark:text-gray-500 italic"
            >
              None configured
            </p>
          {/if}
        </div>
      </Card>
    </div>
  {/if}
</div>

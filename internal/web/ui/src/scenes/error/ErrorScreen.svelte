<script>
  import { Card, Heading } from "flowbite-svelte";
  import { systemError } from "$lib/api/api";
  import { push } from "svelte-spa-router";
  import { Routes } from "$lib/common/routes";

  if (!$systemError) {
    push(Routes.Home);
  }
</script>

<Card
  class="bg-red-50 dark:bg-red-900/20 border-red-200 dark:border-red-800 p-5"
>
  <div class="flex flex-col gap-2">
    <Heading tag="h4">Server Communications Failure</Heading>
    <p class="text-red-800 dark:text-red-200">
      {$systemError}
    </p>
  </div>
</Card>

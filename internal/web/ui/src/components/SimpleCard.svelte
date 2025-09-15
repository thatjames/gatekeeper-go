<script>
  let { title, description } = $props();
</script>

<div class="flex flex-col gap-2 p-4 bg-gray-200 rounded-lg dark:bg-gray-700">
  <div class="text-lg font-bold dark:text-white">{title}</div>
  <div class="text-sm">{description}</div>
</div>

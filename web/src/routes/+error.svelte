<script>
  import { page } from "$app/stores";
</script>

<div class="py-4 text-center">
  <p class="text-3xl font-semibold">{$page.status}</p>
  <div class="h-2"></div>
  <p class="text-lg">{$page.error?.message}</p>
</div>

<script lang="ts">
  import { cn } from "$lib/utils";
  import type { HTMLImgAttributes } from "svelte/elements";

  const {
    class: className,
    loading = "lazy",
    ...restProps
  }: HTMLImgAttributes = $props();
</script>

<img
  class={cn(
    "inline-flex aspect-square items-center justify-center rounded border object-cover text-xs",
    className,
  )}
  {...restProps}
  {loading}
/>

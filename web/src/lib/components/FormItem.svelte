<script lang="ts">
  import { cn } from "$lib/utils";
  import type { Snippet } from "svelte";

  export type Props = {
    class?: string;
    children?: Snippet;
  };

  const { class: className, children }: Props = $props();
</script>

<div class={cn("flex flex-col gap-2", className)}>
  {@render children?.()}
</div>

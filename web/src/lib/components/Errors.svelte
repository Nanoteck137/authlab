<script lang="ts">
  type Props = {
    errors?: string[];
  };

  const { errors }: Props = $props();
</script>

{#if errors}
  {#each errors as error}
    {#if error.length > 0}
      <p class="text-sm text-destructive">{error}</p>
    {/if}
  {/each}
{/if}

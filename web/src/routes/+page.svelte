<script>
  import { Button, Input } from "@nanoteck137/nano-ui";
  import QuickCodeModal from "./QuickCodeModal.svelte";

  let openQuickCodeModal = $state(false);
</script>

<p class="p-4 text-xl">Home Page</p>

<Button
  onclick={() => {
    openQuickCodeModal = true;
  }}
>
  Open Quick Code Login
</Button>

<QuickCodeModal bind:open={openQuickCodeModal} />

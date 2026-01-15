<script lang="ts">
  import { getApiClient, handleApiError } from "$lib";
  import Errors from "$lib/components/Errors.svelte";
  import FormItem from "$lib/components/FormItem.svelte";
  import { Button, Dialog, Input, Label } from "@nanoteck137/nano-ui";
  import { zod } from "sveltekit-superforms/adapters";
  import { defaults, superForm } from "sveltekit-superforms/client";
  import { z } from "zod";
  import Spinner from "$lib/components/Spinner.svelte";
  import toast from "svelte-5-french-toast";
  import { goto } from "$app/navigation";

  const Schema = z.object({
    code: z.string().min(1),
  });

  export type Props = {
    open: boolean;
  };

  let { open = $bindable() }: Props = $props();
  const apiClient = getApiClient();

  $effect(() => {
    if (open) {
      reset({});
    }
  });

  const { form, errors, enhance, reset, submitting } = superForm(
    defaults(zod(Schema)),
    {
      SPA: true,
      validators: zod(Schema),
      dataType: "json",
      resetForm: true,
      async onUpdate({ form }) {
        if (form.valid) {
          const formData = form.data;

          const res = await apiClient.authLoginQuickCode({
            quickCode: formData.code,
          });
          if (!res.success) {
            return handleApiError(res.error);
          }

          open = false;

          toast.success("Successfully logged in");
          reset({ data: {} });
        }
      },
    },
  );
</script>

<Dialog.Root bind:open>
  <Dialog.Content>
    <Dialog.Header>
      <Dialog.Title>Enter the QuickCode</Dialog.Title>
    </Dialog.Header>

    <form class="flex flex-col gap-4" use:enhance>
      <FormItem>
        <Label for="code">Code</Label>
        <Input id="code" name="code" type="text" bind:value={$form.code} />
        <Errors errors={$errors.code} />
      </FormItem>

      <Dialog.Footer class="gap-2 sm:gap-0">
        <Button
          variant="outline"
          onclick={() => {
            open = false;
          }}
        >
          Close
        </Button>

        <Button type="submit" disabled={$submitting}>
          Submit
          {#if $submitting}
            <Spinner />
          {/if}
        </Button>
      </Dialog.Footer>
    </form>
  </Dialog.Content>
</Dialog.Root>

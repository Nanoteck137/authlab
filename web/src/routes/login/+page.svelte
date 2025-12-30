<script>
  import { getApiClient, handleApiError } from "$lib";
  import { Button } from "@nanoteck137/nano-ui";

  const { data } = $props();
  const apiClient = getApiClient();
</script>

<Button
  onclick={async () => {
    const res = await apiClient.authInitiate();
    if (!res.success) {
      return handleApiError(res.error);
    }

    window.location.replace(res.data.authUrl);
  }}
>
  Login with PocketID
</Button>

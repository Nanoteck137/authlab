<script lang="ts">
  import { invalidateAll } from "$app/navigation";
  import { getApiClient, handleApiError } from "$lib";
  import type { AuthInitiateQuick } from "$lib/api/types.js";
  import { onMount } from "svelte";

  const { data } = $props();
  const apiClient = getApiClient();

  let auth = $state<AuthInitiateQuick | null>(null);

  async function test() {
    const res = await apiClient.authInitiateQuick();
    if (!res.success) {
      return handleApiError(res.error);
    }

    auth = res.data;
  }

  // onMount(() => {
  //   test();
  // });

  $effect(() => {
    if (!auth) {
      test();
    }
  });

  $effect(() => {
    if (!auth) {
      return;
    }

    console.log("RUNNING");

    const expiresAtDate = new Date(auth.expiresAt);

    const pollInterval = setInterval(async () => {
      console.log("INTERVAL");
      try {
        if (new Date() > expiresAtDate) {
          clearInterval(pollInterval);

          auth = null;
          // win?.close();
          // resolve({ isSuccess: false, message: `authentication timeout` });

          return;
        }

        if (!auth) return;

        const res = await apiClient.getAuthTokenFromQuickCode(auth.code);
        if (!res.success) {
          clearInterval(pollInterval);
          // resolve({
          //   isSuccess: false,
          //   message: `authentication polling failed: ${res.error.message}`,
          // });
          return;
        }

        if (res.data.token) {
          clearInterval(pollInterval);
          console.log("Token", res.data.token);

          localStorage.setItem("token", res.data.token);
          invalidateAll();
        }
      } catch (error) {
        clearInterval(pollInterval);
        console.error("auth catch error", error);
        // resolve({
        //   isSuccess: false,
        //   message: `authentication failed for unknown reason`,
        // });
      }
    }, 2000);

    return () => {
      clearInterval(pollInterval);
      console.log("ENDING RUNNING");
    };
  });
</script>

<p>Code: {auth?.code}</p>

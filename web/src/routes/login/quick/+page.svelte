<script lang="ts">
  import { invalidateAll } from "$app/navigation";
  import { getApiClient, handleApiError } from "$lib";
  import type { AuthQuickConnectInitiate } from "$lib/api/types.js";
  import { onMount } from "svelte";

  const { data } = $props();
  const apiClient = getApiClient();

  let auth = $state<AuthQuickConnectInitiate | null>(null);

  async function test() {
    const res = await apiClient.authQuickConnectInitiate();
    if (!res.success) {
      return handleApiError(res.error);
    }

    auth = res.data;
    console.log(auth);
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

        const res = await apiClient.authGetQuickConnectStatus({
          code: auth.code,
          challenge: auth.challenge,
        });
        if (!res.success) {
          clearInterval(pollInterval);
          // resolve({
          //   isSuccess: false,
          //   message: `authentication polling failed: ${res.error.message}`,
          // });
          return handleApiError(res.error);
        }

        console.log("STATUS", res.data.status);

        if (res.data.status === "completed") {
          const res = await apiClient.authFinishQuickConnect({
            code: auth.code,
            challenge: auth.challenge,
          });
          if (!res.success) {
            clearInterval(pollInterval);
            auth = null;

            return handleApiError(res.error);
          }

          clearInterval(pollInterval);
          console.log("Token", res.data.token);
          localStorage.setItem("token", res.data.token);
          invalidateAll();
        } else if (res.data.status === "pending") {
        } else if (res.data.status === "expired") {
          clearInterval(pollInterval);
          auth = null;
        } else {
          clearInterval(pollInterval);
          auth = null;
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

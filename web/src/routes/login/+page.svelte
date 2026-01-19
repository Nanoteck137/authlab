<script lang="ts">
  import { invalidateAll } from "$app/navigation";
  import { getApiClient, handleApiError } from "$lib";
  import { Button } from "@nanoteck137/nano-ui";
  import toast from "svelte-5-french-toast";

  const { data } = $props();
  const apiClient = getApiClient();

  type LoginSuccess = {
    isSuccess: true;
    token: string;
  };
  type LoginError = {
    isSuccess: false;
    message: string;
  };
  type LoginResult = LoginSuccess | LoginError;

  async function loginWithPolling(providerId: string): Promise<LoginResult> {
    const res = await apiClient.authProviderInitiate({ providerId });
    if (!res.success) {
      handleApiError(res.error);
      return Promise.resolve({
        isSuccess: false,
        message: `failed to initiate auth: ${res.error.message}`,
      });
    }

    const { requestId, challenge, expiresAt, authUrl } = res.data;

    console.log("Request ID:", requestId);
    console.log("Opening authentication window...");

    const win = window.open(authUrl, "auth_window", "width=500,height=600");

    return new Promise((resolve, reject) => {
      const expiresAtDate = new Date(expiresAt);
      console.log("Request Expires At", expiresAtDate);

      const pollInterval = setInterval(async () => {
        try {
          if (new Date() > expiresAtDate) {
            clearInterval(pollInterval);

            win?.close();
            resolve({ isSuccess: false, message: `authentication timeout` });

            return;
          }

          const res = await apiClient.authGetProviderStatus({
            requestId: requestId,
            challenge: challenge,
          });
          if (!res.success) {
            clearInterval(pollInterval);
            resolve({
              isSuccess: false,
              message: `authentication polling failed: ${res.error.message}`,
            });
            return;
          }

          /*
          if (res.data.code) {
            clearInterval(pollInterval);

            win?.close();

            resolve({
              isSuccess: true,
              code: res.data.code!,
              state: requestId,
            });
          }
            */

          if (res.data.status === "completed") {
            clearInterval(pollInterval);

            const res = await apiClient.authFinishProvider({
              requestId,
              challenge,
            });
            if (!res.success) {
              resolve({
                isSuccess: false,
                message: `authentication failed to get code: ${res.error.message}`,
              });
              return;
            }

            win?.close();

            resolve({
              isSuccess: true,
              token: res.data.token,
            });
          } else if (res.data.status === "failed") {
            clearInterval(pollInterval);
            resolve({
              isSuccess: false,
              message: `authentication failed for unknown reason`,
            });
          } else if (res.data.status === "expired") {
            clearInterval(pollInterval);
            resolve({
              isSuccess: false,
              message: `authentication session expired`,
            });
          } else if (res.data.status === "pending") {
            return;
          } else {
            clearInterval(pollInterval);
            resolve({
              isSuccess: false,
              message: `authentication failed for unknown reason`,
            });
          }
        } catch (error) {
          clearInterval(pollInterval);
          console.error("auth catch error", error);
          resolve({
            isSuccess: false,
            message: `authentication failed for unknown reason`,
          });
        }
      }, 2000);
    });
  }
</script>

{#each data.providers as provider}
  <Button
    onclick={async () => {
      const res = await loginWithPolling(provider.id);
      if (!res.isSuccess) {
        toast.error(`login failed: ${res.message}`);
        return;
      }

      console.log("login", res);

      localStorage.setItem("token", res.token);
      invalidateAll();
    }}
  >
    Login with {provider.displayName}
  </Button>
{/each}

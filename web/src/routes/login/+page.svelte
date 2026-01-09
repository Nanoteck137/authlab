<script lang="ts">
  import { getApiClient, handleApiError } from "$lib";
  import { Button } from "@nanoteck137/nano-ui";
  import toast from "svelte-5-french-toast";

  const { data } = $props();
  const apiClient = getApiClient();

  type LoginSuccess = {
    isSuccess: true;
    code: string;
    state: string;
  };
  type LoginError = {
    isSuccess: false;
    message: string;
  };
  type LoginResult = LoginSuccess | LoginError;

  async function loginWithPolling(): Promise<LoginResult> {
    // Step 1: Initiate authentication
    const res = await apiClient.authInitiate();
    if (!res.success) {
      handleApiError(res.error);
      return Promise.resolve({
        isSuccess: false,
        message: `failed to initiate auth: ${res.error.message}`,
      });
    }

    const { sessionId, expiresAt, authUrl } = res.data;

    console.log("Session ID:", sessionId);
    console.log("Opening authentication window...");

    const win = window.open(authUrl, "auth_window", "width=500,height=600");

    return new Promise((resolve, reject) => {
      const expiresAtDate = new Date(expiresAt);
      console.log("Session Expires At", expiresAtDate);

      const pollInterval = setInterval(async () => {
        try {
          if (new Date() > expiresAtDate) {
            clearInterval(pollInterval);
            resolve({ isSuccess: false, message: `authentication timeout` });

            return;
          }

          const res = await apiClient.getAuthCode(sessionId);
          if (!res.success) {
            clearInterval(pollInterval);
            resolve({
              isSuccess: false,
              message: `authentication polling failed: ${res.error.message}`,
            });
            return;
          }

          console.log(res.data);

          if (res.data.code) {
            clearInterval(pollInterval);

            win?.close();

            resolve({
              isSuccess: true,
              code: res.data.code!,
              state: sessionId,
            });
          }

          // if (res.data.status === "completed") {
          // } else if (res.data.status === "failed") {
          //   clearInterval(pollInterval);
          //   resolve({
          //     isSuccess: false,
          //     message: `authentication failed for unknown reason`,
          //   });
          // } else if (res.data.status === "expired") {
          //   clearInterval(pollInterval);
          //   resolve({
          //     isSuccess: false,
          //     message: `authentication session expired`,
          //   });
          // } else if (res.data.status === "pending") {
          //   return;
          // } else {
          //   clearInterval(pollInterval);
          //   resolve({
          //     isSuccess: false,
          //     message: `authentication failed for unknown reason`,
          //   });
          // }
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

<Button
  onclick={async () => {
    const login = await loginWithPolling();
    if (!login.isSuccess) {
      toast.error(`login failed: ${login.message}`);
      return;
    }

    console.log("login", login);

    const res = await apiClient.authLoginWithCode(login);
    if (!res.success) {
      return handleApiError(res.error);
    }

    console.log(res);
  }}
>
  Login with PocketID
</Button>

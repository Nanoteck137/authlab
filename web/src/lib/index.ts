// place files you want to import through the `$lib` alias in this folder.

import { ApiClient } from "$lib/api/client";
import { getContext, setContext } from "svelte";
import toast from "svelte-5-french-toast";

export function handleApiError(err: {
  code: number;
  type: string;
  message: string;
}) {
  toast.error(`API Error: ${err.type} (${err.code}): ${err.message}`);
  console.error("API Error", err);
}

export function setApiClientAuth(apiClient: ApiClient, token?: string) {
  if (token) {
    apiClient.headers.set("Authorization", "Bearer " + token);
  } else {
    apiClient.headers.delete("Authorization");
  }
}

const API_CLIENT_KEY = Symbol("API_CLIENT");

export function setApiClient(baseUrl: string, token?: string) {
  const apiClient = new ApiClient(baseUrl);
  setApiClientAuth(apiClient, token);
  return setContext(API_CLIENT_KEY, apiClient);
}

export function getApiClient() {
  return getContext<ReturnType<typeof setApiClient>>(API_CLIENT_KEY);
}

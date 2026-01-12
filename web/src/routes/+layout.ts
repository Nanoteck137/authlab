import { PUBLIC_API_ADDRESS } from "$env/static/public";
import { setApiClientAuth } from "$lib";
import { ApiClient } from "$lib/api/client";
import type { GetMe } from "$lib/api/types";
import { error } from "@sveltejs/kit";
import type { LayoutLoad } from "./$types";

export const prerender = false;
export const ssr = false;

export const load: LayoutLoad = async ({ url }) => {
  console.log("LAYOUT");

  let addr = PUBLIC_API_ADDRESS;
  if (addr === "") {
    addr = url.origin;
  }

  const apiClient = new ApiClient(addr);
  const token = localStorage.getItem("token") ?? undefined;
  setApiClientAuth(apiClient, token);

  let user: GetMe | null = null;
  if (token) {
    const res = await apiClient.getMe();
    if (!res.success) {
      error(res.error.code, { message: res.error.message });
    }

    user = res.data;
  }

  return {
    apiClient,
    user,
  };
};

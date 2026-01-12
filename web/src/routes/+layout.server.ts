import type { GetMe } from "$lib/api/types";
import { error } from "@sveltejs/kit";
import type { LayoutServerLoad } from "./$types";

export const load: LayoutServerLoad = async ({ locals }) => {
  let user: GetMe | null = null;
  if (locals.token) {
    const res = await locals.apiClient.getMe();
    if (!res.success) {
      if (res.error.type !== "INVALID_AUTH") {
        throw error(res.error.code, { message: res.error.message });
      }

      user = null;
    } else {
      user = res.data;
    }
  }

  return {
    apiAddress: locals.apiAddress,
    userToken: locals.token,
    user,
  };
};

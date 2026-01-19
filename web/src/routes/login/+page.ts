import { error, redirect } from "@sveltejs/kit";
import type { PageLoad } from "./$types";

export const load: PageLoad = async ({ parent }) => {
  const data = await parent();

  if (data.user) {
    throw redirect(303, "/");
  }

  const providers = await data.apiClient.authGetProviders();
  if (!providers.success) {
    throw error(providers.error.code, { message: providers.error.message });
  }

  return {
    ...data,
    providers: providers.data.providers,
  };
};

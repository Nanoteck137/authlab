import { error } from "@sveltejs/kit";
import type { PageServerLoad } from "./$types";

export const load: PageServerLoad = async ({ locals }) => {
  const providers = await locals.apiClient.getAuthProviders();
  if (!providers.success) {
    throw error(providers.error.code, { message: providers.error.message });
  }

  return {
    providers: providers.data.providers,
  };
};

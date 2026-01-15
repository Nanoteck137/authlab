import { redirect } from "@sveltejs/kit";
import type { PageLoad } from "./$types";

export const load: PageLoad = async ({ parent }) => {
  const data = await parent();

  if (data.user) {
    throw redirect(303, "/");
  }

  return {
    ...data,
  };
};

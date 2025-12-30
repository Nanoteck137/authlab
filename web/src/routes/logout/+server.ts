import { setApiClientAuth } from "$lib";
import { redirect } from "@sveltejs/kit";

export const POST = async ({ cookies, locals }) => {
  cookies.delete("auth", {
    path: "/",
    sameSite: "strict",
  });
  locals.user = undefined;
  setApiClientAuth(locals.apiClient, undefined);

  throw redirect(303, "/login");
};

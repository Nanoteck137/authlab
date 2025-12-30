import { error, redirect, type RequestHandler } from "@sveltejs/kit";

export const GET: RequestHandler = async ({ locals, url }) => {
  const code = url.searchParams.get("code");
  if (!code) {
    throw error(400, "No code present in url");
  }

  const state = url.searchParams.get("state");
  if (!state) {
    throw error(400, "No state present in url");
  }

  const res = await locals.apiClient.authLoginWithCode({ code, state });
  if (!res.success) {
    throw error(res.error.code, { message: res.error.message });
  }

  console.log(res.data.token);

  throw redirect(301, "/");
};

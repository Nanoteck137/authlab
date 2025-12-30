import { SignupBody } from "$lib/api/types";
import { capitilize } from "$lib/utils";
import { error, redirect } from "@sveltejs/kit";
import { fail, setError, superValidate } from "sveltekit-superforms";
import { zod } from "sveltekit-superforms/adapters";
import { assert, type Equals } from "tsafe";
import type { z } from "zod";
import type { Actions, PageServerLoad } from "./$types";

const Body = SignupBody;
const schema = Body.extend({
  username: Body.shape.username,
  password: Body.shape.password,
  passwordConfirm: Body.shape.passwordConfirm,
});

// eslint-disable-next-line @typescript-eslint/no-unused-expressions
assert<Equals<keyof z.infer<typeof Body>, keyof z.infer<typeof schema>>>;

export const load: PageServerLoad = async () => {
  const form = await superValidate(zod(schema));

  return {
    form,
  };
};

export const actions: Actions = {
  default: async ({ locals, request }) => {
    const form = await superValidate(request, zod(schema));

    if (!form.valid) {
      return fail(400, { form });
    }

    const res = await locals.apiClient.signup(form.data);
    if (!res.success) {
      if (res.error.type === "VALIDATION_ERROR") {
        const extra = res.error.extra as Record<
          keyof z.infer<typeof schema>,
          string | undefined
        >;

        setError(form, "username", capitilize(extra.username ?? ""));
        setError(form, "password", capitilize(extra.password ?? ""));
        setError(
          form,
          "passwordConfirm",
          capitilize(extra.passwordConfirm ?? ""),
        );

        return fail(400, { form });
      } else if (res.error.type === "USER_ALREADY_EXISTS") {
        return setError(form, "username", "User already exists");
      } else {
        throw error(res.error.code, { message: res.error.message });
      }
    }

    throw redirect(302, "/login");
  },
};

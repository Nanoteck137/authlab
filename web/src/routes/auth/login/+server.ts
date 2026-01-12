import { setApiClientAuth } from "$lib";
import { json, type RequestHandler } from "@sveltejs/kit";

export const POST: RequestHandler = async ({ locals, cookies, request }) => {
  const body = await request.json();
  console.log(body);

  if (body.token) {
    setApiClientAuth(locals.apiClient, body.token);
    const user = await locals.apiClient.getMe();
    if (!user.success) {
      return json(
        { isSuccess: false, error: { message: "failed to get me" } },
        { status: 400 },
      );
    }

    const data = {
      token: body.token,
      user: {
        id: user.data.id,
      },
    };

    cookies.set("auth", JSON.stringify(data), {
      path: "/",
      sameSite: "strict",
    });

    return json({
      isSuccess: true,
    });
  }

  return json(
    { isSuccess: false, error: { message: "no token in body" } },
    { status: 400 },
  );
};

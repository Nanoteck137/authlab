import type { GetMe, Playlist } from "$lib/api/types";
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

  let quickPlaylistIds = [] as string[];
  let playlists: Playlist[] | null = null;

  if (user) {
    if (user.quickPlaylist) {
      const res = await locals.apiClient.getUserQuickPlaylistItemIds();
      if (!res.success) {
        throw error(res.error.code, { message: res.error.message });
      }

      quickPlaylistIds = res.data.trackIds;
    }

    {
      const res = await locals.apiClient.getPlaylists();
      if (!res.success) {
        throw error(res.error.code, { message: res.error.message });
      }

      playlists = res.data.playlists;
    }
  }

  /*
  let queueId: string | null = null;

  if (locals.user) {
    const res = await locals.apiClient.getDefaultQueue("authlab-web-app");
    if (!res.success) {
      throw error(res.error.code, { message: res.error.message });
    }

    queueId = res.data.id;
  }
    */

  return {
    apiAddress: locals.apiAddress,
    userToken: locals.token,
    user,
    quickPlaylistIds,
    userPlaylists: playlists,
    // queueId: queueId ?? "LOCAL",
  };
};

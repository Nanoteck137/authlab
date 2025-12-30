// See https://kit.svelte.dev/docs/types#app

import type { ApiClient } from "$lib/api/client";
import type { GetMe } from "$lib/api/types";

// for information about these interfaces
declare global {
  namespace App {
    interface Error {
      type?: string;
    }
    interface Locals {
      apiAddress: string;
      apiClient: ApiClient;
      token?: string;
      user?: GetMe;
    }
    // interface PageData {}
    // interface PageState {}
    // interface Platform {}
  }
}

export {};

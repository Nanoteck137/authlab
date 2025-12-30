import { z } from "zod";
import * as api from "./types";
import { BaseApiClient, createUrl, type ExtraOptions } from "./base-client";


export class ApiClient extends BaseApiClient {
  url: ClientUrls;

  constructor(baseUrl: string) {
    super(baseUrl);
    this.url = new ClientUrls(baseUrl);
  }
  
  authInitiate(options?: ExtraOptions) {
    return this.request("/api/v1/auth/initiate", "POST", api.AuthInitiate, z.any(), undefined, options)
  }
  
  authLoginWithCode(body: api.AuthLoginWithCodeBody, options?: ExtraOptions) {
    return this.request("/api/v1/auth/loginWithCode", "POST", api.AuthLoginWithCode, z.any(), body, options)
  }
  
  createApiToken(body: api.CreateApiTokenBody, options?: ExtraOptions) {
    return this.request("/api/v1/user/apitoken", "POST", api.CreateApiToken, z.any(), body, options)
  }
  
  deleteApiToken(id: string, options?: ExtraOptions) {
    return this.request(`/api/v1/user/apitoken/${id}`, "DELETE", z.undefined(), z.any(), undefined, options)
  }
  
  getAllApiTokens(options?: ExtraOptions) {
    return this.request("/api/v1/user/apitoken", "GET", api.GetAllApiTokens, z.any(), undefined, options)
  }
  
  getMe(options?: ExtraOptions) {
    return this.request("/api/v1/auth/me", "GET", api.GetMe, z.any(), undefined, options)
  }
  
  getSystemInfo(options?: ExtraOptions) {
    return this.request("/api/v1/system/info", "GET", api.GetSystemInfo, z.any(), undefined, options)
  }
  
  updateUserSettings(body: api.UpdateUserSettingsBody, options?: ExtraOptions) {
    return this.request("/api/v1/user/settings", "PATCH", z.undefined(), z.any(), body, options)
  }
}

export class ClientUrls {
  baseUrl: string;

  constructor(baseUrl: string) {
    this.baseUrl = baseUrl;
  }
  
  authInitiate() {
    return createUrl(this.baseUrl, "/api/v1/auth/initiate")
  }
  
  authLoginWithCode() {
    return createUrl(this.baseUrl, "/api/v1/auth/loginWithCode")
  }
  
  createApiToken() {
    return createUrl(this.baseUrl, "/api/v1/user/apitoken")
  }
  
  deleteApiToken(id: string) {
    return createUrl(this.baseUrl, `/api/v1/user/apitoken/${id}`)
  }
  
  getAllApiTokens() {
    return createUrl(this.baseUrl, "/api/v1/user/apitoken")
  }
  
  getMe() {
    return createUrl(this.baseUrl, "/api/v1/auth/me")
  }
  
  getSystemInfo() {
    return createUrl(this.baseUrl, "/api/v1/system/info")
  }
  
  updateUserSettings() {
    return createUrl(this.baseUrl, "/api/v1/user/settings")
  }
}

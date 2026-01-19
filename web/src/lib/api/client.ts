import { z } from "zod";
import * as api from "./types";
import { BaseApiClient, createUrl, type ExtraOptions } from "./base-client";


export class ApiClient extends BaseApiClient {
  url: ClientUrls;

  constructor(baseUrl: string) {
    super(baseUrl);
    this.url = new ClientUrls(baseUrl);
  }
  
  
  authClaimQuickConnectCode(body: api.AuthClaimQuickConnectCodeBody, options?: ExtraOptions) {
    return this.request("/api/v1/auth/quick-connect/claim", "POST", z.undefined(), z.any(), body, options)
  }
  
  authFinishProvider(body: api.AuthFinishProviderBody, options?: ExtraOptions) {
    return this.request("/api/v1/auth/providers/finish", "POST", api.AuthFinishProvider, z.any(), body, options)
  }
  
  authFinishQuickConnect(body: api.AuthFinishQuickConnectBody, options?: ExtraOptions) {
    return this.request("/api/v1/auth/quick-connect/finish", "POST", api.AuthFinishQuickConnect, z.any(), body, options)
  }
  
  authGetProviderStatus(body: api.AuthGetProviderStatusBody, options?: ExtraOptions) {
    return this.request("/api/v1/auth/provider/status", "POST", api.AuthGetProviderStatus, z.any(), body, options)
  }
  
  authGetProviders(options?: ExtraOptions) {
    return this.request("/api/v1/auth/providers", "GET", api.GetAuthProviders, z.any(), undefined, options)
  }
  
  authGetQuickConnectStatus(body: api.AuthGetQuickConnectStatusBody, options?: ExtraOptions) {
    return this.request("/api/v1/auth/quick-connect/status", "POST", api.AuthGetQuickConnectStatus, z.any(), body, options)
  }
  
  authProviderInitiate(body: api.AuthInitiateBody, options?: ExtraOptions) {
    return this.request("/api/v1/auth/providers/initiate", "POST", api.AuthInitiate, z.any(), body, options)
  }
  
  authQuickConnectInitiate(options?: ExtraOptions) {
    return this.request("/api/v1/auth/quick-connect/initiate", "POST", api.AuthQuickConnectInitiate, z.any(), undefined, options)
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
  
  authCallback() {
    return createUrl(this.baseUrl, "/api/v1/auth/providers/callback")
  }
  
  authClaimQuickConnectCode() {
    return createUrl(this.baseUrl, "/api/v1/auth/quick-connect/claim")
  }
  
  authFinishProvider() {
    return createUrl(this.baseUrl, "/api/v1/auth/providers/finish")
  }
  
  authFinishQuickConnect() {
    return createUrl(this.baseUrl, "/api/v1/auth/quick-connect/finish")
  }
  
  authGetProviderStatus() {
    return createUrl(this.baseUrl, "/api/v1/auth/provider/status")
  }
  
  authGetProviders() {
    return createUrl(this.baseUrl, "/api/v1/auth/providers")
  }
  
  authGetQuickConnectStatus() {
    return createUrl(this.baseUrl, "/api/v1/auth/quick-connect/status")
  }
  
  authProviderInitiate() {
    return createUrl(this.baseUrl, "/api/v1/auth/providers/initiate")
  }
  
  authQuickConnectInitiate() {
    return createUrl(this.baseUrl, "/api/v1/auth/quick-connect/initiate")
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

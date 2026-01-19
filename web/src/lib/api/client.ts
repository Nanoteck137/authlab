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
  
  authFinishQuickConnect(body: api.AuthFinishQuickConnectBody, options?: ExtraOptions) {
    return this.request("/api/v1/auth/quick-connect/finish", "POST", api.AuthFinishQuickConnect, z.any(), body, options)
  }
  
  authGetProviderStatus(body: api.AuthGetProviderStatusBody, options?: ExtraOptions) {
    return this.request("/api/v1/auth/provider/status", "POST", api.AuthGetProviderStatus, z.any(), body, options)
  }
  
  authGetQuickConnectStatus(code: string, options?: ExtraOptions) {
    return this.request(`/api/v1/auth/quick-connect/status/${code}`, "GET", api.AuthGetQuickConnectStatus, z.any(), undefined, options)
  }
  
  authInitiate(providerId: string, options?: ExtraOptions) {
    return this.request(`/api/v1/auth/initiate/${providerId}`, "POST", api.AuthInitiate, z.any(), undefined, options)
  }
  
  authLoginWithCode(body: api.AuthLoginWithCodeBody, options?: ExtraOptions) {
    return this.request("/api/v1/auth/login-with-code", "POST", api.AuthLoginWithCode, z.any(), body, options)
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
  
  getAuthCode(requestId: string, options?: ExtraOptions) {
    return this.request(`/api/v1/auth/code/${requestId}`, "GET", api.GetAuthCode, z.any(), undefined, options)
  }
  
  getAuthProviders(options?: ExtraOptions) {
    return this.request("/api/v1/auth/providers", "GET", api.GetAuthProviders, z.any(), undefined, options)
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
    return createUrl(this.baseUrl, "/api/v1/auth/callback")
  }
  
  authClaimQuickConnectCode() {
    return createUrl(this.baseUrl, "/api/v1/auth/quick-connect/claim")
  }
  
  authFinishQuickConnect() {
    return createUrl(this.baseUrl, "/api/v1/auth/quick-connect/finish")
  }
  
  authGetProviderStatus() {
    return createUrl(this.baseUrl, "/api/v1/auth/provider/status")
  }
  
  authGetQuickConnectStatus(code: string) {
    return createUrl(this.baseUrl, `/api/v1/auth/quick-connect/status/${code}`)
  }
  
  authInitiate(providerId: string) {
    return createUrl(this.baseUrl, `/api/v1/auth/initiate/${providerId}`)
  }
  
  authLoginWithCode() {
    return createUrl(this.baseUrl, "/api/v1/auth/login-with-code")
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
  
  getAuthCode(requestId: string) {
    return createUrl(this.baseUrl, `/api/v1/auth/code/${requestId}`)
  }
  
  getAuthProviders() {
    return createUrl(this.baseUrl, "/api/v1/auth/providers")
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

import { z } from "zod";
import * as api from "./types";
import { BaseApiClient, createUrl, type ExtraOptions } from "./base-client";


export class ApiClient extends BaseApiClient {
  url: ClientUrls;

  constructor(baseUrl: string) {
    super(baseUrl);
    this.url = new ClientUrls(baseUrl);
  }
  
  
  authCreateQuickCodeToken(body: api.GetAuthTokenFromQuickCodeBody, options?: ExtraOptions) {
    return this.request("/api/v1/auth/quick-code/token", "POST", api.GetAuthTokenFromQuickCode, z.any(), body, options)
  }
  
  authGetQuickCodeStatus(code: string, options?: ExtraOptions) {
    return this.request(`/api/v1/auth/quick-code/status/${code}`, "GET", api.AuthGetQuickCodeStatus, z.any(), undefined, options)
  }
  
  authInitiate(providerId: string, options?: ExtraOptions) {
    return this.request(`/api/v1/auth/initiate/${providerId}`, "POST", api.AuthInitiate, z.any(), undefined, options)
  }
  
  authInitiateQuick(options?: ExtraOptions) {
    return this.request("/api/v1/auth/initiate/quick", "POST", api.AuthInitiateQuick, z.any(), undefined, options)
  }
  
  authLoginQuickCode(body: api.AuthLoginQuickCodeBody, options?: ExtraOptions) {
    return this.request("/api/v1/auth/login-quick-code", "POST", z.undefined(), z.any(), body, options)
  }
  
  authLoginWithCode(body: api.AuthLoginWithCodeBody, options?: ExtraOptions) {
    return this.request("/api/v1/auth/login-with-code", "POST", api.AuthLoginWithCode, z.any(), body, options)
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
  
  authCreateQuickCodeToken() {
    return createUrl(this.baseUrl, "/api/v1/auth/quick-code/token")
  }
  
  authGetQuickCodeStatus(code: string) {
    return createUrl(this.baseUrl, `/api/v1/auth/quick-code/status/${code}`)
  }
  
  authInitiate(providerId: string) {
    return createUrl(this.baseUrl, `/api/v1/auth/initiate/${providerId}`)
  }
  
  authInitiateQuick() {
    return createUrl(this.baseUrl, "/api/v1/auth/initiate/quick")
  }
  
  authLoginQuickCode() {
    return createUrl(this.baseUrl, "/api/v1/auth/login-quick-code")
  }
  
  authLoginWithCode() {
    return createUrl(this.baseUrl, "/api/v1/auth/login-with-code")
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

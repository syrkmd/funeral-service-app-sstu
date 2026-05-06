import { apiClient, useMockApi } from "./client";
import { mockProxyApi } from "./mock/proxy.mock";
import type { IPAccessRule, IPAccessType, RateLimitSettings } from "../app/store/proxyStore";

export type ProxyConfig = {
  ipRules: IPAccessRule[];
  defaultPolicy: "allow" | "deny";
  rateLimitSettings: RateLimitSettings;
};

export type ProxyApi = {
  getConfig: () => Promise<ProxyConfig>;
  addIPRule: (rule: Omit<IPAccessRule, "id" | "addedAt">) => Promise<IPAccessRule>;
  removeIPRule: (id: string) => Promise<void>;
  setDefaultPolicy: (policy: "allow" | "deny") => Promise<ProxyConfig>;
  updateRateLimits: (settings: RateLimitSettings) => Promise<ProxyConfig>;
};

const realProxyApi: ProxyApi = {
  async getConfig() {
    const response = await apiClient.get<ProxyConfig>("/proxy/config");
    return response.data;
  },

  async addIPRule(rule) {
    const response = await apiClient.post<IPAccessRule>("/proxy/ip-rules", rule);
    return response.data;
  },

  async removeIPRule(id) {
    await apiClient.delete(`/proxy/ip-rules/${id}`);
  },

  async setDefaultPolicy(policy) {
    const response = await apiClient.patch<ProxyConfig>("/proxy/default-policy", {
      policy,
    });
    return response.data;
  },

  async updateRateLimits(settings) {
    const response = await apiClient.patch<ProxyConfig>("/proxy/rate-limits", settings);
    return response.data;
  },
};

const proxyApi = useMockApi ? mockProxyApi : realProxyApi;

export const getProxyConfig = proxyApi.getConfig;
export const addIPRule = proxyApi.addIPRule;
export const removeIPRule = proxyApi.removeIPRule;
export const setDefaultPolicy = proxyApi.setDefaultPolicy;
export const updateRateLimits = proxyApi.updateRateLimits;
export type { IPAccessType };

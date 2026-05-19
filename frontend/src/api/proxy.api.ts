import { proxyApiClient } from "./client";
import { fetchDashboardRateLimits, type DashboardIPRule } from "./metrics.api";
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

export type IPAccessDecision = {
  ip: string;
  allowed: boolean;
  decision: string;
  reason: string;
  verificationRequired?: boolean;
  matchedRuleId?: string;
  matchedValue?: string;
};

function mapBackendRuleType(type: DashboardIPRule["type"]): IPAccessType {
  if (type === "allowlist") return "allow";
  if (type === "denylist") return "deny";
  return "gray";
}

function mapFrontendRuleType(type: IPAccessType): DashboardIPRule["type"] {
  if (type === "allow") return "allowlist";
  if (type === "deny") return "denylist";
  return "graylist";
}

function mapBackendIPRule(rule: DashboardIPRule): IPAccessRule {
  return {
    id: rule.id,
    ip: rule.value,
    type: mapBackendRuleType(rule.type),
    note: rule.description,
    addedAt: "-",
  };
}

const realProxyApi: ProxyApi = {
  async getConfig() {
    const [ipRulesResponse, rateLimits] = await Promise.all([
      proxyApiClient.get<DashboardIPRule[]>("/api/ip_access/lists"),
      fetchDashboardRateLimits(),
    ]);

    const primaryRule = rateLimits.activeRateLimitRules[0];

    return {
      ipRules: (ipRulesResponse.data || []).map(mapBackendIPRule),
      defaultPolicy: "allow",
      rateLimitSettings: {
        rpsLimit: primaryRule?.rps || 0,
        rpmLimit: primaryRule?.rpm || 0,
      },
    };
  },

  async addIPRule(rule) {
    const response = await proxyApiClient.post<DashboardIPRule>("/api/ip_access/lists", {
      type: mapFrontendRuleType(rule.type),
      value: rule.ip,
      description: rule.note,
    });
    return mapBackendIPRule(response.data);
  },

  async removeIPRule(id) {
    await proxyApiClient.delete(`/api/ip_access/lists/${id}`);
  },

  async setDefaultPolicy(policy) {
    const config = await realProxyApi.getConfig();
    return { ...config, defaultPolicy: policy };
  },

  async updateRateLimits(settings) {
    const config = await realProxyApi.getConfig();
    return { ...config, rateLimitSettings: settings };
  },
};

const proxyApi = realProxyApi;

export const getProxyConfig = proxyApi.getConfig;
export const addIPRule = proxyApi.addIPRule;
export const removeIPRule = proxyApi.removeIPRule;
export const setDefaultPolicy = proxyApi.setDefaultPolicy;
export const updateRateLimits = proxyApi.updateRateLimits;
export async function checkIPAccess(ip: string): Promise<IPAccessDecision> {
  const response = await proxyApiClient.get("/api/ip_access/check", {
    params: { ip },
  });
  const data = response.data || {};

  return {
    ip: data.ip || "",
    allowed: Boolean(data.allowed),
    decision: data.decision || "",
    reason: data.reason || "",
    verificationRequired: Boolean(data.verification_required ?? data.verificationRequired),
    matchedRuleId: data.matched_rule_id ?? data.matchedRuleId,
    matchedValue: data.matched_value ?? data.matchedValue,
  };
}
export type { IPAccessType };

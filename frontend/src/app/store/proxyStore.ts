import { create } from "zustand";
import {
  addIPRule as addIPRuleRequest,
  checkIPAccess as checkIPAccessRequest,
  getProxyConfig,
  removeIPRule as removeIPRuleRequest,
  setDefaultPolicy as setDefaultPolicyRequest,
  updateRateLimits as updateRateLimitsRequest,
} from "../../api/proxy.api";

export type IPAccessType = "allow" | "deny" | "gray";

export type IPAccessRule = {
  id: string;
  ip: string;
  type: IPAccessType;
  note?: string;
  addedAt: string;
};

export type RateLimitSettings = {
  rpsLimit: number;
  rpmLimit: number;
};

type ProxyStore = {
  ipRules: IPAccessRule[];
  defaultPolicy: "allow" | "deny";
  rateLimitSettings: RateLimitSettings;
  isLoading: boolean;
  error: string | null;
  loadConfig: () => Promise<void>;
  addIPRule: (rule: Omit<IPAccessRule, "id" | "addedAt">) => Promise<void>;
  removeIPRule: (id: string) => Promise<void>;
  setDefaultPolicy: (policy: "allow" | "deny") => Promise<void>;
  updateRateLimits: (settings: RateLimitSettings) => Promise<void>;
  checkIPAccess: (ip: string) => Promise<"allowed" | "denied" | "captcha">;
};

export const useProxyStore = create<ProxyStore>()((set) => ({
  ipRules: [],
  defaultPolicy: "allow",
  rateLimitSettings: {
    rpsLimit: 100,
    rpmLimit: 5000,
  },
  isLoading: false,
  error: null,

  loadConfig: async () => {
    set({ isLoading: true, error: null });

    try {
      const config = await getProxyConfig();
      set({ ...config, isLoading: false });
    } catch (error) {
      set({
        isLoading: false,
        error: error instanceof Error ? error.message : "Failed to load proxy config",
      });
    }
  },

  addIPRule: async (rule) => {
    const newRule = await addIPRuleRequest(rule);
    set((state) => ({
      ipRules: [...state.ipRules, newRule],
    }));
  },

  removeIPRule: async (id) => {
    await removeIPRuleRequest(id);
    set((state) => ({
      ipRules: state.ipRules.filter((rule) => rule.id !== id),
    }));
  },

  setDefaultPolicy: async (policy) => {
    const config = await setDefaultPolicyRequest(policy);
    set(config);
  },

  updateRateLimits: async (settings) => {
    const config = await updateRateLimitsRequest(settings);
    set(config);
  },

  checkIPAccess: async (ip) => {
    const decision = await checkIPAccessRequest(ip);

    if (decision.verificationRequired || decision.decision === "gray") {
      return "captcha";
    }

    return decision.allowed ? "allowed" : "denied";
  },
}));

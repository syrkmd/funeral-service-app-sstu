import type { ProxyApi, ProxyConfig } from "../proxy.api";
import type { IPAccessRule } from "../../app/store/proxyStore";

const STORAGE_KEY = "proxy-mock-config";
const LEGACY_STORAGE_KEY = "proxy-storage";

const initialConfig: ProxyConfig = {
  ipRules: [
    {
      id: "1",
      ip: "192.168.1.100",
      type: "allow",
      note: "Office network",
      addedAt: "2026-04-15",
    },
    {
      id: "2",
      ip: "10.0.0.0/8",
      type: "deny",
      note: "Blocked range",
      addedAt: "2026-04-10",
    },
    {
      id: "3",
      ip: "172.16.0.1",
      type: "gray",
      note: "Requires verification",
      addedAt: "2026-04-20",
    },
  ],
  defaultPolicy: "allow",
  rateLimitSettings: {
    rpsLimit: 100,
    rpmLimit: 5000,
  },
};

function readConfig(): ProxyConfig {
  if (typeof window === "undefined") {
    return initialConfig;
  }

  const storedConfig = window.localStorage.getItem(STORAGE_KEY);

  if (!storedConfig) {
    const legacyConfig = readLegacyConfig();

    if (legacyConfig) {
      writeConfig(legacyConfig);
      return legacyConfig;
    }

    writeConfig(initialConfig);
    return initialConfig;
  }

  try {
    return JSON.parse(storedConfig);
  } catch {
    writeConfig(initialConfig);
    return initialConfig;
  }
}

function readLegacyConfig(): ProxyConfig | null {
  if (typeof window === "undefined") {
    return null;
  }

  const storedConfig = window.localStorage.getItem(LEGACY_STORAGE_KEY);

  if (!storedConfig) {
    return null;
  }

  try {
    const parsed = JSON.parse(storedConfig);
    const state = parsed?.state;

    if (!state?.ipRules || !state?.defaultPolicy || !state?.rateLimitSettings) {
      return null;
    }

    return {
      ipRules: state.ipRules,
      defaultPolicy: state.defaultPolicy,
      rateLimitSettings: state.rateLimitSettings,
    };
  } catch {
    return null;
  }
}

function writeConfig(config: ProxyConfig) {
  if (typeof window !== "undefined") {
    window.localStorage.setItem(STORAGE_KEY, JSON.stringify(config));
  }
}

export const mockProxyApi: ProxyApi = {
  async getConfig() {
    return readConfig();
  },

  async addIPRule(rule) {
    const config = readConfig();
    const newRule: IPAccessRule = {
      ...rule,
      id: Date.now().toString(),
      addedAt: new Date().toISOString().split("T")[0],
    };

    writeConfig({
      ...config,
      ipRules: [...config.ipRules, newRule],
    });

    return newRule;
  },

  async removeIPRule(id) {
    const config = readConfig();
    writeConfig({
      ...config,
      ipRules: config.ipRules.filter((rule) => rule.id !== id),
    });
  },

  async setDefaultPolicy(policy) {
    const config = {
      ...readConfig(),
      defaultPolicy: policy,
    };

    writeConfig(config);
    return config;
  },

  async updateRateLimits(settings) {
    const config = {
      ...readConfig(),
      rateLimitSettings: settings,
    };

    writeConfig(config);
    return config;
  },
};

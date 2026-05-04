import { create } from 'zustand';
import { persist } from 'zustand/middleware';

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
  addIPRule: (rule: Omit<IPAccessRule, "id" | "addedAt">) => void;
  removeIPRule: (id: string) => void;
  setDefaultPolicy: (policy: "allow" | "deny") => void;
  updateRateLimits: (settings: RateLimitSettings) => void;
  checkIPAccess: (ip: string) => "allowed" | "denied" | "captcha";
};

export const useProxyStore = create<ProxyStore>()(
  persist(
    (set, get) => ({
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
  addIPRule: (rule) =>
    set((state) => ({
      ipRules: [
        ...state.ipRules,
        {
          ...rule,
          id: Date.now().toString(),
          addedAt: new Date().toISOString().split("T")[0],
        },
      ],
    })),
  removeIPRule: (id) =>
    set((state) => ({
      ipRules: state.ipRules.filter((rule) => rule.id !== id),
    })),
  setDefaultPolicy: (policy) => set({ defaultPolicy: policy }),
  updateRateLimits: (settings) => set({ rateLimitSettings: settings }),
  checkIPAccess: (ip) => {
    const state = get();
    const rule = state.ipRules.find((r) => r.ip === ip);

    if (rule) {
      if (rule.type === "allow") return "allowed";
      if (rule.type === "deny") return "denied";
      if (rule.type === "gray") return "captcha";
    }

    return state.defaultPolicy === "allow" ? "allowed" : "denied";
  },
    }),
    {
      name: 'proxy-storage',
    }
  )
);

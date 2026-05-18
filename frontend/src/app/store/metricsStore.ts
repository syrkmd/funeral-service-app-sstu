import { create } from 'zustand';
import type { ProxyDashboardMetrics } from '../../api/metrics.api';

export type RPSDataPoint = {
  time: string;
  value: number;
};

type MetricsStore = {
  rpsData: RPSDataPoint[];
  totalRequests: number;
  errors: number;
  activeClients: number;
  ordersToday: number;
  avgLatency: number;
  traffic: number;
  cacheHits: number;
  cacheMisses: number;
  cacheStores: number;
  rateLimitViolations: number;
  blockedRequests: number;
  upstreamStatus: string;

  // Actions
  addRPSDataPoint: (dataPoint: RPSDataPoint) => void;
  incrementTotalRequests: (count?: number) => void;
  incrementErrors: (count?: number) => void;
  updateActiveClients: (count: number) => void;
  updateMetrics: (metrics: Partial<MetricsStore>) => void;
  updateProxyMetrics: (metrics: ProxyDashboardMetrics) => void;
};

export const useMetricsStore = create<MetricsStore>()((set) => ({
      rpsData: [],
      totalRequests: 0,
      errors: 0,
      activeClients: 0,
      ordersToday: 0,
      avgLatency: 0,
      traffic: 0,
      cacheHits: 0,
      cacheMisses: 0,
      cacheStores: 0,
      rateLimitViolations: 0,
      blockedRequests: 0,
      upstreamStatus: "unknown",

      addRPSDataPoint: (dataPoint) =>
        set((state) => ({
          rpsData: [...state.rpsData.slice(-11), dataPoint],
        })),

      incrementTotalRequests: (count = 1) =>
        set((state) => ({
          totalRequests: state.totalRequests + count,
        })),

      incrementErrors: (count = 1) =>
        set((state) => ({
          errors: state.errors + count,
        })),

      updateActiveClients: (count) =>
        set({ activeClients: count }),

      updateMetrics: (metrics) =>
        set((state) => ({
          ...state,
          ...metrics,
        })),

      updateProxyMetrics: (metrics) =>
        set((state) => ({
          ...state,
          ...metrics,
        })),
}));

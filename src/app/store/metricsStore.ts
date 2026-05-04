import { create } from 'zustand';
import { persist } from 'zustand/middleware';

export type RPSDataPoint = {
  time: string;
  value: number;
};

export type ActivityLogEntry = {
  id: string;
  type: 'ORDER_CREATED' | 'STATUS_CHANGED' | 'PAYMENT_CONFIRMED';
  user: string;
  orderId?: string;
  time: string;
  timestamp: number;
};

type MetricsStore = {
  rpsData: RPSDataPoint[];
  activityLog: ActivityLogEntry[];
  totalRequests: number;
  errors: number;
  activeClients: number;
  ordersToday: number;
  avgLatency: number;
  traffic: number;

  // Actions
  addRPSDataPoint: (dataPoint: RPSDataPoint) => void;
  addActivityLogEntry: (entry: Omit<ActivityLogEntry, 'id' | 'timestamp'>) => void;
  incrementTotalRequests: (count?: number) => void;
  incrementErrors: (count?: number) => void;
  updateActiveClients: (count: number) => void;
  incrementOrdersToday: () => void;
  updateMetrics: (metrics: Partial<MetricsStore>) => void;
};

export const useMetricsStore = create<MetricsStore>()(
  persist(
    (set) => ({
      rpsData: generateInitialRPSData(),
      activityLog: [],
      totalRequests: 45231,
      errors: 23,
      activeClients: 1284,
      ordersToday: 16,
      avgLatency: 24,
      traffic: 1.2,

      addRPSDataPoint: (dataPoint) =>
        set((state) => ({
          rpsData: [...state.rpsData.slice(-11), dataPoint],
        })),

      addActivityLogEntry: (entry) =>
        set((state) => ({
          activityLog: [
            {
              ...entry,
              id: Date.now().toString() + Math.random(),
              timestamp: Date.now(),
            },
            ...state.activityLog,
          ].slice(0, 10), // Keep only last 10 entries
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

      incrementOrdersToday: () =>
        set((state) => ({
          ordersToday: state.ordersToday + 1,
        })),

      updateMetrics: (metrics) =>
        set((state) => ({
          ...state,
          ...metrics,
        })),
    }),
    {
      name: 'metrics-storage',
    }
  )
);

// Helper function to generate initial RPS data
function generateInitialRPSData(): RPSDataPoint[] {
  const now = new Date();
  const data: RPSDataPoint[] = [];

  for (let i = 11; i >= 0; i--) {
    const time = new Date(now.getTime() - i * 5000); // 5 seconds apart
    data.push({
      time: time.toLocaleTimeString('ru-RU', { hour: '2-digit', minute: '2-digit', second: '2-digit' }),
      value: Math.floor(Math.random() * 40) + 30, // Random value between 30-70
    });
  }

  return data;
}

// Mock API functions (ready for real backend integration)
export async function fetchMetrics() {
  // Simulate API delay
  await new Promise(resolve => setTimeout(resolve, 100));

  return {
    totalRequests: Math.floor(Math.random() * 1000) + 45000,
    errors: Math.floor(Math.random() * 10) + 20,
    activeClients: Math.floor(Math.random() * 200) + 1200,
    avgLatency: Math.floor(Math.random() * 10) + 20,
    traffic: parseFloat((Math.random() * 0.5 + 1.0).toFixed(2)),
  };
}

export async function fetchRPSData() {
  // Simulate API delay
  await new Promise(resolve => setTimeout(resolve, 50));

  const now = new Date();
  return {
    time: now.toLocaleTimeString('ru-RU', { hour: '2-digit', minute: '2-digit', second: '2-digit' }),
    value: Math.floor(Math.random() * 40) + 30,
  };
}

export async function fetchActivityLog() {
  // Simulate API delay
  await new Promise(resolve => setTimeout(resolve, 100));

  // In real implementation, this would fetch from backend
  return [];
}

import { create } from 'zustand';

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

  // Actions
  addRPSDataPoint: (dataPoint: RPSDataPoint) => void;
  incrementTotalRequests: (count?: number) => void;
  incrementErrors: (count?: number) => void;
  updateActiveClients: (count: number) => void;
  updateMetrics: (metrics: Partial<MetricsStore>) => void;
};

export const useMetricsStore = create<MetricsStore>()((set) => ({
      rpsData: generateInitialRPSData(),
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
}));

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

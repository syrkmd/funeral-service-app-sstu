import { create } from 'zustand';
import { persist } from 'zustand/middleware';

type AdminAuthStore = {
  isAdminAuth: boolean;
  setAdminAuth: (value: boolean) => void;
  clearAdminAuth: () => void;
};

export const useAdminAuthStore = create<AdminAuthStore>()(
  persist(
    (set) => ({
      isAdminAuth: false,
      setAdminAuth: (value) => set({ isAdminAuth: value }),
      clearAdminAuth: () => set({ isAdminAuth: false }),
    }),
    {
      name: 'admin-auth-storage',
    }
  )
);

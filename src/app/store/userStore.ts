import { create } from 'zustand';
import { persist } from 'zustand/middleware';

type UserStore = {
  currentUserPhone: string | null;
  setUserPhone: (phone: string) => void;
  clearUserPhone: () => void;
};

export const useUserStore = create<UserStore>()(
  persist(
    (set) => ({
      currentUserPhone: null,
      setUserPhone: (phone) => set({ currentUserPhone: phone }),
      clearUserPhone: () => set({ currentUserPhone: null }),
    }),
    {
      name: 'user-storage',
    }
  )
);

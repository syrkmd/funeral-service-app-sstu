import { create } from "zustand";
import { getAdminSession, loginAdmin, logoutAdmin } from "../../api/users.api";

type AdminAuthStore = {
  isAdminAuth: boolean;
  isLoading: boolean;
  error: string | null;
  restoreAdminAuth: () => Promise<boolean>;
  setAdminAuth: (login: string, password: string) => Promise<boolean>;
  clearAdminAuth: () => Promise<void>;
};

export const useAdminAuthStore = create<AdminAuthStore>()((set) => ({
  isAdminAuth: false,
  isLoading: false,
  error: null,

  restoreAdminAuth: async () => {
    set({ isLoading: true, error: null });

    try {
      const session = await getAdminSession();
      set({ isAdminAuth: session.isAuthenticated, isLoading: false });
      return session.isAuthenticated;
    } catch (error) {
      set({
        isAdminAuth: false,
        isLoading: false,
        error: error instanceof Error ? error.message : "Failed to restore admin session",
      });
      return false;
    }
  },

  setAdminAuth: async (login, password) => {
    set({ isLoading: true, error: null });

    try {
      const session = await loginAdmin(login, password);
      set({ isAdminAuth: session.isAuthenticated, isLoading: false });
      return session.isAuthenticated;
    } catch (error) {
      set({
        isAdminAuth: false,
        isLoading: false,
        error: error instanceof Error ? error.message : "Failed to login",
      });
      return false;
    }
  },

  clearAdminAuth: async () => {
    await logoutAdmin();
    set({ isAdminAuth: false, error: null });
  },
}));

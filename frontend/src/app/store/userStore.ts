import { create } from "zustand";
import {
  getAccountSession,
  logoutAccount,
  verifyAccountCode,
} from "../../api/users.api";

type UserStore = {
  currentUserPhone: string | null;
  isAuthenticated: boolean;
  isLoading: boolean;
  error: string | null;
  restoreSession: () => Promise<boolean>;
  verifyCode: (phone: string, code: string) => Promise<boolean>;
  clearUserPhone: () => Promise<void>;
};

export const useUserStore = create<UserStore>()((set) => ({
  currentUserPhone: null,
  isAuthenticated: false,
  isLoading: false,
  error: null,

  restoreSession: async () => {
    set({ isLoading: true, error: null });

    try {
      const session = await getAccountSession();
      set({
        currentUserPhone: session.phone,
        isAuthenticated: session.isAuthenticated,
        isLoading: false,
      });
      return session.isAuthenticated;
    } catch (error) {
      set({
        currentUserPhone: null,
        isAuthenticated: false,
        isLoading: false,
        error: error instanceof Error ? error.message : "Failed to restore session",
      });
      return false;
    }
  },

  verifyCode: async (phone, code) => {
    set({ isLoading: true, error: null });

    try {
      const session = await verifyAccountCode(phone, code);
      set({
        currentUserPhone: session.phone,
        isAuthenticated: session.isAuthenticated,
        isLoading: false,
      });
      return session.isAuthenticated;
    } catch (error) {
      set({
        isLoading: false,
        error: error instanceof Error ? error.message : "Failed to verify account",
      });
      return false;
    }
  },

  clearUserPhone: async () => {
    await logoutAccount();
    set({
      currentUserPhone: null,
      isAuthenticated: false,
      error: null,
    });
  },
}));

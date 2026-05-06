import type { AccountSession, AdminSession, UsersApi } from "../users.api";

const ACCOUNT_SESSION_KEY = "account-session";
const ADMIN_SESSION_KEY = "admin-session";
const LEGACY_ACCOUNT_AUTH_KEY = "account_authenticated";
const LEGACY_ACCOUNT_PHONE_KEY = "account_phone";
const LEGACY_USER_STORE_KEY = "user-storage";
const LEGACY_ADMIN_STORE_KEY = "admin-auth-storage";
const DEMO_ACCOUNT_CODE = "4821";

function readAccountSession(): AccountSession {
  if (typeof window === "undefined") {
    return { isAuthenticated: false, phone: null };
  }

  const session = window.localStorage.getItem(ACCOUNT_SESSION_KEY);

  if (session) {
    return JSON.parse(session);
  }

  const legacySession = readLegacyAccountSession();

  if (legacySession.isAuthenticated) {
    writeAccountSession(legacySession);
  }

  return legacySession;
}

function writeAccountSession(session: AccountSession) {
  if (typeof window !== "undefined") {
    window.localStorage.setItem(ACCOUNT_SESSION_KEY, JSON.stringify(session));
  }
}

function clearAccountSession() {
  if (typeof window !== "undefined") {
    window.localStorage.removeItem(ACCOUNT_SESSION_KEY);
    window.localStorage.removeItem(LEGACY_ACCOUNT_AUTH_KEY);
    window.localStorage.removeItem(LEGACY_ACCOUNT_PHONE_KEY);
    window.localStorage.removeItem(LEGACY_USER_STORE_KEY);
  }
}

function readLegacyAccountSession(): AccountSession {
  if (typeof window === "undefined") {
    return { isAuthenticated: false, phone: null };
  }

  const isAuthenticated = window.localStorage.getItem(LEGACY_ACCOUNT_AUTH_KEY);
  const phone = window.localStorage.getItem(LEGACY_ACCOUNT_PHONE_KEY);

  if (isAuthenticated && phone) {
    return { isAuthenticated: true, phone };
  }

  const storedUser = window.localStorage.getItem(LEGACY_USER_STORE_KEY);

  if (!storedUser) {
    return { isAuthenticated: false, phone: null };
  }

  try {
    const parsed = JSON.parse(storedUser);
    const currentUserPhone = parsed?.state?.currentUserPhone;
    return currentUserPhone
      ? { isAuthenticated: true, phone: currentUserPhone }
      : { isAuthenticated: false, phone: null };
  } catch {
    return { isAuthenticated: false, phone: null };
  }
}

function readAdminSession(): AdminSession {
  if (typeof window === "undefined") {
    return { isAuthenticated: false };
  }

  const session = window.localStorage.getItem(ADMIN_SESSION_KEY);

  if (session) {
    return JSON.parse(session);
  }

  const legacySession = readLegacyAdminSession();

  if (legacySession.isAuthenticated) {
    writeAdminSession(legacySession);
  }

  return legacySession;
}

function writeAdminSession(session: AdminSession) {
  if (typeof window !== "undefined") {
    window.localStorage.setItem(ADMIN_SESSION_KEY, JSON.stringify(session));
  }
}

function clearAdminSession() {
  if (typeof window !== "undefined") {
    window.localStorage.removeItem(ADMIN_SESSION_KEY);
    window.localStorage.removeItem(LEGACY_ADMIN_STORE_KEY);
  }
}

function readLegacyAdminSession(): AdminSession {
  if (typeof window === "undefined") {
    return { isAuthenticated: false };
  }

  const storedAdmin = window.localStorage.getItem(LEGACY_ADMIN_STORE_KEY);

  if (!storedAdmin) {
    return { isAuthenticated: false };
  }

  try {
    const parsed = JSON.parse(storedAdmin);
    return { isAuthenticated: Boolean(parsed?.state?.isAdminAuth) };
  } catch {
    return { isAuthenticated: false };
  }
}

export const mockUsersApi: UsersApi = {
  async verifyAccountCode(phone, code) {
    if (code !== DEMO_ACCOUNT_CODE) {
      throw new Error("Неверный код");
    }

    const session = {
      isAuthenticated: true,
      phone,
    };

    writeAccountSession(session);
    return session;
  },

  async getAccountSession() {
    return readAccountSession();
  },

  async logoutAccount() {
    clearAccountSession();
  },

  async loginAdmin(login, password) {
    if (!login || !password) {
      throw new Error("Введите логин и пароль");
    }

    const session = { isAuthenticated: true };
    writeAdminSession(session);
    return session;
  },

  async getAdminSession() {
    return readAdminSession();
  },

  async logoutAdmin() {
    clearAdminSession();
  },
};

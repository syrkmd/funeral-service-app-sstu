import { apiClient, useMockApi } from "./client";
import { mockUsersApi } from "./mock/users.mock";

export type AccountSession = {
  isAuthenticated: boolean;
  phone: string | null;
};

export type AdminSession = {
  isAuthenticated: boolean;
};

export type UsersApi = {
  verifyAccountCode: (phone: string, code: string) => Promise<AccountSession>;
  getAccountSession: () => Promise<AccountSession>;
  logoutAccount: () => Promise<void>;
  loginAdmin: (login: string, password: string) => Promise<AdminSession>;
  getAdminSession: () => Promise<AdminSession>;
  logoutAdmin: () => Promise<void>;
};

const realUsersApi: UsersApi = {
  async verifyAccountCode(phone, code) {
    const response = await apiClient.post<AccountSession>("/users/account/verify", {
      phone,
      code,
    });
    return response.data;
  },

  async getAccountSession() {
    const response = await apiClient.get<AccountSession>("/users/account/session");
    return response.data;
  },

  async logoutAccount() {
    await apiClient.post("/users/account/logout");
  },

  async loginAdmin(login, password) {
    const response = await apiClient.post<AdminSession>("/admin/session", {
      login,
      password,
    });
    return response.data;
  },

  async getAdminSession() {
    const response = await apiClient.get<AdminSession>("/admin/session");
    return response.data;
  },

  async logoutAdmin() {
    await apiClient.delete("/admin/session");
  },
};

const usersApi = useMockApi ? mockUsersApi : realUsersApi;

export const verifyAccountCode = usersApi.verifyAccountCode;
export const getAccountSession = usersApi.getAccountSession;
export const logoutAccount = usersApi.logoutAccount;
export const loginAdmin = usersApi.loginAdmin;
export const getAdminSession = usersApi.getAdminSession;
export const logoutAdmin = usersApi.logoutAdmin;

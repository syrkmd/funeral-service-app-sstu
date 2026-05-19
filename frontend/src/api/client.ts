import axios from "axios";

const baseURL = import.meta.env.VITE_API_URL || "";
const proxyBaseURL = import.meta.env.VITE_PROXY_API_URL || "";

export const apiClient = axios.create({
  baseURL,
  withCredentials: true,
  headers: {
    "Content-Type": "application/json",
    Accept: "application/json",
  },
});

export const proxyApiClient = axios.create({
  baseURL: proxyBaseURL,
  headers: {
    "Content-Type": "application/json",
    Accept: "application/json",
  },
});

function handleApiError(label: string) {
  return (error: any) => {
    const requestBaseURL = error.config?.baseURL || "";
    console.error(`[${label}] Request failed`, {
      baseURL: requestBaseURL,
      url: error.config?.url,
      method: error.config?.method,
      status: error.response?.status,
      data: error.response?.data,
      message: error.message,
    });

    const message =
      error.response?.data?.message ||
      error.response?.statusText ||
      error.message ||
      "API request failed";

    return Promise.reject(new Error(message));
  };
}

apiClient.interceptors.response.use(
  (response) => response,
  handleApiError("API"),
);

proxyApiClient.interceptors.response.use(
  (response) => response,
  handleApiError("Proxy API"),
);

export const useMockApi = import.meta.env.VITE_USE_MOCK === "true";
export const useMockAuth = import.meta.env.VITE_USE_MOCK_AUTH === "true";

console.info("[API] Configuration", {
  baseURL,
  proxyBaseURL,
  useMockApi,
  useMockAuth,
});

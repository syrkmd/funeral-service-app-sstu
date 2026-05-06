import { apiClient, useMockApi } from "./client";
import { mockOrdersApi } from "./mock/orders.mock";
import type { Document, Order, OrderStatus } from "../app/store/ordersStore";

export type CreateOrderData = Omit<Order, "id">;

export type OrdersApi = {
  getOrders: (updatedAfter?: string | null) => Promise<Order[]>;
  getOrderById: (id: string) => Promise<Order | undefined>;
  getOrdersByPhone: (phone: string) => Promise<Order[]>;
  createOrder: (data: CreateOrderData) => Promise<Order>;
  updateOrderStatus: (id: string, status: OrderStatus) => Promise<Order>;
  updatePayment: (id: string, isPaid: boolean) => Promise<Order>;
  addDocument: (orderId: string, document: Document) => Promise<Order>;
  removeDocument: (orderId: string, documentId: number) => Promise<Order>;
};

const realOrdersApi: OrdersApi = {
  async getOrders(updatedAfter) {
    const response = await apiClient.get<Order[]>("/orders", {
      params: updatedAfter ? { updatedAfter } : undefined,
    });
    return response.data;
  },

  async getOrderById(id) {
    const response = await apiClient.get<Order>(`/orders/${id}`);
    return response.data;
  },

  async getOrdersByPhone(phone) {
    const response = await apiClient.get<Order[]>("/orders", {
      params: { phone },
    });
    return response.data;
  },

  async createOrder(data) {
    const response = await apiClient.post<Order>("/orders", data);
    return response.data;
  },

  async updateOrderStatus(id, status) {
    const response = await apiClient.patch<Order>(`/orders/${id}/status`, { status });
    return response.data;
  },

  async updatePayment(id, isPaid) {
    const response = await apiClient.patch<Order>(`/orders/${id}/payment`, { isPaid });
    return response.data;
  },

  async addDocument(orderId, document) {
    const response = await apiClient.post<Order>(`/orders/${orderId}/documents`, document);
    return response.data;
  },

  async removeDocument(orderId, documentId) {
    const response = await apiClient.delete<Order>(`/orders/${orderId}/documents/${documentId}`);
    return response.data;
  },
};

const ordersApi = useMockApi ? mockOrdersApi : realOrdersApi;

export const getOrders = ordersApi.getOrders;
export const getOrderById = ordersApi.getOrderById;
export const getOrdersByPhone = ordersApi.getOrdersByPhone;
export const createOrder = ordersApi.createOrder;
export const updateOrderStatus = ordersApi.updateOrderStatus;
export const updatePayment = ordersApi.updatePayment;
export const addDocument = ordersApi.addDocument;
export const removeDocument = ordersApi.removeDocument;

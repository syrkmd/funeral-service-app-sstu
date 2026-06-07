import { apiClient, useMockApi } from "./client";
import { mockOrdersApi } from "./mock/orders.mock";
import type { Document, Order, OrderStatus } from "../app/store/ordersStore";

export type CreateOrderData = Pick<
  Order,
  | "date"
  | "client"
  | "deceased"
  | "serviceDate"
  | "serviceTime"
  | "serviceAddress"
  | "cemetery"
  | "cemeteryNotes"
  | "cemeteryPlotId"
  | "cemeteryPlotCode"
  | "services"
  | "products"
  | "documents"
>;
export type UpdateOrderClientData = Order["client"];
export type UpdateOrderDeceasedData = Order["deceased"];
export type UpdateOrderCeremonyData = {
  serviceDate: string;
  serviceTime: string;
  serviceAddress: string;
  cemetery: string;
  cemeteryNotes?: string | null;
};
export type PayOrderData = {
  cardNumber: string;
  cvv: string;
  expiryDate: string;
};

export type OrdersApi = {
  getOrders: (updatedAfter?: string | null) => Promise<Order[]>;
  getOrderById: (id: string) => Promise<Order | undefined>;
  getOrdersByPhone: (phone: string) => Promise<Order[]>;
  createOrder: (data: CreateOrderData) => Promise<Order>;
  updateOrderStatus: (id: string, status: OrderStatus) => Promise<Order>;
  updatePayment: (id: string, isPaid: boolean) => Promise<Order>;
  payOrder: (id: string, data: PayOrderData) => Promise<Order>;
  updateClient: (id: string, data: UpdateOrderClientData) => Promise<Order>;
  updateDeceased: (id: string, data: UpdateOrderDeceasedData) => Promise<Order>;
  updateOrderDate: (id: string, date: string) => Promise<Order>;
  updateCeremony: (id: string, data: UpdateOrderCeremonyData) => Promise<Order>;
  replaceServices: (id: string, serviceIds: number[]) => Promise<Order>;
  replaceProducts: (id: string, productIds: number[]) => Promise<Order>;
  applyDiscount: (id: string, discountAmount: number, reason?: string) => Promise<Order>;
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

  async payOrder(id, data) {
    const response = await apiClient.post<Order>(`/orders/${id}/pay`, data);
    return response.data;
  },

  async updateClient(id, data) {
    const response = await apiClient.patch<Order>(`/orders/${id}/client`, data);
    return response.data;
  },

  async updateDeceased(id, data) {
    const response = await apiClient.patch<Order>(`/orders/${id}/deceased`, data);
    return response.data;
  },

  async updateOrderDate(id, date) {
    const response = await apiClient.patch<Order>(`/orders/${id}/date`, { date });
    return response.data;
  },

  async updateCeremony(id, data) {
    const response = await apiClient.patch<Order>(`/orders/${id}/ceremony`, data);
    return response.data;
  },

  async replaceServices(id, serviceIds) {
    const response = await apiClient.put<Order>(`/orders/${id}/services`, { serviceIds });
    return response.data;
  },

  async replaceProducts(id, productIds) {
    const response = await apiClient.put<Order>(`/orders/${id}/products`, { productIds });
    return response.data;
  },

  async applyDiscount(id, discountAmount, reason) {
    const response = await apiClient.patch<Order>(`/orders/${id}/discount`, { discountAmount, reason });
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
export const payOrder = ordersApi.payOrder;
export const updateClient = ordersApi.updateClient;
export const updateDeceased = ordersApi.updateDeceased;
export const updateOrderDate = ordersApi.updateOrderDate;
export const updateCeremony = ordersApi.updateCeremony;
export const replaceServices = ordersApi.replaceServices;
export const replaceProducts = ordersApi.replaceProducts;
export const applyDiscount = ordersApi.applyDiscount;
export const addDocument = ordersApi.addDocument;
export const removeDocument = ordersApi.removeDocument;

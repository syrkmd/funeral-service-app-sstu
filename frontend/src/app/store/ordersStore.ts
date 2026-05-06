import { create } from "zustand";
import {
  addDocument as addOrderDocument,
  createOrder as createOrderRequest,
  getOrders,
  removeDocument as removeOrderDocument,
  updateOrderStatus as updateOrderStatusRequest,
  updatePayment,
  type CreateOrderData,
} from "../../api/orders.api";

export type OrderStatus = "processing" | "confirmed" | "completed" | "cancelled";

export type Document = {
  id: number;
  name: string;
  type: string;
  date: string;
  size: string;
};

export type Service = {
  name: string;
  price: number;
};

export type Product = {
  name: string;
  price: number;
};

export type Order = {
  id: string;
  date: string;
  createdAt?: string;
  updatedAt?: string;
  statusUpdatedAt?: string;
  paymentConfirmedAt?: string;
  status: OrderStatus;
  total: number;
  phone: string;
  isPaid: boolean;
  client: {
    name: string;
    phone: string;
    email: string;
  };
  deceased: {
    name: string;
    dateOfBirth: string;
    dateOfDeath: string;
  };
  services: Service[];
  products: Product[];
  documents: Document[];
};

type OrdersStore = {
  orders: Order[];
  lastUpdate: string | null;
  isLoading: boolean;
  error: string | null;
  loadOrders: (updatedAfter?: string | null) => Promise<void>;
  getOrderById: (id: string) => Order | undefined;
  createOrder: (data: CreateOrderData) => Promise<Order>;
  updateOrderStatus: (id: string, status: OrderStatus) => Promise<void>;
  updateOrderPayment: (id: string, isPaid: boolean) => Promise<void>;
  addDocument: (orderId: string, document: Document) => Promise<void>;
  removeDocument: (orderId: string, documentId: number) => Promise<void>;
};

function mergeOrders(currentOrders: Order[], incomingOrders: Order[]) {
  const ordersById = new Map(currentOrders.map((order) => [order.id, order]));

  incomingOrders.forEach((order) => {
    ordersById.set(order.id, order);
  });

  return Array.from(ordersById.values());
}

export const useOrdersStore = create<OrdersStore>()((set, get) => ({
  orders: [],
  lastUpdate: null,
  isLoading: false,
  error: null,

  loadOrders: async (updatedAfter) => {
    set({ isLoading: true, error: null });

    try {
      const orders = await getOrders(updatedAfter);
      set((state) => ({
        orders: updatedAfter ? mergeOrders(state.orders, orders) : orders,
        lastUpdate: new Date().toISOString(),
        isLoading: false,
      }));
    } catch (error) {
      set({
        error: error instanceof Error ? error.message : "Failed to load orders",
        isLoading: false,
      });
    }
  },

  getOrderById: (id) => get().orders.find((order) => order.id === id),

  createOrder: async (data) => {
    const createdOrder = await createOrderRequest(data);
    set((state) => ({
      orders: mergeOrders(state.orders, [createdOrder]),
      lastUpdate: new Date().toISOString(),
    }));

    return createdOrder;
  },

  updateOrderStatus: async (id, status) => {
    const updatedOrder = await updateOrderStatusRequest(id, status);
    set((state) => ({
      orders: mergeOrders(state.orders, [updatedOrder]),
      lastUpdate: new Date().toISOString(),
    }));
  },

  updateOrderPayment: async (id, isPaid) => {
    const updatedOrder = await updatePayment(id, isPaid);
    set((state) => ({
      orders: mergeOrders(state.orders, [updatedOrder]),
      lastUpdate: new Date().toISOString(),
    }));
  },

  addDocument: async (orderId, document) => {
    const updatedOrder = await addOrderDocument(orderId, document);
    set((state) => ({
      orders: mergeOrders(state.orders, [updatedOrder]),
      lastUpdate: new Date().toISOString(),
    }));
  },

  removeDocument: async (orderId, documentId) => {
    const updatedOrder = await removeOrderDocument(orderId, documentId);
    set((state) => ({
      orders: mergeOrders(state.orders, [updatedOrder]),
      lastUpdate: new Date().toISOString(),
    }));
  },
}));

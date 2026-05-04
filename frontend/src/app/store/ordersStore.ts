import { create } from 'zustand';
import { persist } from 'zustand/middleware';

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
  status: OrderStatus;
  total: number;
  phone: string; // User's phone for account linking
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
  setOrders: (orders: Order[]) => void;
  addOrder: (order: Order) => void;
  addOrUpdateOrders: (newOrders: Order[]) => void;
  getOrderById: (id: string) => Order | undefined;
  updateOrderStatus: (id: string, status: OrderStatus) => void;
  updateOrderPayment: (id: string, isPaid: boolean) => void;
  addDocument: (orderId: string, document: Document) => void;
  removeDocument: (orderId: string, documentId: number) => void;
  setLastUpdate: (timestamp: string) => void;
};

export const useOrdersStore = create<OrdersStore>()(
  persist(
    (set, get) => ({
  orders: [
    {
      id: "ORD-2L9X8K3P",
      date: "2026-04-21",
      status: "processing",
      total: 898000,
      phone: "+79991234567",
      isPaid: false,
      client: {
        name: "Иванов Иван Иванович",
        phone: "+7 (999) 123-45-67",
        email: "ivanov@example.com",
      },
      deceased: {
        name: "Иванова Мария Петровна",
        dateOfBirth: "1955-03-15",
        dateOfDeath: "2026-04-20",
      },
      services: [
        { name: "Традиционные похороны", price: 450000 },
        { name: "Бальзамирование", price: 65000 },
        { name: "Транспортировка", price: 35000 },
      ],
      products: [
        { name: "Гроб дубовый премиум", price: 320000 },
        { name: "Венок", price: 28000 },
      ],
      documents: [
        {
          id: 1,
          name: "Свидетельство о смерти.pdf",
          type: "PDF",
          date: "2026-04-21",
          size: "1.2 MB",
        },
        {
          id: 2,
          name: "Договор на оказание услуг.pdf",
          type: "PDF",
          date: "2026-04-21",
          size: "850 KB",
        },
      ],
    },
    {
      id: "ORD-5M2N7Q9R",
      date: "2026-04-15",
      status: "confirmed",
      total: 450000,
      phone: "+79991234567",
      isPaid: true,
      client: {
        name: "Петров Петр Петрович",
        phone: "+7 (999) 234-56-78",
        email: "petrov@example.com",
      },
      deceased: {
        name: "Петрова Анна Ивановна",
        dateOfBirth: "1948-07-22",
        dateOfDeath: "2026-04-14",
      },
      services: [
        { name: "Традиционные похороны", price: 450000 },
      ],
      products: [],
      documents: [
        {
          id: 3,
          name: "Свидетельство о смерти.pdf",
          type: "PDF",
          date: "2026-04-15",
          size: "1.1 MB",
        },
      ],
    },
    {
      id: "ORD-8K4P1W6T",
      date: "2026-04-10",
      status: "completed",
      total: 720000,
      phone: "+79993456789",
      isPaid: true,
      client: {
        name: "Сидоров Сидор Сидорович",
        phone: "+7 (999) 345-67-89",
        email: "sidorov@example.com",
      },
      deceased: {
        name: "Сидорова Елена Васильевна",
        dateOfBirth: "1960-11-05",
        dateOfDeath: "2026-04-09",
      },
      services: [
        { name: "Традиционные похороны", price: 450000 },
        { name: "Кремация", price: 120000 },
      ],
      products: [
        { name: "Урна керамическая", price: 150000 },
      ],
      documents: [],
    },
  ],
  lastUpdate: null,
  setOrders: (orders) => set({ orders, lastUpdate: new Date().toISOString() }),
  addOrder: (order) => set((state) => ({
    orders: [...state.orders, order],
    lastUpdate: new Date().toISOString()
  })),
  addOrUpdateOrders: (newOrders) =>
    set((state) => {
      const updatedOrders = [...state.orders];
      newOrders.forEach((newOrder) => {
        const index = updatedOrders.findIndex((o) => o.id === newOrder.id);
        if (index >= 0) {
          // Update existing order
          updatedOrders[index] = newOrder;
        } else {
          // Add new order
          updatedOrders.push(newOrder);
        }
      });
      return {
        orders: updatedOrders,
        lastUpdate: new Date().toISOString(),
      };
    }),
  getOrderById: (id) => get().orders.find((order) => order.id === id),
  setLastUpdate: (timestamp) => set({ lastUpdate: timestamp }),
  updateOrderStatus: (id, status) =>
    set((state) => ({
      orders: state.orders.map((order) =>
        order.id === id ? { ...order, status } : order
      ),
    })),
  updateOrderPayment: (id, isPaid) =>
    set((state) => ({
      orders: state.orders.map((order) =>
        order.id === id ? { ...order, isPaid } : order
      ),
    })),
  addDocument: (orderId, document) =>
    set((state) => ({
      orders: state.orders.map((order) =>
        order.id === orderId
          ? { ...order, documents: [...order.documents, document] }
          : order
      ),
    })),
  removeDocument: (orderId, documentId) =>
    set((state) => ({
      orders: state.orders.map((order) =>
        order.id === orderId
          ? {
              ...order,
              documents: order.documents.filter((doc) => doc.id !== documentId),
            }
          : order
      ),
    })),
  }),
  {
    name: 'orders-storage',
  }
)
);

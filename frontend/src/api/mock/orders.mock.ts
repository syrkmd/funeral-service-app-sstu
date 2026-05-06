import type { CreateOrderData, OrdersApi } from "../orders.api";
import type { Document, Order, OrderStatus } from "../../app/store/ordersStore";

const STORAGE_KEY = "orders-mock-storage";
const LEGACY_STORAGE_KEY = "orders-storage";

const initialOrders: Order[] = [
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
    services: [{ name: "Традиционные похороны", price: 450000 }],
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
    products: [{ name: "Урна керамическая", price: 150000 }],
    documents: [],
  },
];

let memoryOrders = [...initialOrders];

const delay = () => new Promise((resolve) => setTimeout(resolve, 150));

function readOrders(): Order[] {
  if (typeof window === "undefined") {
    return memoryOrders;
  }

  const storedOrders = window.localStorage.getItem(STORAGE_KEY);

  if (!storedOrders) {
    const legacyOrders = readLegacyOrders();

    if (legacyOrders) {
      writeOrders(legacyOrders);
      return legacyOrders;
    }

    writeOrders(initialOrders);
    return initialOrders;
  }

  try {
    return JSON.parse(storedOrders);
  } catch {
    writeOrders(initialOrders);
    return initialOrders;
  }
}

function readLegacyOrders(): Order[] | null {
  if (typeof window === "undefined") {
    return null;
  }

  const storedOrders = window.localStorage.getItem(LEGACY_STORAGE_KEY);

  if (!storedOrders) {
    return null;
  }

  try {
    const parsed = JSON.parse(storedOrders);
    return Array.isArray(parsed?.state?.orders) ? parsed.state.orders : null;
  } catch {
    return null;
  }
}

function writeOrders(orders: Order[]) {
  memoryOrders = orders;

  if (typeof window !== "undefined") {
    window.localStorage.setItem(STORAGE_KEY, JSON.stringify(orders));
  }
}

function requireOrder(id: string): Order {
  const order = readOrders().find((item) => item.id === id);

  if (!order) {
    throw new Error(`Order ${id} was not found`);
  }

  return order;
}

function updateOrder(id: string, updater: (order: Order) => Order): Order {
  let updatedOrder = requireOrder(id);
  const orders = readOrders().map((order) => {
    if (order.id !== id) {
      return order;
    }

    updatedOrder = updater(order);
    return updatedOrder;
  });

  writeOrders(orders);
  return updatedOrder;
}

export const mockOrdersApi: OrdersApi = {
  async getOrders() {
    await delay();
    return readOrders();
  },

  async getOrderById(id) {
    await delay();
    return readOrders().find((order) => order.id === id);
  },

  async getOrdersByPhone(phone) {
    await delay();
    return readOrders().filter((order) => order.phone === phone);
  },

  async createOrder(data: CreateOrderData) {
    await delay();
    const createdAt = data.createdAt || new Date().toISOString();
    const order: Order = {
      ...data,
      createdAt,
      updatedAt: createdAt,
      id: `ORD-${Date.now().toString(36).toUpperCase()}`,
    };

    writeOrders([...readOrders(), order]);
    return order;
  },

  async updateOrderStatus(id: string, status: OrderStatus) {
    await delay();
    const statusUpdatedAt = new Date().toISOString();
    return updateOrder(id, (order) => ({
      ...order,
      status,
      statusUpdatedAt,
      updatedAt: statusUpdatedAt,
    }));
  },

  async updatePayment(id: string, isPaid: boolean) {
    await delay();
    const paymentConfirmedAt = new Date().toISOString();
    return updateOrder(id, (order) => ({
      ...order,
      isPaid,
      paymentConfirmedAt: isPaid ? paymentConfirmedAt : undefined,
      updatedAt: paymentConfirmedAt,
    }));
  },

  async addDocument(orderId: string, document: Document) {
    await delay();
    return updateOrder(orderId, (order) => ({
      ...order,
      documents: [...order.documents, document],
    }));
  },

  async removeDocument(orderId: string, documentId: number) {
    await delay();
    return updateOrder(orderId, (order) => ({
      ...order,
      documents: order.documents.filter((document) => document.id !== documentId),
    }));
  },
};

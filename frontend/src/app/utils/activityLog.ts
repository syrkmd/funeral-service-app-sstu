import type { Order } from "../store/ordersStore";

export type ActivityLogEntry = {
  id: string;
  type: "ORDER_CREATED" | "STATUS_CHANGED" | "PAYMENT_CONFIRMED";
  user: string;
  orderId?: string;
  time: string;
  timestamp: number;
};

function toTimestamp(value: string | undefined) {
  if (!value) {
    return Date.now();
  }

  const timestamp = Date.parse(value);
  return Number.isNaN(timestamp) ? Date.now() : timestamp;
}

function toEventTime(value: string | undefined) {
  return value || new Date().toISOString();
}

export function buildActivityLog(orders: Order[], limit = 10): ActivityLogEntry[] {
  return orders
    .flatMap((order) => {
      const createdAt = order.createdAt || order.date;
      const events: ActivityLogEntry[] = [
        {
          id: `${order.id}:created`,
          type: "ORDER_CREATED",
          user: order.client.name,
          orderId: order.id,
          time: toEventTime(createdAt),
          timestamp: toTimestamp(createdAt),
        },
      ];

      if (order.status !== "processing") {
        const statusChangedAt = order.statusUpdatedAt || order.updatedAt || order.date;

        events.push({
          id: `${order.id}:status:${order.status}`,
          type: "STATUS_CHANGED",
          user: order.client.name,
          orderId: order.id,
          time: toEventTime(statusChangedAt),
          timestamp: toTimestamp(statusChangedAt),
        });
      }

      if (order.isPaid) {
        const paymentConfirmedAt = order.paymentConfirmedAt || order.updatedAt || order.date;

        events.push({
          id: `${order.id}:payment-confirmed`,
          type: "PAYMENT_CONFIRMED",
          user: order.client.name,
          orderId: order.id,
          time: toEventTime(paymentConfirmedAt),
          timestamp: toTimestamp(paymentConfirmedAt),
        });
      }

      return events;
    })
    .sort((a, b) => b.timestamp - a.timestamp)
    .slice(0, limit);
}

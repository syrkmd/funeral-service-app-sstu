import type { Order } from '../store/ordersStore';

/**
 * Fetch orders from the backend
 * Ready for real API integration - replace mock logic with actual fetch/axios calls
 *
 * @param updatedAfter - Optional timestamp to fetch only updates after this time
 * @returns Array of orders
 */
export async function fetchOrders(updatedAfter?: string | null): Promise<Order[]> {
  // Simulate network delay
  await new Promise(resolve => setTimeout(resolve, 300));

  // MOCK IMPLEMENTATION
  // In production, replace with:
  /*
  const url = updatedAfter
    ? `/api/orders?updatedAfter=${encodeURIComponent(updatedAfter)}`
    : '/api/orders';

  const response = await fetch(url);
  if (!response.ok) {
    throw new Error(`Failed to fetch orders: ${response.statusText}`);
  }

  return await response.json();
  */

  // For now, return empty array to avoid duplicating initial mock data
  // Real implementation would fetch from backend
  return [];
}

/**
 * Create a new order
 * Ready for real API integration
 *
 * @param order - Order to create
 * @returns Created order with server-assigned ID
 */
export async function createOrder(order: Omit<Order, 'id'>): Promise<Order> {
  // Simulate network delay
  await new Promise(resolve => setTimeout(resolve, 200));

  // MOCK IMPLEMENTATION
  // In production, replace with:
  /*
  const response = await fetch('/api/orders', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify(order),
  });

  if (!response.ok) {
    throw new Error(`Failed to create order: ${response.statusText}`);
  }

  return await response.json();
  */

  // Mock: generate ID and return order
  return {
    ...order,
    id: `ORD-${Date.now().toString(36).toUpperCase()}`,
  } as Order;
}

/**
 * Update order status
 * Ready for real API integration
 *
 * @param orderId - Order ID
 * @param status - New status
 */
export async function updateOrderStatus(orderId: string, status: string): Promise<void> {
  // Simulate network delay
  await new Promise(resolve => setTimeout(resolve, 150));

  // MOCK IMPLEMENTATION
  // In production, replace with:
  /*
  const response = await fetch(`/api/orders/${orderId}/status`, {
    method: 'PATCH',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({ status }),
  });

  if (!response.ok) {
    throw new Error(`Failed to update order status: ${response.statusText}`);
  }
  */

  console.log(`[API Mock] Updated order ${orderId} status to ${status}`);
}

/**
 * Update order payment status
 * Ready for real API integration
 *
 * @param orderId - Order ID
 * @param isPaid - Payment status
 */
export async function updateOrderPayment(orderId: string, isPaid: boolean): Promise<void> {
  // Simulate network delay
  await new Promise(resolve => setTimeout(resolve, 150));

  // MOCK IMPLEMENTATION
  // In production, replace with:
  /*
  const response = await fetch(`/api/orders/${orderId}/payment`, {
    method: 'PATCH',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({ isPaid }),
  });

  if (!response.ok) {
    throw new Error(`Failed to update payment status: ${response.statusText}`);
  }
  */

  console.log(`[API Mock] Updated order ${orderId} payment to ${isPaid ? 'paid' : 'unpaid'}`);
}

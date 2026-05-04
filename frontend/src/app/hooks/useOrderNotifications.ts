import { useEffect, useRef } from 'react';
import { useOrdersStore } from '../store/ordersStore';

/**
 * Hook to show notifications when new orders are received
 * Tracks order count and shows console notification when count increases
 */
export function useOrderNotifications() {
  const orders = useOrdersStore((state) => state.orders);
  const previousCountRef = useRef(orders.length);

  useEffect(() => {
    const currentCount = orders.length;
    const previousCount = previousCountRef.current;

    if (currentCount > previousCount) {
      const newOrdersCount = currentCount - previousCount;
      console.log(
        `%c[New Orders] ${newOrdersCount} new order${newOrdersCount > 1 ? 's' : ''} received!`,
        'background: #4CAF50; color: white; padding: 4px 8px; border-radius: 4px; font-weight: bold;'
      );

      // In a real app, you could show a toast notification here
      // Example: toast.success(`${newOrdersCount} new order(s) received`);
    }

    previousCountRef.current = currentCount;
  }, [orders.length]);
}

import { useEffect, useRef } from 'react';
import { useOrdersStore } from '../store/ordersStore';
import { fetchOrders } from '../utils/ordersApi';

const POLLING_INTERVAL = 5000; // 5 seconds

export function useOrderPolling() {
  const addOrUpdateOrders = useOrdersStore((state) => state.addOrUpdateOrders);
  const lastUpdate = useOrdersStore((state) => state.lastUpdate);
  const pollingRef = useRef<NodeJS.Timeout | null>(null);

  useEffect(() => {
    // Initial fetch on mount
    const loadOrders = async () => {
      try {
        const orders = await fetchOrders(lastUpdate);
        if (orders && orders.length > 0) {
          addOrUpdateOrders(orders);
          console.log(`[Order Polling] Loaded ${orders.length} orders`);
        }
      } catch (error) {
        console.error('[Order Polling] Error loading orders:', error);
      }
    };

    loadOrders();

    // Start polling
    pollingRef.current = setInterval(async () => {
      // Only poll if tab is visible
      if (document.visibilityState !== 'visible') {
        return;
      }

      try {
        const orders = await fetchOrders(lastUpdate);
        if (orders && orders.length > 0) {
          addOrUpdateOrders(orders);
          console.log(`[Order Polling] Received ${orders.length} order updates`);
        }
      } catch (error) {
        console.error('[Order Polling] Error fetching orders:', error);
      }
    }, POLLING_INTERVAL);

    // Cleanup
    return () => {
      if (pollingRef.current) {
        clearInterval(pollingRef.current);
      }
    };
  }, [addOrUpdateOrders, lastUpdate]);
}

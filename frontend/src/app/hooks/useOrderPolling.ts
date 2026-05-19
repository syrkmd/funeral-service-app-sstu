import { useEffect, useRef } from 'react';
import { useOrdersStore } from '../store/ordersStore';

const POLLING_INTERVAL = 5000; // 5 seconds

export function useOrderPolling() {
  const pollingRef = useRef<ReturnType<typeof setInterval> | null>(null);

  useEffect(() => {
    const loadLatestOrders = async () => {
      const { lastUpdate, loadOrders } = useOrdersStore.getState();
      await loadOrders(lastUpdate);
    };

    // Initial fetch on mount
    const loadInitialOrders = async () => {
      try {
        await loadLatestOrders();
      } catch (error) {
        console.error('[Order Polling] Error loading orders:', error);
      }
    };

    loadInitialOrders();

    // Start polling
    pollingRef.current = setInterval(async () => {
      if (document.visibilityState !== 'visible') {
        return;
      }

      try {
        await loadLatestOrders();
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
  }, []);
}

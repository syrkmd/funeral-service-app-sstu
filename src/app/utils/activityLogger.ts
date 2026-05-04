import { useMetricsStore } from '../store/metricsStore';

export function logOrderCreated(userName: string, orderId: string) {
  const addActivityLogEntry = useMetricsStore.getState().addActivityLogEntry;
  const incrementOrdersToday = useMetricsStore.getState().incrementOrdersToday;

  addActivityLogEntry({
    type: 'ORDER_CREATED',
    user: userName,
    orderId,
    time: new Date().toISOString(),
  });

  incrementOrdersToday();
}

export function logStatusChanged(userName: string, orderId: string) {
  const addActivityLogEntry = useMetricsStore.getState().addActivityLogEntry;

  addActivityLogEntry({
    type: 'STATUS_CHANGED',
    user: userName,
    orderId,
    time: new Date().toISOString(),
  });
}

export function logPaymentConfirmed(userName: string, orderId: string) {
  const addActivityLogEntry = useMetricsStore.getState().addActivityLogEntry;

  addActivityLogEntry({
    type: 'PAYMENT_CONFIRMED',
    user: userName,
    orderId,
    time: new Date().toISOString(),
  });
}

export async function fetchMetrics() {
  await new Promise((resolve) => setTimeout(resolve, 100));

  return {
    totalRequests: Math.floor(Math.random() * 1000) + 45000,
    errors: Math.floor(Math.random() * 10) + 20,
    activeClients: Math.floor(Math.random() * 200) + 1200,
    avgLatency: Math.floor(Math.random() * 10) + 20,
    traffic: parseFloat((Math.random() * 0.5 + 1.0).toFixed(2)),
  };
}

export async function fetchRPSData() {
  await new Promise((resolve) => setTimeout(resolve, 50));

  const now = new Date();

  return {
    time: now.toLocaleTimeString("ru-RU", {
      hour: "2-digit",
      minute: "2-digit",
      second: "2-digit",
    }),
    value: Math.floor(Math.random() * 40) + 30,
  };
}

export async function fetchActivityLog() {
  await new Promise((resolve) => setTimeout(resolve, 100));

  return [];
}

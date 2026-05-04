// import { RouterProvider } from "react-router";
// import { router } from "./routes";

// export default function App() {
//   return <RouterProvider router={router} />;
// }

import { RouterProvider } from "react-router";
import { router } from "./routes";
import { useEffect } from "react";
import { useOrdersStore } from "./store/ordersStore";


export default function App() {
  useEffect(() => {
    const handleStorageChange = (event: StorageEvent) => {
      if (event.key === "orders-storage" && event.newValue) {
        try {
          const newData = JSON.parse(event.newValue);

          if (newData?.state?.orders) {
            useOrdersStore.getState().setOrders(newData.state.orders);
          }
        } catch (e) {
          console.error("Ошибка парсинга localStorage:", e);
        }
      }
    };

    window.addEventListener("storage", handleStorageChange);

    return () => window.removeEventListener("storage", handleStorageChange);
  }, []);

  return <RouterProvider router={router} />;
}
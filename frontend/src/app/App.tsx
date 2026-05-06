import { RouterProvider } from "react-router";
import { router } from "./routes";
import { useEffect } from "react";
import { useOrdersStore } from "./store/ordersStore";

export default function App() {
  useEffect(() => {
    useOrdersStore.getState().loadOrders();
  }, []);

  return <RouterProvider router={router} />;
}

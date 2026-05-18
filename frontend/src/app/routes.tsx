import { createBrowserRouter } from "react-router";
import { Root } from "./components/Root";
import { Home } from "./components/Home";
import { Services } from "./components/Services";
import { Catalog } from "./components/Catalog";
import { OrderForm } from "./components/OrderForm";
import { Contacts } from "./components/Contacts";
import { AdminLogin } from "./components/admin/AdminLogin";
import { AdminLayout } from "./components/admin/AdminLayout";
import { Dashboard } from "./components/admin/Dashboard";
import { Orders } from "./components/admin/Orders";
import { OrderDetails } from "./components/admin/OrderDetails";
import { IPAccess } from "./components/admin/IPAccess";
import { RateLimiting } from "./components/admin/RateLimiting";
import { AccountLogin } from "./components/account/AccountLogin";
import { AccountVerify } from "./components/account/AccountVerify";
import { AccountLayout } from "./components/account/AccountLayout";
import { AccountOrders } from "./components/account/AccountOrders";
import { AccountOrderDetails } from "./components/account/AccountOrderDetails";

export const router = createBrowserRouter([
  {
    path: "/",
    Component: Root,
    children: [
      { index: true, Component: Home },
      { path: "services", Component: Services },
      { path: "catalog", Component: Catalog },
      { path: "order", Component: OrderForm },
      { path: "contacts", Component: Contacts },
    ],
  },
  {
    path: "/admin/login",
    Component: AdminLogin,
  },
  {
    path: "/admin",
    Component: AdminLayout,
    children: [
      { index: true, Component: Dashboard },
      { path: "orders", Component: Orders },
      { path: "orders/:orderId", Component: OrderDetails },
      { path: "ip-access", Component: IPAccess },
      { path: "rate-limiting", Component: RateLimiting },
    ],
  },
  {
    path: "/account/login",
    Component: AccountLogin,
  },
  {
    path: "/account/verify",
    Component: AccountVerify,
  },
  {
    path: "/account",
    Component: AccountLayout,
    children: [
      { index: true, Component: AccountOrders },
      { path: "orders/:orderId", Component: AccountOrderDetails },
    ],
  },
]);

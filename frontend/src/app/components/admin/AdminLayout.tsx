import { useEffect } from "react";
import { Link, Outlet, useLocation, useNavigate } from "react-router";
import { useAdminAuthStore } from "../../store/adminAuthStore";

export function AdminLayout() {
  const location = useLocation();
  const navigate = useNavigate();
  const isAdminAuth = useAdminAuthStore((state) => state.isAdminAuth);
  const clearAdminAuth = useAdminAuthStore((state) => state.clearAdminAuth);

  useEffect(() => {
    // Check authentication
    if (!isAdminAuth) {
      navigate("/admin/login");
    }
  }, [isAdminAuth, navigate]);

  const handleLogout = () => {
    clearAdminAuth();
    navigate("/admin/login");
  };

  const isActive = (path: string) => {
    return location.pathname === path || location.pathname.startsWith(path + '/');
  };

  const menuItems = [
    { path: "/admin", label: "Dashboard", exact: true },
    { path: "/admin/orders", label: "Orders" },
    { path: "/admin/ip-access", label: "IP Access" },
    { path: "/admin/logs", label: "Logs" },
    { path: "/admin/rate-limiting", label: "Rate Limiting" },
  ];

  return (
    <div className="flex min-h-screen bg-background">
      {/* Sidebar */}
      <aside className="w-64 bg-card border-r border-border">
        <div className="p-6 border-b border-border">
          <h1 className="text-xl text-foreground">Admin Panel</h1>
          <p className="text-sm text-muted-foreground">Управление системой</p>
        </div>
        <nav className="p-4">
          <ul className="space-y-2">
            {menuItems.map((item) => {
              const active = item.exact
                ? location.pathname === item.path
                : isActive(item.path);

              return (
                <li key={item.path}>
                  <Link
                    to={item.path}
                    className={`block px-4 py-2 rounded-lg transition-colors ${
                      active
                        ? "bg-primary text-primary-foreground"
                        : "text-muted-foreground hover:bg-secondary hover:text-foreground"
                    }`}
                  >
                    {item.label}
                  </Link>
                </li>
              );
            })}
          </ul>
        </nav>
      </aside>

      {/* Main Content */}
      <div className="flex-1 flex flex-col">
        {/* Topbar */}
        <header className="bg-card border-b border-border px-8 py-4 flex items-center justify-between">
          <h2 className="text-xl text-foreground">
            {menuItems.find(item =>
              item.exact
                ? location.pathname === item.path
                : location.pathname.startsWith(item.path)
            )?.label || "Admin"}
          </h2>
          <button
            onClick={handleLogout}
            className="px-4 py-2 text-sm text-muted-foreground hover:text-foreground transition-colors"
          >
            Выйти
          </button>
        </header>

        {/* Main Content Area */}
        <main className="flex-1 p-8">
          <Outlet />
        </main>
      </div>
    </div>
  );
}

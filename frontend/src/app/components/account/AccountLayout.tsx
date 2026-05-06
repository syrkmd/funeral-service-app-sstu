import { useEffect } from "react";
import { Outlet, Link, useNavigate, useLocation } from "react-router";
import { useUserStore } from "../../store/userStore";

export function AccountLayout() {
  const navigate = useNavigate();
  const location = useLocation();
  const restoreSession = useUserStore((state) => state.restoreSession);
  const clearUserPhone = useUserStore((state) => state.clearUserPhone);

  useEffect(() => {
    restoreSession().then((isAuthenticated) => {
      if (!isAuthenticated) {
        navigate("/account/login");
      }
    });
  }, [navigate, restoreSession]);

  const handleLogout = async () => {
    await clearUserPhone();
    navigate("/account/login");
  };

  const isActive = (path: string) => {
    return location.pathname === path || location.pathname.startsWith(path + '/');
  };

  return (
    <div className="min-h-screen bg-background">
      <header className="border-b border-border bg-card">
        <nav className="max-w-7xl mx-auto px-6 py-4">
          <div className="flex items-center justify-between">
            <Link to="/" className="text-xl text-foreground">
              Ритуальные услуги
            </Link>
            <div className="flex items-center gap-8">
              <Link
                to="/account"
                className={`transition-colors ${
                  isActive("/account")
                    ? "text-primary"
                    : "text-muted-foreground hover:text-foreground"
                }`}
              >
                Мои заказы
              </Link>
              <button
                onClick={handleLogout}
                className="text-muted-foreground hover:text-foreground transition-colors"
              >
                Выйти
              </button>
            </div>
          </div>
        </nav>
      </header>

      <main className="max-w-7xl mx-auto px-6 py-12">
        <Outlet />
      </main>
    </div>
  );
}

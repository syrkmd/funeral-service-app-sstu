import { useEffect } from "react";
import { Outlet, Link, useNavigate, useLocation } from "react-router";
import { useUserStore } from "../../store/userStore";
import { normalizePhone } from "../../utils/phoneUtils";

export function AccountLayout() {
  const navigate = useNavigate();
  const location = useLocation();
  const setUserPhone = useUserStore((state) => state.setUserPhone);
  const clearUserPhone = useUserStore((state) => state.clearUserPhone);

  useEffect(() => {
    // Проверка авторизации (демо)
    const isAuthenticated = localStorage.getItem("account_authenticated");
    const storedPhone = localStorage.getItem("account_phone");

    if (!isAuthenticated) {
      navigate("/account/login");
    } else if (storedPhone) {
      // Restore normalized phone to store on page load
      const normalizedPhone = normalizePhone(storedPhone);
      setUserPhone(normalizedPhone);
    }
  }, [navigate, setUserPhone]);

  const handleLogout = () => {
    localStorage.removeItem("account_authenticated");
    localStorage.removeItem("account_phone");
    clearUserPhone();
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

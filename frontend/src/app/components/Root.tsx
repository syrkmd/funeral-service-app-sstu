import { Outlet, Link, useLocation } from "react-router";

export function Root() {
  const location = useLocation();

  const isActive = (path: string) => {
    if (path === "/") {
      return location.pathname === "/";
    }
    return location.pathname.startsWith(path);
  };

  return (
    <div className="min-h-screen bg-background">
      <header className="border-b border-border bg-card">
        <nav className="max-w-7xl mx-auto px-6 py-4">
          <div className="flex items-center justify-between">
            <Link to="/" className="text-xl text-foreground">
              Ритуальные услуги
            </Link>
            <div className="flex gap-8">
              <Link
                to="/services"
                className={`transition-colors ${
                  isActive("/services")
                    ? "text-primary"
                    : "text-muted-foreground hover:text-foreground"
                }`}
              >
                Услуги
              </Link>
              <Link
                to="/catalog"
                className={`transition-colors ${
                  isActive("/catalog")
                    ? "text-primary"
                    : "text-muted-foreground hover:text-foreground"
                }`}
              >
                Каталог
              </Link>
              <Link
                to="/order"
                className={`transition-colors ${
                  isActive("/order")
                    ? "text-primary"
                    : "text-muted-foreground hover:text-foreground"
                }`}
              >
                Заказ
              </Link>
              <Link
                to="/contacts"
                className={`transition-colors ${
                  isActive("/contacts")
                    ? "text-primary"
                    : "text-muted-foreground hover:text-foreground"
                }`}
              >
                Контакты
              </Link>
              <Link
                to="/account"
                className={`transition-colors ${
                  isActive("/account")
                    ? "text-primary"
                    : "text-muted-foreground hover:text-foreground"
                }`}
              >
                Личный кабинет
              </Link>
            </div>
          </div>
        </nav>
      </header>
      <main>
        <Outlet />
      </main>
      <footer className="border-t border-border bg-card mt-24">
        <div className="max-w-7xl mx-auto px-6 py-8">
          <p className="text-muted-foreground text-center">
            © 2026 Ритуальные услуги. Работаем круглосуточно
          </p>
        </div>
      </footer>
    </div>
  );
}

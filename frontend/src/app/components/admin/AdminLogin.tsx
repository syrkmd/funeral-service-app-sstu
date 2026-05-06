import { useState } from "react";
import { useNavigate } from "react-router";
import { useAdminAuthStore } from "../../store/adminAuthStore";

export function AdminLogin() {
  const navigate = useNavigate();
  const setAdminAuth = useAdminAuthStore((state) => state.setAdminAuth);
  const [login, setLogin] = useState("");
  const [password, setPassword] = useState("");

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    if (await setAdminAuth(login, password)) {
      navigate("/admin");
    }
  };

  return (
    <div className="min-h-screen bg-background flex items-center justify-center px-6">
      <div className="w-full max-w-md">
        <div className="bg-card border border-border rounded-lg p-8">
          <h1 className="text-2xl text-foreground mb-2 text-center">Admin Panel</h1>
          <p className="text-sm text-muted-foreground mb-8 text-center">
            Вход в панель управления
          </p>

          <form onSubmit={handleSubmit} className="space-y-6">
            <div>
              <label htmlFor="login" className="block mb-2 text-foreground">
                Логин
              </label>
              <input
                id="login"
                type="text"
                required
                value={login}
                onChange={(e) => setLogin(e.target.value)}
                className="w-full px-4 py-3 bg-input-background border border-border rounded-lg focus:outline-none focus:ring-2 focus:ring-ring"
                placeholder="Введите логин"
              />
            </div>

            <div>
              <label htmlFor="password" className="block mb-2 text-foreground">
                Пароль
              </label>
              <input
                id="password"
                type="password"
                required
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                className="w-full px-4 py-3 bg-input-background border border-border rounded-lg focus:outline-none focus:ring-2 focus:ring-ring"
                placeholder="Введите пароль"
              />
            </div>

            <button
              type="submit"
              className="w-full bg-primary text-primary-foreground py-3 rounded-lg hover:opacity-90 transition-opacity"
            >
              Войти
            </button>
          </form>

          <div className="mt-6 text-center">
            <a href="/" className="text-sm text-muted-foreground hover:text-foreground">
              ← Вернуться на сайт
            </a>
          </div>
        </div>
      </div>
    </div>
  );
}

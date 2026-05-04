import { useState } from "react";
import { useNavigate } from "react-router";

export function AccountLogin() {
  const navigate = useNavigate();
  const [phone, setPhone] = useState("");
  const [error, setError] = useState("");

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    setError("");

    if (!phone.trim()) {
      setError("Введите номер телефона");
      return;
    }

    if (!/^[\d\s\-\(\)\+]+$/.test(phone)) {
      setError("Некорректный номер телефона");
      return;
    }

    // В реальном приложении здесь был бы API-запрос
    navigate("/account/verify", { state: { phone } });
  };

  return (
    <div className="min-h-screen bg-background flex items-center justify-center px-6">
      <div className="w-full max-w-md">
        <div className="bg-card border border-border rounded-lg p-8">
          <h1 className="text-2xl text-foreground mb-2 text-center">Личный кабинет</h1>
          <p className="text-sm text-muted-foreground mb-8 text-center">
            Введите номер телефона для входа
          </p>

          <form onSubmit={handleSubmit} className="space-y-6">
            <div>
              <label htmlFor="phone" className="block mb-2 text-foreground">
                Номер телефона *
              </label>
              <input
                id="phone"
                type="tel"
                required
                value={phone}
                onChange={(e) => setPhone(e.target.value)}
                className={`w-full px-4 py-3 bg-input-background border rounded-lg focus:outline-none focus:ring-2 focus:ring-ring ${
                  error ? "border-destructive" : "border-border"
                }`}
                placeholder="+7 (999) 123-45-67"
              />
              {error && (
                <p className="text-destructive text-sm mt-1">{error}</p>
              )}
            </div>

            <button
              type="submit"
              className="w-full bg-primary text-primary-foreground py-3 rounded-lg hover:opacity-90 transition-opacity"
            >
              Получить код
            </button>
          </form>

          <div className="mt-6 text-center">
            <a href="/" className="text-sm text-muted-foreground hover:text-foreground">
              ← Вернуться на главную
            </a>
          </div>
        </div>
      </div>
    </div>
  );
}

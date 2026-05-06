import { useState } from "react";
import { useNavigate, useLocation } from "react-router";
import { useUserStore } from "../../store/userStore";
import { normalizePhone } from "../../utils/phoneUtils";

export function AccountVerify() {
  const navigate = useNavigate();
  const location = useLocation();
  const phone = location.state?.phone || "";
  const [code, setCode] = useState("");
  const [error, setError] = useState("");
  const verifyCode = useUserStore((state) => state.verifyCode);

  const DEMO_CODE = "4821";

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError("");

    if (!code.trim()) {
      setError("Введите код");
      return;
    }

    // Normalize phone for consistent matching with orders
    const normalizedPhone = normalizePhone(phone);
    const isVerified = await verifyCode(normalizedPhone, code);

    if (!isVerified) {
      setError("Неверный код");
      return;
    }

    navigate("/account");
  };

  if (!phone) {
    navigate("/account/login");
    return null;
  }

  return (
    <div className="min-h-screen bg-background flex items-center justify-center px-6">
      <div className="w-full max-w-md">
        <div className="bg-card border border-border rounded-lg p-8">
          <h1 className="text-2xl text-foreground mb-2 text-center">Подтверждение</h1>
          <p className="text-sm text-muted-foreground mb-6 text-center">
            Код отправлен на номер {phone}
          </p>

          {/* Demo hint */}
          <div className="bg-secondary/50 border border-border rounded-lg p-4 mb-6">
            <p className="text-sm text-muted-foreground mb-1">
              Введите код, отправленный на ваш номер
            </p>
            <p className="text-sm text-muted-foreground">
              (в демо отображается на экране)
            </p>
            <p className="text-lg text-foreground mt-2">
              Код: <span className="font-mono">{DEMO_CODE}</span>
            </p>
          </div>

          <form onSubmit={handleSubmit} className="space-y-6">
            <div>
              <label htmlFor="code" className="block mb-2 text-foreground">
                Введите код *
              </label>
              <input
                id="code"
                type="text"
                required
                value={code}
                onChange={(e) => setCode(e.target.value)}
                className={`w-full px-4 py-3 bg-input-background border rounded-lg focus:outline-none focus:ring-2 focus:ring-ring font-mono text-lg text-center tracking-wider ${
                  error ? "border-destructive" : "border-border"
                }`}
                placeholder="____"
                maxLength={4}
              />
              {error && (
                <p className="text-destructive text-sm mt-1">{error}</p>
              )}
            </div>

            <button
              type="submit"
              className="w-full bg-primary text-primary-foreground py-3 rounded-lg hover:opacity-90 transition-opacity"
            >
              Войти
            </button>
          </form>

          <div className="mt-6 text-center">
            <button
              onClick={() => navigate("/account/login")}
              className="text-sm text-muted-foreground hover:text-foreground"
            >
              ← Изменить номер
            </button>
          </div>
        </div>
      </div>
    </div>
  );
}

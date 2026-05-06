import { useEffect, useState } from "react";
import { useProxyStore, type IPAccessType } from "../../store/proxyStore";

export function IPAccess() {
  const ipRules = useProxyStore((state) => state.ipRules);
  const defaultPolicy = useProxyStore((state) => state.defaultPolicy);
  const loadConfig = useProxyStore((state) => state.loadConfig);
  const addIPRule = useProxyStore((state) => state.addIPRule);
  const removeIPRule = useProxyStore((state) => state.removeIPRule);
  const setDefaultPolicy = useProxyStore((state) => state.setDefaultPolicy);
  const checkIPAccess = useProxyStore((state) => state.checkIPAccess);

  const [showForm, setShowForm] = useState(false);
  const [newIP, setNewIP] = useState("");
  const [newType, setNewType] = useState<IPAccessType>("allow");
  const [newNote, setNewNote] = useState("");

  const [checkIP, setCheckIP] = useState("");
  const [checkResult, setCheckResult] = useState<string | null>(null);

  useEffect(() => {
    loadConfig();
  }, [loadConfig]);

  const handleAddIP = async () => {
    if (!newIP) return;

    await addIPRule({
      ip: newIP,
      type: newType,
      note: newNote || undefined,
    });

    setNewIP("");
    setNewType("allow");
    setNewNote("");
    setShowForm(false);
  };

  const handleCheckIP = () => {
    if (!checkIP) return;
    const result = checkIPAccess(checkIP);
    setCheckResult(result);
  };

  const getTypeLabel = (type: IPAccessType) => {
    if (type === "allow") return "Разрешён";
    if (type === "deny") return "Запрещён";
    return "Проверка";
  };

  const getTypeBadge = (type: IPAccessType) => {
    const styles = {
      allow: "bg-primary/20 text-primary border border-primary/40",
      deny: "bg-destructive/20 text-destructive border border-destructive/40",
      gray: "bg-secondary text-secondary-foreground border border-border",
    };

    return (
      <span className={`px-3 py-1 rounded-full text-xs ${styles[type]}`}>
        {getTypeLabel(type)}
      </span>
    );
  };

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex justify-between items-center">
        <h2 className="text-2xl text-foreground">IP Access Control</h2>
        <button
          onClick={() => setShowForm(!showForm)}
          className="px-4 py-2 bg-primary text-primary-foreground rounded-lg hover:opacity-90 transition-opacity"
        >
          {showForm ? "Отмена" : "+ Добавить IP"}
        </button>
      </div>

      {/* Default Policy */}
      <div className="bg-card border border-border rounded-lg p-6">
        <h3 className="text-lg mb-4 text-foreground">Политика по умолчанию</h3>
        <div className="flex items-center gap-4">
          <label className="flex items-center gap-2">
            <input
              type="radio"
              checked={defaultPolicy === "allow"}
              onChange={() => setDefaultPolicy("allow")}
              className="w-4 h-4"
            />
            <span className="text-foreground">Разрешить</span>
          </label>
          <label className="flex items-center gap-2">
            <input
              type="radio"
              checked={defaultPolicy === "deny"}
              onChange={() => setDefaultPolicy("deny")}
              className="w-4 h-4"
            />
            <span className="text-foreground">Запретить</span>
          </label>
        </div>
        <p className="text-sm text-muted-foreground mt-2">
          Применяется к IP-адресам, не входящим в списки
        </p>
      </div>

      {/* IP Check Tool */}
      <div className="bg-card border border-border rounded-lg p-6">
        <h3 className="text-lg mb-4 text-foreground">Проверка IP</h3>
        <div className="flex gap-4 items-start">
          <div className="flex-1">
            <input
              type="text"
              placeholder="192.168.1.1"
              value={checkIP}
              onChange={(e) => setCheckIP(e.target.value)}
              className="w-full px-4 py-2 bg-input-background border border-border rounded-lg focus:outline-none focus:ring-2 focus:ring-ring"
            />
          </div>
          <button
            onClick={handleCheckIP}
            className="px-6 py-2 bg-primary text-primary-foreground rounded-lg hover:opacity-90 transition-opacity"
          >
            Проверить доступ
          </button>
        </div>
        {checkResult && (
          <div className="mt-4 p-4 bg-secondary/30 border border-border rounded-lg">
            <p className="text-sm text-foreground">
              Результат:{" "}
              <span className="font-medium">
                {checkResult === "allowed" && "✓ Разрешён"}
                {checkResult === "denied" && "✗ Запрещён"}
                {checkResult === "captcha" && "⚠ Требуется проверка (captcha)"}
              </span>
            </p>
          </div>
        )}
      </div>

      {/* Add Form */}
      {showForm && (
        <div className="bg-card border border-border rounded-lg p-6">
          <h3 className="text-lg mb-4 text-foreground">Новое правило</h3>
          <div className="space-y-4">
            <div>
              <label className="block text-sm text-foreground mb-2">
                IP адрес *
              </label>
              <input
                type="text"
                placeholder="192.168.1.1 или 10.0.0.0/8 или 192.168.1.1-192.168.1.255"
                value={newIP}
                onChange={(e) => setNewIP(e.target.value)}
                className="w-full px-4 py-2 bg-input-background border border-border rounded-lg focus:outline-none focus:ring-2 focus:ring-ring"
              />
              <p className="text-xs text-muted-foreground mt-1">
                Поддерживаемые форматы: одиночный IP, CIDR, диапазон
              </p>
            </div>
            <div>
              <label className="block text-sm text-foreground mb-2">Тип *</label>
              <select
                value={newType}
                onChange={(e) => setNewType(e.target.value as IPAccessType)}
                className="w-full px-4 py-2 bg-input-background border border-border rounded-lg focus:outline-none focus:ring-2 focus:ring-ring"
              >
                <option value="allow">Разрешён</option>
                <option value="deny">Запрещён</option>
                <option value="gray">Проверка (captcha)</option>
              </select>
            </div>
            <div>
              <label className="block text-sm text-foreground mb-2">
                Заметка (опционально)
              </label>
              <input
                type="text"
                placeholder="Описание правила"
                value={newNote}
                onChange={(e) => setNewNote(e.target.value)}
                className="w-full px-4 py-2 bg-input-background border border-border rounded-lg focus:outline-none focus:ring-2 focus:ring-ring"
              />
            </div>
            <button
              onClick={handleAddIP}
              className="px-6 py-2 bg-primary text-primary-foreground rounded-lg hover:opacity-90 transition-opacity"
            >
              Добавить
            </button>
          </div>
        </div>
      )}

      {/* Rules Table */}
      <div className="bg-card border border-border rounded-lg overflow-hidden">
        <table className="w-full">
          <thead className="bg-secondary/50 border-b border-border">
            <tr>
              <th className="px-6 py-3 text-left text-xs text-muted-foreground">
                IP Address
              </th>
              <th className="px-6 py-3 text-left text-xs text-muted-foreground">
                Type
              </th>
              <th className="px-6 py-3 text-left text-xs text-muted-foreground">
                Note
              </th>
              <th className="px-6 py-3 text-left text-xs text-muted-foreground">
                Added
              </th>
              <th className="px-6 py-3 text-left text-xs text-muted-foreground">
                Actions
              </th>
            </tr>
          </thead>
          <tbody>
            {ipRules.map((rule) => (
              <tr
                key={rule.id}
                className="border-b border-border last:border-0 hover:bg-secondary/30 transition-colors"
              >
                <td className="px-6 py-4 text-sm text-foreground font-mono">
                  {rule.ip}
                </td>
                <td className="px-6 py-4">{getTypeBadge(rule.type)}</td>
                <td className="px-6 py-4 text-sm text-muted-foreground">
                  {rule.note || "-"}
                </td>
                <td className="px-6 py-4 text-sm text-muted-foreground">
                  {rule.addedAt}
                </td>
                <td className="px-6 py-4">
                  <button
                    onClick={() => removeIPRule(rule.id)}
                    className="text-sm text-destructive hover:underline"
                  >
                    Удалить
                  </button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
}

import { useCallback, useEffect, useState } from "react";
import { fetchDashboardIPAccess, sumValues, type DashboardIPAccessResponse } from "../../../api/metrics.api";
import { checkIPAccess as checkIPAccessRequest } from "../../../api/proxy.api";
import { useProxyStore, type IPAccessType } from "../../store/proxyStore";

type IPCheckResult = {
  status: "allowed" | "denied" | "captcha" | "error";
  reason?: string;
  matchedRuleId?: string;
  matchedValue?: string;
};

export function IPAccess() {
  const ipRules = useProxyStore((state) => state.ipRules);
  const loadConfig = useProxyStore((state) => state.loadConfig);
  const addIPRule = useProxyStore((state) => state.addIPRule);
  const removeIPRule = useProxyStore((state) => state.removeIPRule);
  const [ipAccessStats, setIPAccessStats] = useState<DashboardIPAccessResponse | null>(null);
  const [loadError, setLoadError] = useState<string | null>(null);
  const [actionError, setActionError] = useState<string | null>(null);

  const [checkIP, setCheckIP] = useState("");
  const [checkResult, setCheckResult] = useState<IPCheckResult | null>(null);
  const [showForm, setShowForm] = useState(false);
  const [newIP, setNewIP] = useState("");
  const [newType, setNewType] = useState<IPAccessType>("allow");
  const [newNote, setNewNote] = useState("");

  const loadIPAccess = useCallback(async () => {
    try {
      await loadConfig();
      const stats = await fetchDashboardIPAccess();
      setIPAccessStats(stats);
      setLoadError(null);
    } catch (error) {
      setLoadError(error instanceof Error ? error.message : "Failed to load IP access data");
    }
  }, [loadConfig]);

  useEffect(() => {
    loadIPAccess();
    const interval = setInterval(loadIPAccess, 2000);

    return () => clearInterval(interval);
  }, [loadIPAccess]);

  const openAddForm = () => {
    setShowForm(true);
    setActionError(null);
  };

  const handleAddIP = async () => {
    const normalizedIP = newIP.trim();
    if (!normalizedIP) return;

    try {
      await addIPRule({
        ip: normalizedIP,
        type: newType,
        note: newNote.trim() || undefined,
      });
      setNewIP("");
      setNewNote("");
      setShowForm(false);
      setActionError(null);
      await loadIPAccess();
    } catch (error) {
      setActionError(error instanceof Error ? error.message : "Failed to add IP rule");
    }
  };

  const handleDeleteRule = async (id: string) => {
    try {
      await removeIPRule(id);
      setActionError(null);
      await loadIPAccess();
    } catch (error) {
      setActionError(error instanceof Error ? error.message : "Failed to delete IP rule");
    }
  };

  const handleCheckIP = async () => {
    const normalizedIP = checkIP.trim();
    if (!normalizedIP) return;

    if (normalizedIP.includes("/") || normalizedIP.includes("-")) {
      setCheckResult({
        status: "error",
        reason: "Введите конкретный IP-адрес, например 10.0.0.5",
      });
      return;
    }

    try {
      const result = await checkIPAccessRequest(normalizedIP);
      const status = result.verificationRequired || result.decision === "gray"
        ? "captcha"
        : result.allowed
        ? "allowed"
        : "denied";

      setCheckResult({
        status,
        reason: result.reason,
        matchedRuleId: result.matchedRuleId,
        matchedValue: result.matchedValue,
      });
    } catch (error) {
      setCheckResult({
        status: "error",
        reason: error instanceof Error ? error.message : "Failed to check IP",
      });
    }
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

  const allowCount = ipRules.filter((rule) => rule.type === "allow").length;
  const denyCount = ipRules.filter((rule) => rule.type === "deny").length;
  const grayCount = ipRules.filter((rule) => rule.type === "gray").length;
  const decisionCount = sumValues(ipAccessStats?.denyStatistics);

  return (
    <div className="space-y-6">
      {loadError && (
        <div className="bg-card border border-border rounded-lg p-4 text-sm text-destructive">
          {loadError}
        </div>
      )}
      {actionError && (
        <div className="bg-card border border-border rounded-lg p-4 text-sm text-destructive">
          {actionError}
        </div>
      )}

      {/* Header */}
      <div className="flex justify-between items-center">
        <h2 className="text-2xl text-foreground">IP Access Control</h2>
        <button
          onClick={openAddForm}
          className="px-4 py-2 bg-primary text-primary-foreground rounded-lg hover:opacity-90 transition-opacity"
        >
          Add Rule
        </button>
      </div>

      {/* IP Stats */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
        <div className="bg-card border border-border rounded-lg p-6">
          <div className="text-sm text-muted-foreground mb-2">Allowlist</div>
          <div className="text-3xl text-foreground">{allowCount}</div>
        </div>
        <div className="bg-card border border-border rounded-lg p-6">
          <div className="text-sm text-muted-foreground mb-2">Denylist</div>
          <div className="text-3xl text-foreground">{denyCount}</div>
        </div>
        <div className="bg-card border border-border rounded-lg p-6">
          <div className="text-sm text-muted-foreground mb-2">Graylist</div>
          <div className="text-3xl text-foreground">{grayCount}</div>
        </div>
        <div className="bg-card border border-border rounded-lg p-6">
          <div className="text-sm text-muted-foreground mb-2">Decisions</div>
          <div className="text-3xl text-foreground">{decisionCount}</div>
        </div>
      </div>

      {/* Add Form */}
      {showForm && (
        <div className="bg-card border border-border rounded-lg p-6">
          <div className="flex items-center justify-between mb-4">
            <h3 className="text-lg text-foreground">Новое IP правило</h3>
            <button
              onClick={() => setShowForm(false)}
              className="text-sm text-muted-foreground hover:text-foreground"
            >
              Отмена
            </button>
          </div>
          <div className="space-y-4">
            <div>
              <label className="block text-sm text-foreground mb-2">
                IP / CIDR *
              </label>
              <input
                type="text"
                placeholder="10.0.0.5 или 10.0.0.0/8"
                value={newIP}
                onChange={(event) => setNewIP(event.target.value)}
                className="w-full px-4 py-2 bg-input-background border border-border rounded-lg focus:outline-none focus:ring-2 focus:ring-ring"
              />
            </div>
            <div>
              <label className="block text-sm text-foreground mb-2">Type *</label>
              <select
                value={newType}
                onChange={(event) => setNewType(event.target.value as IPAccessType)}
                className="w-full px-4 py-2 bg-input-background border border-border rounded-lg focus:outline-none focus:ring-2 focus:ring-ring"
              >
                <option value="allow">Allow</option>
                <option value="deny">Deny</option>
                <option value="gray">Gray / Challenge</option>
              </select>
            </div>
            <div>
              <label className="block text-sm text-foreground mb-2">
                Note / comment
              </label>
              <input
                type="text"
                placeholder="Описание правила"
                value={newNote}
                onChange={(event) => setNewNote(event.target.value)}
                className="w-full px-4 py-2 bg-input-background border border-border rounded-lg focus:outline-none focus:ring-2 focus:ring-ring"
              />
            </div>
            <button
              onClick={handleAddIP}
              disabled={!newIP.trim()}
              className="px-6 py-2 bg-primary text-primary-foreground rounded-lg hover:opacity-90 transition-opacity disabled:opacity-50 disabled:cursor-not-allowed"
            >
              Добавить правило
            </button>
          </div>
        </div>
      )}

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
                {checkResult.status === "allowed" && "✓ Разрешён"}
                {checkResult.status === "denied" && "✗ Запрещён"}
                {checkResult.status === "captcha" && "⚠ Требуется проверка (captcha)"}
                {checkResult.status === "error" && `Ошибка: ${checkResult.reason}`}
              </span>
            </p>
            {checkResult.reason && checkResult.status !== "error" && (
              <p className="text-sm text-muted-foreground mt-2">
                Reason: {checkResult.reason}
              </p>
            )}
            {checkResult.matchedRuleId && (
              <p className="text-sm text-muted-foreground mt-1">
                Matched rule: {checkResult.matchedRuleId}
                {checkResult.matchedValue && ` (${checkResult.matchedValue})`}
              </p>
            )}
          </div>
        )}
      </div>

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
                    onClick={() => handleDeleteRule(rule.id)}
                    className="text-sm text-destructive hover:underline"
                  >
                    Delete
                  </button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
        {ipRules.length === 0 && (
          <div className="text-center py-8 text-muted-foreground">
            IP rules не найдены
          </div>
        )}
      </div>

      {/* Matched Rules Stats */}
      <div className="bg-card border border-border rounded-lg p-6">
        <h3 className="text-lg mb-4 text-foreground">Статистика правил</h3>
        {Object.keys(ipAccessStats?.matchedRulesStatistics || {}).length === 0 ? (
          <div className="text-center py-8 text-muted-foreground">
            Совпадений по правилам пока нет
          </div>
        ) : (
          <div className="space-y-3">
            {Object.entries(ipAccessStats?.matchedRulesStatistics || {}).map(([ruleId, count]) => (
              <div key={ruleId} className="flex items-center justify-between py-3 border-b border-border last:border-0">
                <div className="text-sm text-foreground font-mono">{ruleId}</div>
                <div className="text-sm text-muted-foreground">{count.toLocaleString()}</div>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}

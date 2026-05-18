import { useEffect, useState } from "react";
import { fetchDashboardRateLimits, sumValues, type DashboardRateLimitsResponse } from "../../../api/metrics.api";
import { useProxyStore } from "../../store/proxyStore";

export function RateLimiting() {
  const rateLimitSettings = useProxyStore((state) => state.rateLimitSettings);
  const loadConfig = useProxyStore((state) => state.loadConfig);
  const updateRateLimits = useProxyStore((state) => state.updateRateLimits);
  const [liveRateLimits, setLiveRateLimits] = useState<DashboardRateLimitsResponse | null>(null);

  const [editMode, setEditMode] = useState(false);
  const [rpsLimit, setRpsLimit] = useState(rateLimitSettings.rpsLimit);
  const [rpmLimit, setRpmLimit] = useState(rateLimitSettings.rpmLimit);
  const [saveMessage, setSaveMessage] = useState(false);
  const [loadError, setLoadError] = useState<string | null>(null);

  useEffect(() => {
    loadConfig();
  }, [loadConfig]);

  useEffect(() => {
    const loadRateLimits = async () => {
      try {
        const data = await fetchDashboardRateLimits();
        setLiveRateLimits(data);
        setLoadError(null);
      } catch (error) {
        setLoadError(error instanceof Error ? error.message : "Failed to load rate limits");
      }
    };

    loadRateLimits();
    const interval = setInterval(loadRateLimits, 2000);

    return () => clearInterval(interval);
  }, []);

  useEffect(() => {
    setRpsLimit(rateLimitSettings.rpsLimit);
    setRpmLimit(rateLimitSettings.rpmLimit);
  }, [rateLimitSettings.rpmLimit, rateLimitSettings.rpsLimit]);

  const primaryRule = liveRateLimits?.activeRateLimitRules[0];
  const effectiveSettings = {
    rpsLimit: primaryRule?.rps || rateLimitSettings.rpsLimit,
    rpmLimit: primaryRule?.rpm || rateLimitSettings.rpmLimit,
  };

  const getBucketUsage = (type: string, limit: number) => {
    const buckets = (liveRateLimits?.currentBucketUsage || []).filter((bucket) => bucket.type === type);
    if (!buckets.length || limit <= 0) return 0;

    const lowestTokensRemaining = Math.min(...buckets.map((bucket) => bucket.tokensRemaining));
    return Math.max(0, Math.round(limit - lowestTokensRemaining));
  };

  const limits = {
    rps: {
      limit: effectiveSettings.rpsLimit,
      current: getBucketUsage("rps", effectiveSettings.rpsLimit),
      percentage: effectiveSettings.rpsLimit > 0 ? (getBucketUsage("rps", effectiveSettings.rpsLimit) / effectiveSettings.rpsLimit) * 100 : 0,
    },
    rpm: {
      limit: effectiveSettings.rpmLimit,
      current: getBucketUsage("rpm", effectiveSettings.rpmLimit),
      percentage: effectiveSettings.rpmLimit > 0 ? (getBucketUsage("rpm", effectiveSettings.rpmLimit) / effectiveSettings.rpmLimit) * 100 : 0,
    },
  };

  const violators = (liveRateLimits?.blockedIps || []).map((item) => ({
        ip: item.ip,
        requests: item.count,
        time: "Live",
      }));

  const violationTotal = sumValues(liveRateLimits?.violations);

  const handleSave = async () => {
    await updateRateLimits({
      rpsLimit,
      rpmLimit,
    });
    setEditMode(false);
    setSaveMessage(true);
    setTimeout(() => setSaveMessage(false), 3000);
  };

  const handleCancel = () => {
    setRpsLimit(rateLimitSettings.rpsLimit);
    setRpmLimit(rateLimitSettings.rpmLimit);
    setEditMode(false);
  };

  return (
    <div className="space-y-6">
      {loadError && (
        <div className="bg-card border border-border rounded-lg p-4 text-sm text-destructive">
          {loadError}
        </div>
      )}

      {/* Settings */}
      <div className="bg-card border border-border rounded-lg p-6">
        <div className="flex items-center justify-between mb-4">
          <h3 className="text-lg text-foreground">Настройки лимитов</h3>
          {!editMode ? (
            <span className="text-sm text-muted-foreground">Read-only monitoring</span>
          ) : (
            <div className="flex gap-2">
              <button
                onClick={handleCancel}
                className="px-4 py-2 bg-secondary text-secondary-foreground border border-border rounded-lg hover:bg-accent transition-colors text-sm"
              >
                Отмена
              </button>
              <button
                onClick={handleSave}
                className="px-4 py-2 bg-primary text-primary-foreground rounded-lg hover:opacity-90 transition-opacity text-sm"
              >
                Сохранить
              </button>
            </div>
          )}
        </div>

        {saveMessage && (
          <div className="mb-4 p-3 bg-primary/10 border border-primary/30 rounded-lg text-sm text-primary">
            Настройки сохранены
          </div>
        )}

        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <div>
            <label className="block text-sm text-foreground mb-2">
              RPS Limit (Requests Per Second)
            </label>
            <input
              type="number"
              value={rpsLimit}
              onChange={(e) => setRpsLimit(Number(e.target.value))}
              disabled={!editMode}
              className="w-full px-4 py-2 bg-input-background border border-border rounded-lg focus:outline-none focus:ring-2 focus:ring-ring disabled:opacity-50"
            />
            <p className="text-xs text-muted-foreground mt-1">
              Текущий лимит: {effectiveSettings.rpsLimit} RPS
            </p>
          </div>
          <div>
            <label className="block text-sm text-foreground mb-2">
              RPM Limit (Requests Per Minute)
            </label>
            <input
              type="number"
              value={rpmLimit}
              onChange={(e) => setRpmLimit(Number(e.target.value))}
              disabled={!editMode}
              className="w-full px-4 py-2 bg-input-background border border-border rounded-lg focus:outline-none focus:ring-2 focus:ring-ring disabled:opacity-50"
            />
            <p className="text-xs text-muted-foreground mt-1">
              Текущий лимит: {effectiveSettings.rpmLimit} RPM
            </p>
          </div>
        </div>
      </div>

      {/* Current Limits */}
      <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
        <div className="bg-card border border-border rounded-lg p-6">
          <h3 className="text-lg mb-4 text-foreground">
            Requests Per Second (RPS)
          </h3>
          <div className="space-y-3">
            <div className="flex justify-between items-end">
              <span className="text-3xl text-foreground">{limits.rps.current}</span>
              <span className="text-sm text-muted-foreground">
                / {limits.rps.limit}
              </span>
            </div>
            <div className="w-full bg-secondary rounded-full h-2">
              <div
                className="bg-primary h-2 rounded-full transition-all"
                style={{ width: `${Math.min(limits.rps.percentage, 100)}%` }}
              />
            </div>
            <div className="text-sm text-muted-foreground">
              {limits.rps.percentage.toFixed(1)}% используется
            </div>
          </div>
        </div>

        <div className="bg-card border border-border rounded-lg p-6">
          <h3 className="text-lg mb-4 text-foreground">
            Requests Per Minute (RPM)
          </h3>
          <div className="space-y-3">
            <div className="flex justify-between items-end">
              <span className="text-3xl text-foreground">
                {limits.rpm.current.toLocaleString()}
              </span>
              <span className="text-sm text-muted-foreground">
                / {limits.rpm.limit.toLocaleString()}
              </span>
            </div>
            <div className="w-full bg-secondary rounded-full h-2">
              <div
                className="bg-primary h-2 rounded-full transition-all"
                style={{ width: `${Math.min(limits.rpm.percentage, 100)}%` }}
              />
            </div>
            <div className="text-sm text-muted-foreground">
              {limits.rpm.percentage.toFixed(2)}% используется
            </div>
          </div>
        </div>
      </div>

      {/* Active Rules */}
      <div className="bg-card border border-border rounded-lg p-6">
        <div className="flex items-center justify-between mb-4">
          <h3 className="text-lg text-foreground">Активные правила</h3>
          <span className="text-sm text-muted-foreground">
            Нарушений: {violationTotal.toLocaleString()}
          </span>
        </div>
        <div className="space-y-3">
          {(liveRateLimits?.activeRateLimitRules || []).map((rule) => (
            <div key={rule.id} className="grid grid-cols-2 md:grid-cols-4 gap-4 p-4 border border-border rounded-lg text-sm">
              <div>
                <div className="text-muted-foreground">Rule</div>
                <div className="text-foreground font-mono">{rule.id}</div>
              </div>
              <div>
                <div className="text-muted-foreground">Scope</div>
                <div className="text-foreground">{rule.scope}: {rule.value}</div>
              </div>
              <div>
                <div className="text-muted-foreground">RPS / RPM</div>
                <div className="text-foreground">{rule.rps} / {rule.rpm}</div>
              </div>
              <div>
                <div className="text-muted-foreground">Connections</div>
                <div className="text-foreground">{rule.maxConnections}</div>
              </div>
            </div>
          ))}
          {!liveRateLimits?.activeRateLimitRules.length && (
            <div className="text-center py-8 text-muted-foreground">
              Активные правила не найдены
            </div>
          )}
        </div>
      </div>

      {/* Violators */}
      <div className="bg-card border border-border rounded-lg p-6">
        <h3 className="text-lg mb-4 text-foreground">Нарушители лимитов</h3>
        <div className="space-y-3">
          {violators.map((violator, index) => (
            <div
              key={index}
              className="flex items-center justify-between p-4 border border-border rounded-lg"
            >
              <div>
                <div className="text-sm text-foreground font-mono">
                  {violator.ip}
                </div>
                <div className="text-xs text-muted-foreground">
                  Последнее нарушение: {violator.time}
                </div>
              </div>
              <div className="flex items-center gap-4">
                <div className="text-right">
                  <div className="text-sm text-destructive">
                    {violator.requests} запросов
                  </div>
                  <div className="text-xs text-muted-foreground">
                    Превышение лимита
                  </div>
                </div>
              </div>
            </div>
          ))}
        </div>

        {violators.length === 0 && (
          <div className="text-center py-8 text-muted-foreground">
            Нарушений не обнаружено
          </div>
        )}
      </div>
    </div>
  );
}

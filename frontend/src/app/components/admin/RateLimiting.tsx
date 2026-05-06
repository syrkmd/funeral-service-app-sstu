import { useEffect, useState } from "react";
import { useProxyStore } from "../../store/proxyStore";

export function RateLimiting() {
  const rateLimitSettings = useProxyStore((state) => state.rateLimitSettings);
  const loadConfig = useProxyStore((state) => state.loadConfig);
  const updateRateLimits = useProxyStore((state) => state.updateRateLimits);

  const [editMode, setEditMode] = useState(false);
  const [rpsLimit, setRpsLimit] = useState(rateLimitSettings.rpsLimit);
  const [rpmLimit, setRpmLimit] = useState(rateLimitSettings.rpmLimit);
  const [saveMessage, setSaveMessage] = useState(false);

  useEffect(() => {
    loadConfig();
  }, [loadConfig]);

  useEffect(() => {
    setRpsLimit(rateLimitSettings.rpsLimit);
    setRpmLimit(rateLimitSettings.rpmLimit);
  }, [rateLimitSettings.rpmLimit, rateLimitSettings.rpsLimit]);

  // Mock current usage
  const limits = {
    rps: {
      limit: rateLimitSettings.rpsLimit,
      current: 45,
      percentage: (45 / rateLimitSettings.rpsLimit) * 100,
    },
    rpm: {
      limit: rateLimitSettings.rpmLimit,
      current: 2341,
      percentage: (2341 / rateLimitSettings.rpmLimit) * 100,
    },
  };

  const violators = [
    { ip: "192.168.1.50", requests: 150, time: "14:30:45" },
    { ip: "10.0.0.25", requests: 132, time: "14:28:12" },
    { ip: "172.16.0.8", requests: 118, time: "14:25:33" },
  ];

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
      {/* Settings */}
      <div className="bg-card border border-border rounded-lg p-6">
        <div className="flex items-center justify-between mb-4">
          <h3 className="text-lg text-foreground">Настройки лимитов</h3>
          {!editMode ? (
            <button
              onClick={() => setEditMode(true)}
              className="px-4 py-2 bg-primary text-primary-foreground rounded-lg hover:opacity-90 transition-opacity text-sm"
            >
              Изменить
            </button>
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
              Текущий лимит: {rateLimitSettings.rpsLimit} RPS
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
              Текущий лимит: {rateLimitSettings.rpmLimit} RPM
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
                <button className="px-4 py-2 text-sm bg-destructive/10 text-destructive border border-destructive/40 rounded-lg hover:bg-destructive/20 transition-colors">
                  Заблокировать
                </button>
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

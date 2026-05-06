type Document = {
  id: number;
  name: string;
  type: string;
  date: string;
  size: string;
};

const accountDocuments: Document[] = [
  {
    id: 1,
    name: "Свидетельство о смерти",
    type: "PDF",
    date: "2026-04-20",
    size: "1.2 MB",
  },
  {
    id: 2,
    name: "Договор на оказание услуг",
    type: "PDF",
    date: "2026-04-21",
    size: "850 KB",
  },
];

export function AccountDocuments() {
  const handleDownload = (documentName: string) => {
    // В реальном приложении здесь был бы запрос к API
    alert(`Скачивание: ${documentName}`);
  };

  if (accountDocuments.length === 0) {
    return (
      <div className="text-center py-16">
        <div className="mb-6">
          <svg
            className="w-20 h-20 text-muted-foreground mx-auto"
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
          >
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth={1.5}
              d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"
            />
          </svg>
        </div>
        <h2 className="text-xl text-foreground mb-2">Документы не найдены</h2>
        <p className="text-muted-foreground">У вас пока нет доступных документов</p>
      </div>
    );
  }

  return (
    <div>
      <h1 className="text-3xl text-foreground mb-8">Документы</h1>

      <div className="space-y-4">
        {accountDocuments.map((document) => (
          <div
            key={document.id}
            className="bg-card border border-border rounded-lg p-6 hover:shadow-sm transition-shadow"
          >
            <div className="flex items-start justify-between">
              <div className="flex items-start gap-4">
                <div className="w-12 h-12 bg-primary/10 rounded-lg flex items-center justify-center flex-shrink-0">
                  <svg
                    className="w-6 h-6 text-primary"
                    fill="none"
                    stroke="currentColor"
                    viewBox="0 0 24 24"
                  >
                    <path
                      strokeLinecap="round"
                      strokeLinejoin="round"
                      strokeWidth={2}
                      d="M7 21h10a2 2 0 002-2V9.414a1 1 0 00-.293-.707l-5.414-5.414A1 1 0 0012.586 3H7a2 2 0 00-2 2v14a2 2 0 002 2z"
                    />
                  </svg>
                </div>
                <div>
                  <h3 className="text-lg text-foreground mb-1">{document.name}</h3>
                  <div className="flex items-center gap-3 text-sm text-muted-foreground">
                    <span>{document.type}</span>
                    <span>•</span>
                    <span>{document.size}</span>
                    <span>•</span>
                    <span>{document.date}</span>
                  </div>
                </div>
              </div>

              <button
                onClick={() => handleDownload(document.name)}
                className="px-4 py-2 bg-primary text-primary-foreground rounded-lg hover:opacity-90 transition-opacity flex items-center gap-2"
              >
                <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth={2}
                    d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4"
                  />
                </svg>
                Скачать
              </button>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}

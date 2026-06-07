import { type FormEvent, useEffect, useState } from "react";
import { useParams, useNavigate } from "react-router";
import {
  getCatalogProducts,
  getFuneralServices,
  type CatalogProductDto,
  type FuneralServiceDto,
} from "../../../api/catalog.api";
import { useOrdersStore, type OrderStatus } from "../../store/ordersStore";

const predefinedFiles = [
  { value: "death_certificate.pdf", label: "Свидетельство о смерти" },
  { value: "contract.pdf", label: "Договор на оказание услуг" },
  { value: "payment_receipt.pdf", label: "Квитанция об оплате" },
  { value: "medical_report.pdf", label: "Медицинское заключение" },
];

type EditableSection = "client" | "deceased" | "date" | "ceremony" | "services" | "products" | "discount" | null;

export function OrderDetails() {
  const { orderId } = useParams();
  const navigate = useNavigate();
  const [showAddDocument, setShowAddDocument] = useState(false);
  const [newDocName, setNewDocName] = useState("");
  const [newDocType, setNewDocType] = useState("PDF");
  const [newDocFile, setNewDocFile] = useState("");
  const [statusUpdateMessage, setStatusUpdateMessage] = useState(false);
  const [editingSection, setEditingSection] = useState<EditableSection>(null);
  const [availableServices, setAvailableServices] = useState<FuneralServiceDto[]>([]);
  const [availableProducts, setAvailableProducts] = useState<CatalogProductDto[]>([]);
  const [selectedServiceIds, setSelectedServiceIds] = useState<number[]>([]);
  const [selectedProductIds, setSelectedProductIds] = useState<number[]>([]);
  const [catalogError, setCatalogError] = useState("");
  const [saveMessage, setSaveMessage] = useState("");

  const order = useOrdersStore((state) =>
    state.orders.find((o) => o.id === orderId)
  );
  const updateOrderStatus = useOrdersStore((state) => state.updateOrderStatus);
  const updateOrderPayment = useOrdersStore((state) => state.updateOrderPayment);
  const updateOrderClient = useOrdersStore((state) => state.updateOrderClient);
  const updateOrderDeceased = useOrdersStore((state) => state.updateOrderDeceased);
  const updateOrderDate = useOrdersStore((state) => state.updateOrderDate);
  const updateOrderCeremony = useOrdersStore((state) => state.updateOrderCeremony);
  const replaceOrderServices = useOrdersStore((state) => state.replaceOrderServices);
  const replaceOrderProducts = useOrdersStore((state) => state.replaceOrderProducts);
  const applyOrderDiscount = useOrdersStore((state) => state.applyOrderDiscount);
  const addDocument = useOrdersStore((state) => state.addDocument);
  const removeDocument = useOrdersStore((state) => state.removeDocument);

  const finishEdit = (message: string) => {
    setEditingSection(null);
    setSaveMessage(message);
    setTimeout(() => setSaveMessage(""), 3000);
  };

  useEffect(() => {
    let isMounted = true;

    async function loadCatalogItems() {
      try {
        const [services, products] = await Promise.all([
          getFuneralServices(),
          getCatalogProducts(),
        ]);

        if (!isMounted) return;

        setAvailableServices(services);
        setAvailableProducts(products);
      } catch {
        if (isMounted) {
          setCatalogError("Не удалось загрузить каталог услуг и товаров");
        }
      }
    }

    loadCatalogItems();

    return () => {
      isMounted = false;
    };
  }, []);

  const handleStatusChange = async (newStatus: OrderStatus) => {
    if (orderId && order) {
      await updateOrderStatus(orderId, newStatus);
      setStatusUpdateMessage(true);
      setTimeout(() => setStatusUpdateMessage(false), 3000);
    }
  };

  const handlePaymentChange = async () => {
    if (orderId && order) {
      const newPaidStatus = !order.isPaid;
      await updateOrderPayment(orderId, newPaidStatus);
    }
  };

  const handleAddDocument = async () => {
    if (!orderId || !newDocName.trim() || !newDocFile) return;

    const newDocument = {
      id: Date.now(),
      name: newDocName.trim(),
      type: newDocType,
      size: "1.2 MB", // Mock size
      date: new Date().toISOString().split("T")[0],
    };

    await addDocument(orderId, newDocument);

    // Reset form
    setNewDocName("");
    setNewDocType("PDF");
    setNewDocFile("");
    setShowAddDocument(false);
  };

  const handleRemoveDocument = async (documentId: number) => {
    if (!orderId) return;
    if (confirm("Удалить этот документ?")) {
      await removeDocument(orderId, documentId);
    }
  };

  const handleEditClick = (section: EditableSection) => {
    if (!order) return;

    if (section === "services") {
      setSelectedServiceIds(
        availableServices
          .filter((service) => order.services.some((item) => item.name === service.title))
          .map((service) => service.id)
      );
    }

    if (section === "products") {
      setSelectedProductIds(
        availableProducts
          .filter((product) => order.products.some((item) => item.name === product.title))
          .map((product) => product.id)
      );
    }

    setEditingSection(section);
  };

  const toggleServiceSelection = (serviceId: number) => {
    setSelectedServiceIds((current) =>
      current.includes(serviceId)
        ? current.filter((id) => id !== serviceId)
        : [...current, serviceId]
    );
  };

  const toggleProductSelection = (productId: number) => {
    setSelectedProductIds((current) =>
      current.includes(productId)
        ? current.filter((id) => id !== productId)
        : [...current, productId]
    );
  };

  const handleClientSubmit = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    if (!order) return;

    const form = new FormData(event.currentTarget);
    await updateOrderClient(order.id, {
      name: String(form.get("name") || ""),
      phone: String(form.get("phone") || ""),
      email: String(form.get("email") || ""),
    });
    finishEdit("Данные клиента обновлены");
  };

  const handleDeceasedSubmit = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    if (!order) return;

    const form = new FormData(event.currentTarget);
    await updateOrderDeceased(order.id, {
      name: String(form.get("name") || ""),
      dateOfBirth: String(form.get("dateOfBirth") || ""),
      dateOfDeath: String(form.get("dateOfDeath") || ""),
    });
    finishEdit("Данные умершего обновлены");
  };

  const handleDateSubmit = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    if (!order) return;

    const form = new FormData(event.currentTarget);
    await updateOrderDate(order.id, String(form.get("date") || ""));
    finishEdit("Дата заказа обновлена");
  };

  const handleCeremonySubmit = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    if (!order) return;

    const form = new FormData(event.currentTarget);
    await updateOrderCeremony(order.id, {
      serviceDate: String(form.get("serviceDate") || ""),
      serviceTime: String(form.get("serviceTime") || ""),
      serviceAddress: String(form.get("serviceAddress") || ""),
      cemetery: String(form.get("cemetery") || "Основное кладбище"),
      cemeteryNotes: String(form.get("cemeteryNotes") || ""),
    });
    finishEdit("Детали церемонии обновлены");
  };

  const handleServicesSubmit = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    if (!order) return;

    if (selectedServiceIds.length === 0) {
      alert("Выберите хотя бы одну услугу");
      return;
    }

    await replaceOrderServices(order.id, selectedServiceIds);
    finishEdit("Услуги обновлены");
  };

  const handleProductsSubmit = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    if (!order) return;

    await replaceOrderProducts(order.id, selectedProductIds);
    finishEdit("Товары обновлены");
  };

  const handleDiscountSubmit = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    if (!order) return;

    const form = new FormData(event.currentTarget);
    const discountAmount = Number(form.get("discountAmount"));

    if (Number.isNaN(discountAmount) || discountAmount < 0) {
      alert("Введите корректную скидку");
      return;
    }

    await applyOrderDiscount(order.id, discountAmount, String(form.get("reason") || ""));
    finishEdit("Скидка применена");
  };

  if (!order) {
    return (
      <div className="text-center py-16">
        <h2 className="text-xl text-foreground mb-2">Заказ не найден</h2>
        <p className="text-muted-foreground mb-6">
          Заказ с номером {orderId} не существует
        </p>
        <button
          onClick={() => navigate("/admin/orders")}
          className="bg-primary text-primary-foreground px-6 py-3 rounded-lg hover:opacity-90 transition-opacity"
        >
          Вернуться к заказам
        </button>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <button
            onClick={() => navigate('/admin/orders')}
            className="text-sm text-muted-foreground hover:text-foreground mb-2"
          >
            ← Назад к заказам
          </button>
          <h2 className="text-2xl text-foreground">Заказ {order.id}</h2>
          <div className="flex items-center gap-3 mt-1">
            <p className="text-sm text-muted-foreground">{order.date}</p>
            <button
              onClick={() => handleEditClick("date")}
              className="text-sm text-primary hover:underline"
            >
              Изменить дату
            </button>
          </div>
          {editingSection === "date" && (
            <form onSubmit={handleDateSubmit} className="mt-3 flex gap-2">
              <input
                name="date"
                type="date"
                required
                defaultValue={order.date}
                className="px-3 py-2 bg-input-background border border-border rounded-lg focus:outline-none focus:ring-2 focus:ring-ring"
              />
              <button className="px-4 py-2 bg-primary text-primary-foreground rounded-lg">
                Сохранить
              </button>
            </form>
          )}
          {saveMessage && (
            <p className="text-sm text-primary mt-2">{saveMessage}</p>
          )}
        </div>
        <div className="flex gap-3 items-center">
          <div className="relative">
            <select
              value={order.status}
              onChange={(e) => handleStatusChange(e.target.value as OrderStatus)}
              className="px-4 py-2 bg-input-background border border-border rounded-lg focus:outline-none focus:ring-2 focus:ring-ring"
            >
              <option value="processing">В обработке</option>
              <option value="confirmed">Подтверждён</option>
              <option value="completed">Завершён</option>
              <option value="cancelled">Отменён</option>
            </select>
            {statusUpdateMessage && (
              <div className="absolute top-full mt-2 left-0 whitespace-nowrap bg-primary text-primary-foreground px-3 py-1 rounded text-sm">
                Статус обновлён
              </div>
            )}
          </div>
          <button
            onClick={handlePaymentChange}
            disabled={order.isPaid}
            className={`px-4 py-2 rounded-lg transition-colors ${
              order.isPaid
                ? "bg-primary/20 text-primary border border-primary/40 cursor-default"
                : "bg-secondary text-secondary-foreground border border-border hover:bg-accent"
            }`}
          >
            {order.isPaid ? "Оплачен" : "Отметить оплаченным"}
          </button>
        </div>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Client Info */}
        <div className="bg-card border border-border rounded-lg p-6">
          <div className="flex items-center justify-between mb-4">
            <h3 className="text-lg text-foreground">Данные клиента</h3>
            <button
              onClick={() => handleEditClick("client")}
              className="text-sm text-primary hover:underline"
            >
              Изменить
            </button>
          </div>
          {editingSection === "client" ? (
            <form onSubmit={handleClientSubmit} className="space-y-3">
              <input
                name="name"
                required
                defaultValue={order.client.name}
                className="w-full px-4 py-2 bg-input-background border border-border rounded-lg focus:outline-none focus:ring-2 focus:ring-ring"
                placeholder="ФИО"
              />
              <input
                name="phone"
                required
                defaultValue={order.client.phone}
                className="w-full px-4 py-2 bg-input-background border border-border rounded-lg focus:outline-none focus:ring-2 focus:ring-ring"
                placeholder="Телефон"
              />
              <input
                name="email"
                type="email"
                defaultValue={order.client.email}
                className="w-full px-4 py-2 bg-input-background border border-border rounded-lg focus:outline-none focus:ring-2 focus:ring-ring"
                placeholder="Email"
              />
              <div className="flex gap-2">
                <button className="px-4 py-2 bg-primary text-primary-foreground rounded-lg">Сохранить</button>
                <button type="button" onClick={() => setEditingSection(null)} className="px-4 py-2 bg-secondary text-secondary-foreground rounded-lg">Отмена</button>
              </div>
            </form>
          ) : (
          <div className="space-y-3 text-sm">
            <div>
              <span className="text-muted-foreground">ФИО:</span>
              <p className="text-foreground">{order.client.name}</p>
            </div>
            <div>
              <span className="text-muted-foreground">Телефон:</span>
              <p className="text-foreground">{order.client.phone}</p>
            </div>
            <div>
              <span className="text-muted-foreground">Email:</span>
              <p className="text-foreground">{order.client.email}</p>
            </div>
          </div>
          )}
        </div>

        {/* Deceased Info */}
        <div className="bg-card border border-border rounded-lg p-6">
          <div className="flex items-center justify-between mb-4">
            <h3 className="text-lg text-foreground">Данные умершего</h3>
            <button
              onClick={() => handleEditClick("deceased")}
              className="text-sm text-primary hover:underline"
            >
              Изменить
            </button>
          </div>
          {editingSection === "deceased" ? (
            <form onSubmit={handleDeceasedSubmit} className="space-y-3">
              <input
                name="name"
                required
                defaultValue={order.deceased.name}
                className="w-full px-4 py-2 bg-input-background border border-border rounded-lg focus:outline-none focus:ring-2 focus:ring-ring"
                placeholder="ФИО умершего"
              />
              <input
                name="dateOfBirth"
                type="date"
                defaultValue={order.deceased.dateOfBirth}
                className="w-full px-4 py-2 bg-input-background border border-border rounded-lg focus:outline-none focus:ring-2 focus:ring-ring"
              />
              <input
                name="dateOfDeath"
                type="date"
                required
                defaultValue={order.deceased.dateOfDeath}
                className="w-full px-4 py-2 bg-input-background border border-border rounded-lg focus:outline-none focus:ring-2 focus:ring-ring"
              />
              <div className="flex gap-2">
                <button className="px-4 py-2 bg-primary text-primary-foreground rounded-lg">Сохранить</button>
                <button type="button" onClick={() => setEditingSection(null)} className="px-4 py-2 bg-secondary text-secondary-foreground rounded-lg">Отмена</button>
              </div>
            </form>
          ) : (
          <div className="space-y-3 text-sm">
            <div>
              <span className="text-muted-foreground">ФИО:</span>
              <p className="text-foreground">{order.deceased.name}</p>
            </div>
            <div>
              <span className="text-muted-foreground">Дата рождения:</span>
              <p className="text-foreground">{order.deceased.dateOfBirth}</p>
            </div>
            <div>
              <span className="text-muted-foreground">Дата смерти:</span>
              <p className="text-foreground">{order.deceased.dateOfDeath}</p>
            </div>
          </div>
          )}
        </div>
      </div>

      {/* Ceremony Info */}
      <div className="bg-card border border-border rounded-lg p-6">
        <div className="flex items-center justify-between mb-4">
          <h3 className="text-lg text-foreground">Детали церемонии</h3>
          <button
            onClick={() => handleEditClick("ceremony")}
            className="text-sm text-primary hover:underline"
          >
            Изменить
          </button>
        </div>
        {editingSection === "ceremony" ? (
          <form onSubmit={handleCeremonySubmit} className="space-y-4">
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <input
                name="serviceDate"
                type="date"
                required
                defaultValue={order.serviceDate || order.date}
                className="px-4 py-2 bg-input-background border border-border rounded-lg focus:outline-none focus:ring-2 focus:ring-ring"
              />
              <input
                name="serviceTime"
                type="time"
                required
                defaultValue={order.serviceTime || "10:00"}
                className="px-4 py-2 bg-input-background border border-border rounded-lg focus:outline-none focus:ring-2 focus:ring-ring"
              />
              <input
                name="serviceAddress"
                required
                defaultValue={order.serviceAddress || ""}
                className="px-4 py-2 bg-input-background border border-border rounded-lg focus:outline-none focus:ring-2 focus:ring-ring"
                placeholder="Место проведения"
              />
              <input
                name="cemetery"
                required
                defaultValue={order.cemetery || "Основное кладбище"}
                className="px-4 py-2 bg-input-background border border-border rounded-lg focus:outline-none focus:ring-2 focus:ring-ring"
                placeholder="Кладбище"
              />
            </div>
            <textarea
              name="cemeteryNotes"
              rows={3}
              defaultValue={order.cemeteryNotes || ""}
              className="w-full px-4 py-2 bg-input-background border border-border rounded-lg focus:outline-none focus:ring-2 focus:ring-ring resize-none"
              placeholder="Примечания"
            />
            <div className="flex gap-2">
              <button className="px-4 py-2 bg-primary text-primary-foreground rounded-lg">Сохранить</button>
              <button type="button" onClick={() => setEditingSection(null)} className="px-4 py-2 bg-secondary text-secondary-foreground rounded-lg">Отмена</button>
            </div>
          </form>
        ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4 text-sm">
          <div>
            <span className="text-muted-foreground">Дата и время:</span>
            <p className="text-foreground">
              {order.serviceDate || "Не указано"}
              {order.serviceTime ? ` в ${order.serviceTime}` : ""}
            </p>
          </div>
          <div>
            <span className="text-muted-foreground">Место проведения:</span>
            <p className="text-foreground">{order.serviceAddress || "Не указано"}</p>
          </div>
          <div>
            <span className="text-muted-foreground">Место захоронения:</span>
            <p className="text-foreground">{order.cemeteryPlotCode || "Не выбрано"}</p>
          </div>
          <div>
            <span className="text-muted-foreground">Примечания:</span>
            <p className="text-foreground">{order.cemeteryNotes || "Нет примечаний"}</p>
          </div>
        </div>
        )}
      </div>

      {/* Services */}
      <div className="bg-card border border-border rounded-lg p-6">
        <div className="flex items-center justify-between mb-4">
          <h3 className="text-lg text-foreground">Услуги</h3>
          <button
            onClick={() => handleEditClick("services")}
            className="text-sm text-primary hover:underline"
          >
            Изменить
          </button>
        </div>
        {editingSection === "services" ? (
          <form onSubmit={handleServicesSubmit} className="space-y-3">
            {catalogError && (
              <p className="text-sm text-destructive">{catalogError}</p>
            )}
            <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
              {availableServices.map((service) => {
                const isSelected = selectedServiceIds.includes(service.id);

                return (
                  <label
                    key={service.id}
                    className={`flex items-start gap-3 border rounded-lg p-4 cursor-pointer transition-colors ${
                      isSelected ? "border-primary bg-primary/10" : "border-border hover:bg-secondary/50"
                    }`}
                  >
                    <input
                      type="checkbox"
                      checked={isSelected}
                      onChange={() => toggleServiceSelection(service.id)}
                      className="w-5 h-5 accent-primary mt-0.5"
                    />
                    <span className="flex-1">
                      <span className="block text-foreground">{service.title}</span>
                      <span className="block text-sm text-primary">{service.price.toLocaleString()} ₽</span>
                    </span>
                  </label>
                );
              })}
            </div>
            <p className="text-xs text-muted-foreground">Список берётся из текущего справочника услуг. При сохранении заменяются услуги только в этом заказе.</p>
            <div className="flex gap-2">
              <button className="px-4 py-2 bg-primary text-primary-foreground rounded-lg">Сохранить</button>
              <button type="button" onClick={() => setEditingSection(null)} className="px-4 py-2 bg-secondary text-secondary-foreground rounded-lg">Отмена</button>
            </div>
          </form>
        ) : (
        <div className="space-y-2">
          {order.services.map((service, index) => (
            <div key={index} className="flex justify-between py-2 border-b border-border last:border-0">
              <span className="text-foreground">{service.name}</span>
              <span className="text-primary">{service.price.toLocaleString()} ₽</span>
            </div>
          ))}
        </div>
        )}
      </div>

      {/* Products */}
      <div className="bg-card border border-border rounded-lg p-6">
        <div className="flex items-center justify-between mb-4">
          <h3 className="text-lg text-foreground">Товары</h3>
          <button
            onClick={() => handleEditClick("products")}
            className="text-sm text-primary hover:underline"
          >
            Изменить
          </button>
        </div>
        {editingSection === "products" ? (
          <form onSubmit={handleProductsSubmit} className="space-y-3">
            {catalogError && (
              <p className="text-sm text-destructive">{catalogError}</p>
            )}
            <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
              {availableProducts.map((product) => {
                const isSelected = selectedProductIds.includes(product.id);

                return (
                  <label
                    key={product.id}
                    className={`flex items-start gap-3 border rounded-lg p-4 cursor-pointer transition-colors ${
                      isSelected ? "border-primary bg-primary/10" : "border-border hover:bg-secondary/50"
                    }`}
                  >
                    <input
                      type="checkbox"
                      checked={isSelected}
                      onChange={() => toggleProductSelection(product.id)}
                      className="w-5 h-5 accent-primary mt-0.5"
                    />
                    <span className="flex-1">
                      <span className="block text-foreground">{product.title}</span>
                      <span className="block text-sm text-primary">{product.price.toLocaleString()} ₽</span>
                    </span>
                  </label>
                );
              })}
            </div>
            <p className="text-xs text-muted-foreground">Список берётся из каталога товаров. Можно оставить без выбранных товаров.</p>
            <div className="flex gap-2">
              <button className="px-4 py-2 bg-primary text-primary-foreground rounded-lg">Сохранить</button>
              <button type="button" onClick={() => setEditingSection(null)} className="px-4 py-2 bg-secondary text-secondary-foreground rounded-lg">Отмена</button>
            </div>
          </form>
        ) : (
        <div className="space-y-2">
          {order.products.map((product, index) => (
            <div key={index} className="flex justify-between py-2 border-b border-border last:border-0">
              <span className="text-foreground">{product.name}</span>
              <span className="text-primary">{product.price.toLocaleString()} ₽</span>
            </div>
          ))}
        </div>
        )}
      </div>

      {/* Total */}
      <div className="bg-secondary/50 border border-border rounded-lg p-6">
        <div className="flex justify-between items-center">
          <span className="text-xl text-foreground">Итого</span>
          <span className="text-3xl text-primary">{order.total.toLocaleString()} ₽</span>
        </div>
        {editingSection === "discount" ? (
          <form onSubmit={handleDiscountSubmit} className="mt-4 grid grid-cols-1 md:grid-cols-[1fr_2fr_auto_auto] gap-2">
            <input
              name="discountAmount"
              type="number"
              min="0"
              step="1"
              required
              className="px-4 py-2 bg-input-background border border-border rounded-lg focus:outline-none focus:ring-2 focus:ring-ring"
              placeholder="Скидка"
            />
            <input
              name="reason"
              className="px-4 py-2 bg-input-background border border-border rounded-lg focus:outline-none focus:ring-2 focus:ring-ring"
              placeholder="Причина скидки"
            />
            <button className="px-4 py-2 bg-primary text-primary-foreground rounded-lg">Применить</button>
            <button type="button" onClick={() => setEditingSection(null)} className="px-4 py-2 bg-secondary text-secondary-foreground rounded-lg">Отмена</button>
          </form>
        ) : (
          <button
            onClick={() => handleEditClick("discount")}
            className="mt-4 text-sm text-primary hover:underline"
          >
            Добавить скидку
          </button>
        )}
      </div>

      {/* Documents */}
      <div className="bg-card border border-border rounded-lg p-6">
        <div className="flex items-center justify-between mb-4">
          <h3 className="text-lg text-foreground">Документы</h3>
          <button
            onClick={() => setShowAddDocument(!showAddDocument)}
            className="px-4 py-2 bg-primary text-primary-foreground rounded-lg hover:opacity-90 transition-opacity text-sm"
          >
            {showAddDocument ? "Отмена" : "Добавить документ"}
          </button>
        </div>

        {/* Add Document Form */}
        {showAddDocument && (
          <div className="mb-6 p-4 bg-secondary/30 border border-border rounded-lg">
            <div className="space-y-4">
              <div>
                <label className="block text-sm text-foreground mb-2">
                  Название документа *
                </label>
                <input
                  type="text"
                  value={newDocName}
                  onChange={(e) => setNewDocName(e.target.value)}
                  className="w-full px-4 py-2 bg-input-background border border-border rounded-lg focus:outline-none focus:ring-2 focus:ring-ring"
                  placeholder="Например: Свидетельство о смерти"
                />
              </div>
              <div>
                <label className="block text-sm text-foreground mb-2">Тип</label>
                <input
                  type="text"
                  value={newDocType}
                  onChange={(e) => setNewDocType(e.target.value)}
                  className="w-full px-4 py-2 bg-input-background border border-border rounded-lg focus:outline-none focus:ring-2 focus:ring-ring"
                  placeholder="PDF"
                />
              </div>
              <div>
                <label className="block text-sm text-foreground mb-2">
                  Файл *
                </label>
                <select
                  value={newDocFile}
                  onChange={(e) => setNewDocFile(e.target.value)}
                  className="w-full px-4 py-2 bg-input-background border border-border rounded-lg focus:outline-none focus:ring-2 focus:ring-ring"
                >
                  <option value="">Выберите файл</option>
                  {predefinedFiles.map((file) => (
                    <option key={file.value} value={file.value}>
                      {file.label}
                    </option>
                  ))}
                </select>
              </div>
              <button
                onClick={handleAddDocument}
                disabled={!newDocName.trim() || !newDocFile}
                className="px-6 py-2 bg-primary text-primary-foreground rounded-lg hover:opacity-90 transition-opacity disabled:opacity-50 disabled:cursor-not-allowed"
              >
                Добавить
              </button>
            </div>
          </div>
        )}

        {/* Documents List */}
        {order.documents.length === 0 ? (
          <p className="text-muted-foreground text-sm">Документы не добавлены</p>
        ) : (
          <div className="space-y-3">
            {order.documents.map((document) => (
              <div
                key={document.id}
                className="flex items-center justify-between p-4 bg-background border border-border rounded-lg"
              >
                <div className="flex items-center gap-4">
                  <div className="w-10 h-10 bg-primary/10 rounded-lg flex items-center justify-center flex-shrink-0">
                    <svg
                      className="w-5 h-5 text-primary"
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
                    <h4 className="text-foreground text-sm">{document.name}</h4>
                    <div className="flex items-center gap-2 text-xs text-muted-foreground mt-1">
                      <span>{document.type}</span>
                      <span>•</span>
                      <span>{document.size}</span>
                      <span>•</span>
                      <span>{document.date}</span>
                    </div>
                  </div>
                </div>
                <button
                  onClick={() => handleRemoveDocument(document.id)}
                  className="px-4 py-2 text-sm text-destructive hover:bg-destructive/10 rounded-lg transition-colors"
                >
                  Удалить
                </button>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}

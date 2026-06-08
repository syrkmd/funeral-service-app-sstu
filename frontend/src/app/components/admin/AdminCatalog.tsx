import { FormEvent, useEffect, useMemo, useState } from "react";
import {
  archiveCatalogProduct,
  archiveFuneralService,
  createCatalogProduct,
  createFuneralService,
  getCatalogCategories,
  getCatalogProducts,
  getFuneralServices,
  getServiceCategories,
  updateCatalogProduct,
  updateFuneralService,
  type CatalogProductDto,
  type FuneralServiceDto,
  type ProductCategoryDto,
  type ServiceCategoryDto,
} from "../../../api/catalog.api";
import { CatalogItemImage } from "../CatalogItemImage";

type CatalogTab = "services" | "products";

type FormState = {
  id?: number;
  title: string;
  description: string;
  price: string;
  imageUrl: string;
  categoryId: string;
  active: boolean;
  sortOrder: string;
};

const emptyForm: FormState = {
  title: "",
  description: "",
  price: "",
  imageUrl: "",
  categoryId: "",
  active: true,
  sortOrder: "",
};

export function AdminCatalog() {
  const [activeTab, setActiveTab] = useState<CatalogTab>("services");
  const [services, setServices] = useState<FuneralServiceDto[]>([]);
  const [products, setProducts] = useState<CatalogProductDto[]>([]);
  const [serviceCategories, setServiceCategories] = useState<ServiceCategoryDto[]>([]);
  const [productCategories, setProductCategories] = useState<ProductCategoryDto[]>([]);
  const [form, setForm] = useState<FormState>(emptyForm);
  const [isLoading, setIsLoading] = useState(true);
  const [isSaving, setIsSaving] = useState(false);
  const [error, setError] = useState("");
  const [message, setMessage] = useState("");

  const isProductsTab = activeTab === "products";
  const categories = isProductsTab ? productCategories : serviceCategories;
  const items = useMemo(
    () => (isProductsTab ? products : services),
    [isProductsTab, products, services],
  );

  useEffect(() => {
    loadCatalog();
  }, []);

  useEffect(() => {
    setForm(emptyForm);
    setMessage("");
    setError("");
  }, [activeTab]);

  const loadCatalog = async () => {
    setIsLoading(true);
    try {
      const [servicesResponse, productsResponse, serviceCategoriesResponse, productCategoriesResponse] =
        await Promise.all([
          getFuneralServices({ includeInactive: true }),
          getCatalogProducts({ includeInactive: true }),
          getServiceCategories(),
          getCatalogCategories(),
        ]);

      setServices(servicesResponse);
      setProducts(productsResponse);
      setServiceCategories(serviceCategoriesResponse);
      setProductCategories(productCategoriesResponse);
      setError("");
    } catch (loadError) {
      setError(loadError instanceof Error ? loadError.message : "Не удалось загрузить каталог");
    } finally {
      setIsLoading(false);
    }
  };

  const startEdit = (item: FuneralServiceDto | CatalogProductDto) => {
    setForm({
      id: item.id,
      title: item.title,
      description: item.description || "",
      price: String(item.price ?? ""),
      imageUrl: "imageUrl" in item ? item.imageUrl || "" : "",
      categoryId: item.category?.id ? String(item.category.id) : "",
      active: item.active ?? true,
      sortOrder: item.sortOrder == null ? "" : String(item.sortOrder),
    });
    setMessage("");
    setError("");
  };

  const resetForm = () => {
    setForm(emptyForm);
    setMessage("");
    setError("");
  };

  const buildPayload = () => ({
    title: form.title.trim(),
    description: form.description.trim(),
    price: Number(form.price),
    imageUrl: form.imageUrl.trim(),
    categoryId: Number(form.categoryId),
    active: form.active,
    sortOrder: form.sortOrder.trim() ? Number(form.sortOrder) : null,
  });

  const handleSubmit = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    setIsSaving(true);
    setError("");
    setMessage("");

    try {
      const payload = buildPayload();

      if (isProductsTab) {
        if (form.id) {
          await updateCatalogProduct(form.id, payload);
        } else {
          await createCatalogProduct(payload);
        }
      } else {
        const { imageUrl, ...servicePayload } = payload;
        if (form.id) {
          await updateFuneralService(form.id, servicePayload);
        } else {
          await createFuneralService(servicePayload);
        }
      }

      await loadCatalog();
      setForm(emptyForm);
      setMessage(form.id ? "Изменения сохранены" : "Элемент создан");
    } catch (saveError) {
      setError(saveError instanceof Error ? saveError.message : "Не удалось сохранить изменения");
    } finally {
      setIsSaving(false);
    }
  };

  const handleArchiveToggle = async (item: FuneralServiceDto | CatalogProductDto) => {
    setIsSaving(true);
    setError("");
    setMessage("");

    try {
      const isActive = item.active ?? true;

      if (isProductsTab) {
        if (isActive) {
          await archiveCatalogProduct(item.id);
        } else {
          await updateCatalogProduct(item.id, { active: true });
        }
      } else if (isActive) {
        await archiveFuneralService(item.id);
      } else {
        await updateFuneralService(item.id, { active: true });
      }

      await loadCatalog();
      setMessage(isActive ? "Элемент отправлен в архив" : "Элемент восстановлен");
    } catch (toggleError) {
      setError(toggleError instanceof Error ? toggleError.message : "Не удалось изменить статус");
    } finally {
      setIsSaving(false);
    }
  };

  return (
    <div className="space-y-6">
      <div className="flex flex-col gap-4 md:flex-row md:items-center md:justify-between">
        <div>
          <h1 className="text-2xl text-foreground">Каталог</h1>
          <p className="text-sm text-muted-foreground">
            Управление услугами и товарами, которые доступны в заказах.
          </p>
        </div>

        <div className="inline-flex rounded-lg border border-border bg-card p-1">
          <button
            type="button"
            onClick={() => setActiveTab("services")}
            className={`px-4 py-2 text-sm rounded-md transition-colors ${
              activeTab === "services"
                ? "bg-primary text-primary-foreground"
                : "text-muted-foreground hover:text-foreground"
            }`}
          >
            Услуги
          </button>
          <button
            type="button"
            onClick={() => setActiveTab("products")}
            className={`px-4 py-2 text-sm rounded-md transition-colors ${
              activeTab === "products"
                ? "bg-primary text-primary-foreground"
                : "text-muted-foreground hover:text-foreground"
            }`}
          >
            Товары
          </button>
        </div>
      </div>

      {(error || message) && (
        <div
          className={`rounded-lg border p-4 text-sm ${
            error
              ? "border-destructive/30 bg-destructive/10 text-destructive"
              : "border-primary/30 bg-primary/10 text-primary"
          }`}
        >
          {error || message}
        </div>
      )}

      <form onSubmit={handleSubmit} className="bg-card border border-border rounded-lg p-6 space-y-4">
        <div className="flex items-center justify-between">
          <h2 className="text-lg text-foreground">
            {form.id ? "Редактирование" : isProductsTab ? "Новый товар" : "Новая услуга"}
          </h2>
          {form.id && (
            <button
              type="button"
              onClick={resetForm}
              className="text-sm text-muted-foreground hover:text-foreground"
            >
              Создать новый
            </button>
          )}
        </div>

        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <label className="space-y-2">
            <span className="text-sm text-muted-foreground">Название</span>
            <input
              value={form.title}
              onChange={(event) => setForm((prev) => ({ ...prev, title: event.target.value }))}
              required
              className="w-full px-4 py-2 bg-input-background border border-border rounded-lg focus:outline-none focus:ring-2 focus:ring-ring"
            />
          </label>

          <label className="space-y-2">
            <span className="text-sm text-muted-foreground">Категория</span>
            <select
              value={form.categoryId}
              onChange={(event) => setForm((prev) => ({ ...prev, categoryId: event.target.value }))}
              required
              className="w-full px-4 py-2 bg-input-background border border-border rounded-lg focus:outline-none focus:ring-2 focus:ring-ring"
            >
              <option value="">Выберите категорию</option>
              {categories.map((category) => (
                <option key={category.id} value={category.id}>
                  {category.name}
                </option>
              ))}
            </select>
          </label>

          <label className="space-y-2">
            <span className="text-sm text-muted-foreground">Цена</span>
            <input
              type="number"
              min="0"
              value={form.price}
              onChange={(event) => setForm((prev) => ({ ...prev, price: event.target.value }))}
              required
              className="w-full px-4 py-2 bg-input-background border border-border rounded-lg focus:outline-none focus:ring-2 focus:ring-ring"
            />
          </label>

          <label className="space-y-2">
            <span className="text-sm text-muted-foreground">Порядок</span>
            <input
              type="number"
              value={form.sortOrder}
              onChange={(event) => setForm((prev) => ({ ...prev, sortOrder: event.target.value }))}
              className="w-full px-4 py-2 bg-input-background border border-border rounded-lg focus:outline-none focus:ring-2 focus:ring-ring"
            />
          </label>

          {isProductsTab && (
            <label className="space-y-2 md:col-span-2">
              <span className="text-sm text-muted-foreground">Изображение</span>
              <input
                value={form.imageUrl}
                onChange={(event) => setForm((prev) => ({ ...prev, imageUrl: event.target.value }))}
                className="w-full px-4 py-2 bg-input-background border border-border rounded-lg focus:outline-none focus:ring-2 focus:ring-ring"
              />
            </label>
          )}

          <label className="space-y-2 md:col-span-2">
            <span className="text-sm text-muted-foreground">Описание</span>
            <textarea
              value={form.description}
              onChange={(event) => setForm((prev) => ({ ...prev, description: event.target.value }))}
              rows={3}
              className="w-full px-4 py-2 bg-input-background border border-border rounded-lg focus:outline-none focus:ring-2 focus:ring-ring"
            />
          </label>
        </div>

        <label className="flex items-center gap-3 text-sm text-foreground">
          <input
            type="checkbox"
            checked={form.active}
            onChange={(event) => setForm((prev) => ({ ...prev, active: event.target.checked }))}
            className="w-4 h-4"
          />
          Показывать в каталоге
        </label>

        <div className="flex gap-3">
          <button
            type="submit"
            disabled={isSaving}
            className="px-4 py-2 rounded-lg bg-primary text-primary-foreground disabled:opacity-60"
          >
            {isSaving ? "Сохранение..." : "Сохранить"}
          </button>
          <button
            type="button"
            onClick={resetForm}
            className="px-4 py-2 rounded-lg bg-secondary text-secondary-foreground"
          >
            Очистить
          </button>
        </div>
      </form>

      <div className="bg-card border border-border rounded-lg overflow-hidden">
        <div className="px-6 py-4 border-b border-border flex items-center justify-between">
          <h2 className="text-lg text-foreground">{isProductsTab ? "Товары" : "Услуги"}</h2>
          <span className="text-sm text-muted-foreground">{items.length} элементов</span>
        </div>

        {isLoading ? (
          <div className="p-8 text-center text-muted-foreground">Загрузка каталога...</div>
        ) : (
          <table className="w-full">
            <thead className="bg-secondary/50 border-b border-border">
              <tr>
                <th className="px-6 py-3 text-left text-xs text-muted-foreground">Название</th>
                <th className="px-6 py-3 text-left text-xs text-muted-foreground">Категория</th>
                <th className="px-6 py-3 text-left text-xs text-muted-foreground">Цена</th>
                <th className="px-6 py-3 text-left text-xs text-muted-foreground">Порядок</th>
                <th className="px-6 py-3 text-left text-xs text-muted-foreground">Статус</th>
                <th className="px-6 py-3 text-left text-xs text-muted-foreground">Действия</th>
              </tr>
            </thead>
            <tbody>
              {items.map((item) => {
                const isActive = item.active ?? true;

                return (
                  <tr
                    key={item.id}
                    className="border-b border-border last:border-0 hover:bg-secondary/30 transition-colors"
                  >
                    <td className="px-6 py-4">
                      <div className="flex items-center gap-3">
                        <CatalogItemImage
                          kind={isProductsTab ? "product" : "service"}
                          title={item.title}
                          className="w-12 h-12 object-cover rounded border border-border bg-muted flex-shrink-0"
                        />
                        <div className="min-w-0">
                          <div className="text-sm text-foreground">{item.title}</div>
                          <div className="text-xs text-muted-foreground line-clamp-1">
                            {item.description || "Без описания"}
                          </div>
                        </div>
                      </div>
                    </td>
                    <td className="px-6 py-4 text-sm text-muted-foreground">
                      {item.category?.name || "Без категории"}
                    </td>
                    <td className="px-6 py-4 text-sm text-primary">
                      {Number(item.price).toLocaleString()} ₽
                    </td>
                    <td className="px-6 py-4 text-sm text-muted-foreground">
                      {item.sortOrder ?? "—"}
                    </td>
                    <td className="px-6 py-4">
                      <span
                        className={`px-3 py-1 rounded-full text-xs border ${
                          isActive
                            ? "bg-primary/10 text-primary border-primary/30"
                            : "bg-secondary text-muted-foreground border-border"
                        }`}
                      >
                        {isActive ? "Активен" : "Архив"}
                      </span>
                    </td>
                    <td className="px-6 py-4">
                      <div className="flex gap-3">
                        <button
                          type="button"
                          onClick={() => startEdit(item)}
                          className="text-sm text-primary hover:underline"
                        >
                          Изменить
                        </button>
                        <button
                          type="button"
                          disabled={isSaving}
                          onClick={() => handleArchiveToggle(item)}
                          className="text-sm text-muted-foreground hover:text-foreground disabled:opacity-60"
                        >
                          {isActive ? "В архив" : "Вернуть"}
                        </button>
                      </div>
                    </td>
                  </tr>
                );
              })}
            </tbody>
          </table>
        )}
      </div>
    </div>
  );
}

import { apiClient } from "./client";

export type ProductCategoryDto = {
  id: number;
  name: string;
};

export type ServiceCategoryDto = {
  id: number;
  name: string;
};

export type CatalogProductDto = {
  id: number;
  title: string;
  description: string;
  price: number;
  imageUrl?: string | null;
  category: ProductCategoryDto | null;
  active?: boolean;
  sortOrder?: number | null;
};

export type FuneralServiceDto = {
  id: number;
  title: string;
  description: string;
  price: number;
  category: ServiceCategoryDto | null;
  active?: boolean;
  sortOrder?: number | null;
};

export type CatalogProductRequest = {
  title: string;
  description?: string;
  price: number;
  imageUrl?: string;
  categoryId: number;
  active?: boolean;
  sortOrder?: number | null;
};

export type FuneralServiceRequest = {
  title: string;
  description?: string;
  price: number;
  categoryId: number;
  active?: boolean;
  sortOrder?: number | null;
};

export async function getCatalogProducts(options?: { includeInactive?: boolean; categoryId?: number }) {
  const response = await apiClient.get<CatalogProductDto[]>("/catalog/products", {
    params: {
      includeInactive: options?.includeInactive || undefined,
      categoryId: options?.categoryId,
    },
  });
  return response.data;
}

export async function createCatalogProduct(request: CatalogProductRequest) {
  const response = await apiClient.post<CatalogProductDto>("/catalog/products", request);
  return response.data;
}

export async function updateCatalogProduct(id: number, request: Partial<CatalogProductRequest>) {
  const response = await apiClient.patch<CatalogProductDto>(`/catalog/products/${id}`, request);
  return response.data;
}

export async function archiveCatalogProduct(id: number) {
  await apiClient.delete(`/catalog/products/${id}`);
}

export async function getFuneralServices(options?: { includeInactive?: boolean }) {
  const response = await apiClient.get<FuneralServiceDto[]>("/funeral-services", {
    params: {
      includeInactive: options?.includeInactive || undefined,
    },
  });
  return response.data;
}

export async function createFuneralService(request: FuneralServiceRequest) {
  const response = await apiClient.post<FuneralServiceDto>("/funeral-services", request);
  return response.data;
}

export async function updateFuneralService(id: number, request: Partial<FuneralServiceRequest>) {
  const response = await apiClient.patch<FuneralServiceDto>(`/funeral-services/${id}`, request);
  return response.data;
}

export async function archiveFuneralService(id: number) {
  await apiClient.delete(`/funeral-services/${id}`);
}

export async function getServiceCategories() {
  const response = await apiClient.get<ServiceCategoryDto[]>("/service-categories");
  return response.data;
}

export async function getCatalogCategories() {
  const response = await apiClient.get<ProductCategoryDto[]>("/catalog/categories");
  return response.data;
}

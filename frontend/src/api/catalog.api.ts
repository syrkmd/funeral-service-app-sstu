import { apiClient } from "./client";

export type ProductCategoryDto = {
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
};

export async function getCatalogProducts() {
  const response = await apiClient.get<CatalogProductDto[]>("/catalog/products");
  return response.data;
}

export async function getCatalogCategories() {
  const response = await apiClient.get<ProductCategoryDto[]>("/catalog/categories");
  return response.data;
}

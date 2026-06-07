import { apiClient } from "./client";

export type CemeteryPlot = {
  id: number;
  label: string;
  available: boolean;
};

export type CemeterySection = {
  id: number;
  name: string;
  description: string;
};

export async function getCemeterySections() {
  const response = await apiClient.get<CemeterySection[]>("/cemetery/sections");
  return response.data;
}

export async function getCemeteryPlots(sectionName: string) {
  const response = await apiClient.get<CemeteryPlot[]>("/cemetery/plots", {
    params: { sectionName },
  });
  return response.data;
}

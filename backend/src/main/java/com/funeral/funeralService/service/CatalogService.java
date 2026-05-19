package com.funeral.funeralService.service;

import com.funeral.funeralService.dto.CatalogProductDto;
import com.funeral.funeralService.dto.ProductCategoryDto;
import com.funeral.funeralService.entity.CatalogProduct;
import com.funeral.funeralService.entity.ProductCategory;
import com.funeral.funeralService.repository.CatalogProductRepository;
import com.funeral.funeralService.repository.ProductCategoryRepository;
import lombok.RequiredArgsConstructor;
import org.springframework.stereotype.Service;

import java.util.List;

@Service
@RequiredArgsConstructor
public class CatalogService {

    private final CatalogProductRepository catalogProductRepository;
    private final ProductCategoryRepository productCategoryRepository;

    public List<CatalogProductDto> getProducts(Long categoryId) {
        List<CatalogProduct> products;

        if (categoryId != null) {
            products = catalogProductRepository.findByCategoryIdAndActiveTrueOrderBySortOrderAsc(categoryId);
        } else {
            products = catalogProductRepository.findByActiveTrueOrderBySortOrderAsc();
        }

        return products.stream()
                .map(this::toProductDto)
                .toList();
    }

    public List<ProductCategoryDto> getCategories() {
        return productCategoryRepository.findByActiveTrueOrderBySortOrderAsc()
                .stream()
                .map(this::toCategoryDto)
                .toList();
    }


    private CatalogProductDto toProductDto(CatalogProduct product) {
        return new CatalogProductDto(
                product.getId(),
                product.getTitle(),
                product.getDescription(),
                product.getPrice(),
                product.getImageUrl(),
                toCategoryDto(product.getCategory())
        );
    }

    private ProductCategoryDto toCategoryDto(ProductCategory category) {
        if (category == null) {
            return null;
        }

        return new ProductCategoryDto(
                category.getId(),
                category.getName()
        );
    }
}

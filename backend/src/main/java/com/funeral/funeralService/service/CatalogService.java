package com.funeral.funeralService.service;

import com.funeral.funeralService.dto.catalog.request.CreateCatalogProductRequest;
import com.funeral.funeralService.dto.catalog.request.UpdateCatalogProductRequest;
import com.funeral.funeralService.dto.catalog.response.CatalogProductDto;
import com.funeral.funeralService.dto.catalog.response.ProductCategoryDto;
import com.funeral.funeralService.entity.CatalogProduct;
import com.funeral.funeralService.entity.ProductCategory;
import com.funeral.funeralService.exception.CatalogProductNotFoundException;
import com.funeral.funeralService.exception.ProductCategoryNotFoundException;
import com.funeral.funeralService.repository.CatalogProductRepository;
import com.funeral.funeralService.repository.ProductCategoryRepository;
import lombok.RequiredArgsConstructor;
import org.springframework.stereotype.Service;

import java.util.List;

import static com.funeral.funeralService.util.PatchUtils.applyIfPresent;

@Service
@RequiredArgsConstructor
public class CatalogService {

    private final CatalogProductRepository catalogProductRepository;
    private final ProductCategoryRepository productCategoryRepository;

    public List<CatalogProductDto> getProducts(Long categoryId, boolean includeInactive) {
        List<CatalogProduct> products;

        if (categoryId != null) {
            products = includeInactive
                    ? catalogProductRepository.findByCategoryIdOrderBySortOrderAsc(categoryId)
                    : catalogProductRepository.findByCategoryIdAndActiveTrueOrderBySortOrderAsc(categoryId);
        } else {
            products = includeInactive
                    ? catalogProductRepository.findAllByOrderBySortOrderAsc()
                    : catalogProductRepository.findByActiveTrueOrderBySortOrderAsc();
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

    public CatalogProductDto createProduct(CreateCatalogProductRequest request) {
        CatalogProduct product = new CatalogProduct();
        ProductCategory category = findCategory(request.getCategoryId());

        product.setTitle(request.getTitle());
        product.setDescription(request.getDescription());
        product.setPrice(request.getPrice());
        product.setImageUrl(request.getImageUrl());
        product.setCategory(category);
        product.setActive(request.getActive() != null ? request.getActive() : true);
        product.setSortOrder(request.getSortOrder());

        return toProductDto(catalogProductRepository.save(product));
    }

    public CatalogProductDto updateProduct(Long id, UpdateCatalogProductRequest request) {
        CatalogProduct product = findProduct(id);

        applyIfPresent(request.getTitle(), product::setTitle);
        applyIfPresent(request.getDescription(), product::setDescription);
        applyIfPresent(request.getPrice(), product::setPrice);
        applyIfPresent(request.getImageUrl(), product::setImageUrl);
        applyIfPresent(request.getCategoryId(), categoryId -> product.setCategory(findCategory(categoryId)));
        applyIfPresent(request.getActive(), product::setActive);
        applyIfPresent(request.getSortOrder(), product::setSortOrder);

        return toProductDto(catalogProductRepository.save(product));
    }

    public void archiveProduct(Long id) {
        CatalogProduct product = findProduct(id);
        product.setActive(false);
        catalogProductRepository.save(product);
    }

    private CatalogProductDto toProductDto(CatalogProduct product) {
        return new CatalogProductDto(
                product.getId(),
                product.getTitle(),
                product.getDescription(),
                product.getPrice(),
                product.getImageUrl(),
                toCategoryDto(product.getCategory()),
                product.getActive(),
                product.getSortOrder()
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

    private CatalogProduct findProduct(Long id) {
        return catalogProductRepository.findById(id)
                .orElseThrow(() -> new CatalogProductNotFoundException(id));
    }

    private ProductCategory findCategory(Long id) {
        return productCategoryRepository.findById(id)
                .orElseThrow(() -> new ProductCategoryNotFoundException(id));
    }
}

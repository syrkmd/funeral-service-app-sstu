package com.funeral.funeralService.controller;

import com.funeral.funeralService.dto.catalog.request.CreateCatalogProductRequest;
import com.funeral.funeralService.dto.catalog.request.UpdateCatalogProductRequest;
import com.funeral.funeralService.dto.catalog.response.CatalogProductDto;
import com.funeral.funeralService.dto.catalog.response.ProductCategoryDto;
import com.funeral.funeralService.service.CatalogService;
import jakarta.validation.Valid;
import lombok.RequiredArgsConstructor;
import org.springframework.http.HttpStatus;
import org.springframework.web.bind.annotation.*;

import java.util.List;

@RestController
@RequestMapping("/catalog")
@RequiredArgsConstructor
public class CatalogController {

    private final CatalogService service;

    @GetMapping("/products")
    public List<CatalogProductDto> getProducts(
            @RequestParam(required = false) Long categoryId,
            @RequestParam(defaultValue = "false") boolean includeInactive
    ) {
        return service.getProducts(categoryId, includeInactive);
    }

    @PostMapping("/products")
    @ResponseStatus(HttpStatus.CREATED)
    public CatalogProductDto createProduct(@Valid @RequestBody CreateCatalogProductRequest request) {
        return service.createProduct(request);
    }

    @PatchMapping("/products/{id}")
    public CatalogProductDto updateProduct(@PathVariable Long id, @Valid @RequestBody UpdateCatalogProductRequest request) {
        return service.updateProduct(id, request);
    }

    @DeleteMapping("/products/{id}")
    @ResponseStatus(HttpStatus.NO_CONTENT)
    public void archiveProduct(@PathVariable Long id) {
        service.archiveProduct(id);
    }

    @GetMapping("/categories")
    public List<ProductCategoryDto> getCategories() {
        return service.getCategories();
    }
}

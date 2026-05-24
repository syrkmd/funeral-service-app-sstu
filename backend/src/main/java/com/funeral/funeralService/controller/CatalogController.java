package com.funeral.funeralService.controller;

import com.funeral.funeralService.dto.catalog.request.CreateCatalogProductRequest;
import com.funeral.funeralService.dto.catalog.request.UpdateCatalogProductRequest;
import com.funeral.funeralService.dto.catalog.response.CatalogProductDto;
import com.funeral.funeralService.dto.catalog.response.ProductCategoryDto;
import com.funeral.funeralService.service.CatalogService;
import io.swagger.v3.oas.annotations.Operation;
import io.swagger.v3.oas.annotations.Parameter;
import io.swagger.v3.oas.annotations.responses.ApiResponse;
import io.swagger.v3.oas.annotations.tags.Tag;
import jakarta.validation.Valid;
import lombok.RequiredArgsConstructor;
import org.springframework.http.HttpStatus;
import org.springframework.web.bind.annotation.*;

import java.util.List;

@RestController
@RequestMapping("/catalog")
@RequiredArgsConstructor
@Tag(name = "Каталог товаров", description = "Каталог товаров для публичной витрины и админского управления")
public class CatalogController {

    private final CatalogService service;

    @GetMapping("/products")
    @Operation(summary = "Получить товары каталога", description = "Возвращает активные товары для публичного каталога. Админка может передать includeInactive=true, чтобы увидеть архивные товары.")
    @ApiResponse(responseCode = "200", description = "Список товаров возвращён")
    public List<CatalogProductDto> getProducts(
            @Parameter(description = "Необязательный фильтр по id категории товара", example = "1")
            @RequestParam(required = false) Long categoryId,
            @Parameter(description = "Админский флаг для отображения архивных товаров", example = "true")
            @RequestParam(defaultValue = "false") boolean includeInactive
    ) {
        return service.getProducts(categoryId, includeInactive);
    }

    @PostMapping("/products")
    @ResponseStatus(HttpStatus.CREATED)
    @Operation(summary = "Создать товар каталога", description = "Админский endpoint для создания нового товара, доступного при оформлении заказа.")
    @ApiResponse(responseCode = "201", description = "Товар создан")
    @ApiResponse(responseCode = "400", description = "Ошибка валидации")
    @ApiResponse(responseCode = "404", description = "Категория товара не найдена")
    public CatalogProductDto createProduct(@Valid @RequestBody CreateCatalogProductRequest request) {
        return service.createProduct(request);
    }

    @PatchMapping("/products/{id}")
    @Operation(summary = "Изменить товар каталога", description = "Админский endpoint для частичного изменения товара. Также позволяет восстановить архивный товар через active=true.")
    @ApiResponse(responseCode = "200", description = "Товар изменён")
    @ApiResponse(responseCode = "400", description = "Ошибка валидации")
    @ApiResponse(responseCode = "404", description = "Товар или категория не найдены")
    public CatalogProductDto updateProduct(
            @Parameter(description = "Id товара", example = "8") @PathVariable Long id,
            @Valid @RequestBody UpdateCatalogProductRequest request
    ) {
        return service.updateProduct(id, request);
    }

    @DeleteMapping("/products/{id}")
    @ResponseStatus(HttpStatus.NO_CONTENT)
    @Operation(summary = "Архивировать товар каталога", description = "Админский soft delete endpoint. Товар не удаляется физически, а переводится в active=false.")
    @ApiResponse(responseCode = "204", description = "Товар архивирован")
    @ApiResponse(responseCode = "404", description = "Товар не найден")
    public void archiveProduct(@Parameter(description = "Id товара", example = "8") @PathVariable Long id) {
        service.archiveProduct(id);
    }

    @GetMapping("/categories")
    @Operation(summary = "Получить категории товаров", description = "Возвращает активные категории, которые используются для товаров каталога.")
    @ApiResponse(responseCode = "200", description = "Категории товаров возвращены")
    public List<ProductCategoryDto> getCategories() {
        return service.getCategories();
    }
}

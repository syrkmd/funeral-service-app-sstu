package com.funeral.funeralService.controller;

import com.funeral.funeralService.dto.CatalogProductDto;
import com.funeral.funeralService.dto.ProductCategoryDto;
import com.funeral.funeralService.service.CatalogService;
import lombok.RequiredArgsConstructor;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RequestParam;
import org.springframework.web.bind.annotation.RestController;

import java.util.List;

@RestController
@RequestMapping("/catalog")
@RequiredArgsConstructor
public class CatalogController {

    private final CatalogService service;

    @GetMapping("/products")
    public ResponseEntity<List<CatalogProductDto>> getProducts(@RequestParam(required = false) Long categoryId) {
        return ResponseEntity.ok(service.getProducts(categoryId));
    }

    @GetMapping("/categories")
    public ResponseEntity<List<ProductCategoryDto>> getCategories() {
        return ResponseEntity.ok(service.getCategories());
    }
}

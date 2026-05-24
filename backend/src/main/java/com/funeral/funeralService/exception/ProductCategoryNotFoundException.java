package com.funeral.funeralService.exception;

public class ProductCategoryNotFoundException extends RuntimeException {
    public ProductCategoryNotFoundException(Long id) {
        super("Product category not found: " + id);
    }
}

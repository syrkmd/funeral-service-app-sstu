package com.funeral.funeralService.exception;

public class CatalogProductNotFoundException extends RuntimeException {
    public CatalogProductNotFoundException(Long id) {
        super("Catalog product not found: " + id);
    }
}

package com.funeral.funeralService.exception;

public class ServiceCategoryNotFoundException extends RuntimeException {
    public ServiceCategoryNotFoundException(Long id) {
        super("Service category not found: " + id);
    }
}

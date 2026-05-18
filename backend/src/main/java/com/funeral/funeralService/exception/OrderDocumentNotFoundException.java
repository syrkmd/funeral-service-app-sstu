package com.funeral.funeralService.exception;

public class OrderDocumentNotFoundException extends RuntimeException {
    public OrderDocumentNotFoundException(Long documentId) {
        super("Document not found: " + documentId);
    }
}

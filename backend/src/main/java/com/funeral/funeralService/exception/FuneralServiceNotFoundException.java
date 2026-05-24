package com.funeral.funeralService.exception;

public class FuneralServiceNotFoundException extends RuntimeException {
    public FuneralServiceNotFoundException(Long id) {
        super("Funeral service not found: " + id);
    }
}

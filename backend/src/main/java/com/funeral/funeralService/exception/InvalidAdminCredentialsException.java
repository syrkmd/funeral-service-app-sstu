package com.funeral.funeralService.exception;

public class InvalidAdminCredentialsException extends RuntimeException {
    public InvalidAdminCredentialsException() {
        super("Invalid admin credentials");
    }
}

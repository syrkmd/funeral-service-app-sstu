package com.funeral.funeralService.exception;

public class InvalidOrderStatusException extends RuntimeException {

    public InvalidOrderStatusException(String status) {
        super("Недопустимый статус заказа: " + status);
    }
}

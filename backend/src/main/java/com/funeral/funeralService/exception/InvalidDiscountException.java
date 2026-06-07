package com.funeral.funeralService.exception;

public class InvalidDiscountException extends RuntimeException {

    public InvalidDiscountException() {
        super("Скидка не может превышать сумму заказа");
    }
}

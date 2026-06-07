package com.funeral.funeralService.exception;

public class BankServiceUnavailableException extends RuntimeException {

    public BankServiceUnavailableException() {
        super("Сервис банка временно недоступен. Попробуйте повторить оплату позже.");
    }
}

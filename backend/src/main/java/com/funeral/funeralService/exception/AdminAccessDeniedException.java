package com.funeral.funeralService.exception;

public class AdminAccessDeniedException extends RuntimeException {

    public AdminAccessDeniedException() {
        super("Для выполнения операции требуется активная сессия администратора");
    }
}

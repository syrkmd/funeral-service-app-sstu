package com.funeral.funeralService.exception;

public class InvalidCemeteryReservationException extends RuntimeException {

    public InvalidCemeteryReservationException() {
        super("Для бронирования участка необходимо указать дату церемонии");
    }
}

package com.funeral.funeralService.dto.cemetery.response;

import lombok.AllArgsConstructor;
import lombok.Data;

@Data
@AllArgsConstructor
public class ReservePlotResponse {

    private boolean success;

    private String reservationId;
}

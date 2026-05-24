package com.funeral.funeralService.dto.order.request;

import jakarta.validation.constraints.NotBlank;
import lombok.Data;

import java.time.LocalDate;

@Data
public class OrderDocumentRequest {

    private Long id;

    @NotBlank
    private String name;

    @NotBlank
    private String type;

    private LocalDate date;

    private String size;
}

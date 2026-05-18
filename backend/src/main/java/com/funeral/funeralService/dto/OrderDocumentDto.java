package com.funeral.funeralService.dto;

import lombok.Data;

import java.time.LocalDate;

@Data
public class OrderDocumentDto {

    private Long id;

    private String name;

    private String type;

    private LocalDate date;

    private String size;
}

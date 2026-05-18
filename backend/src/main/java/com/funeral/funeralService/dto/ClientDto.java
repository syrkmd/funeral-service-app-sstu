package com.funeral.funeralService.dto;

import jakarta.validation.constraints.Email;
import jakarta.validation.constraints.NotBlank;
import lombok.Data;

@Data
public class ClientDto {

    @NotBlank
    private String name;

    @NotBlank
    private String phone;

    @Email
    private String email;
}

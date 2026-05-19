package com.funeral.funeralService.dto;

import jakarta.validation.constraints.NotBlank;
import lombok.Data;

@Data
public class AccountVerifyRequest {

    @NotBlank
    private String phone;

    @NotBlank
    private String code;
}

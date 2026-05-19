package com.funeral.funeralService.dto;

import jakarta.validation.constraints.NotBlank;
import lombok.Data;

@Data
public class AdminLoginRequest {

    @NotBlank
    private String login;

    @NotBlank
    private String password;
}

package com.funeral.funeralService.dto.admin.request;

import jakarta.validation.constraints.NotBlank;
import lombok.Data;

@Data
public class AdminLoginRequest {

    @NotBlank
    private String login;

    @NotBlank
    private String password;
}

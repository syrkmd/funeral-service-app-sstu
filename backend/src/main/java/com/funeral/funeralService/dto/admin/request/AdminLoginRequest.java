package com.funeral.funeralService.dto.admin.request;

import io.swagger.v3.oas.annotations.media.Schema;
import jakarta.validation.constraints.NotBlank;
import lombok.Data;

@Data
@Schema(description = "Запрос входа администратора")
public class AdminLoginRequest {

    @NotBlank
    @Schema(description = "Логин администратора", example = "admin")
    private String login;

    @NotBlank
    @Schema(description = "Пароль администратора", example = "admin")
    private String password;
}

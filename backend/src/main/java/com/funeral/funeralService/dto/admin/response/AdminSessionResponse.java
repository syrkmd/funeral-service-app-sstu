package com.funeral.funeralService.dto.admin.response;

import io.swagger.v3.oas.annotations.media.Schema;
import lombok.AllArgsConstructor;
import lombok.Data;

@Data
@AllArgsConstructor
@Schema(description = "Состояние сессии администратора")
public class AdminSessionResponse {

    @Schema(description = "Авторизован ли администратор в текущей сессии", example = "true")
    private Boolean isAuthenticated;
}

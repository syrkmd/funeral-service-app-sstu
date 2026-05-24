package com.funeral.funeralService.dto.account.response;

import io.swagger.v3.oas.annotations.media.Schema;
import lombok.AllArgsConstructor;
import lombok.Data;

@Data
@AllArgsConstructor
@Schema(description = "Состояние клиентской сессии")
public class AccountSessionResponse {

    @Schema(description = "Есть ли активная сессия клиента", example = "true")
    private Boolean isAuthenticated;

    @Schema(description = "Телефон, сохранённый в сессии", example = "+79991234567")
    private String phone;
}

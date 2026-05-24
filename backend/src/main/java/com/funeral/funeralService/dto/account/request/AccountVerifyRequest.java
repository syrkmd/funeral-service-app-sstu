package com.funeral.funeralService.dto.account.request;

import io.swagger.v3.oas.annotations.media.Schema;
import jakarta.validation.constraints.NotBlank;
import lombok.Data;

@Data
@Schema(description = "Запрос подтверждения аккаунта клиента")
public class AccountVerifyRequest {

    @NotBlank
    @Schema(description = "Телефон клиента", example = "+79991234567")
    private String phone;

    @NotBlank
    @Schema(description = "Демо-код подтверждения", example = "4821")
    private String code;
}

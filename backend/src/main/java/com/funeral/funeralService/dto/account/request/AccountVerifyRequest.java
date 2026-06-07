package com.funeral.funeralService.dto.account.request;

import io.swagger.v3.oas.annotations.media.Schema;
import jakarta.validation.constraints.NotBlank;
import jakarta.validation.constraints.Pattern;
import lombok.Data;

@Data
@Schema(description = "Запрос подтверждения аккаунта клиента")
public class AccountVerifyRequest {

    @NotBlank
    @Pattern(regexp = "^\\+?[0-9]{10,15}$", message = "Телефон должен содержать от 10 до 15 цифр")
    @Schema(description = "Телефон клиента", example = "+79991234567")
    private String phone;

    @NotBlank
    @Pattern(regexp = "^[0-9]{4}$", message = "Код подтверждения должен состоять из 4 цифр")
    @Schema(description = "Демо-код подтверждения", example = "4821")
    private String code;
}

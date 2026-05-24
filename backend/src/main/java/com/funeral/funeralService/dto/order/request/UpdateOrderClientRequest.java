package com.funeral.funeralService.dto.order.request;

import io.swagger.v3.oas.annotations.media.Schema;
import jakarta.validation.constraints.Email;
import jakarta.validation.constraints.NotBlank;
import lombok.Data;

@Data
@Schema(description = "Запрос изменения контактных данных клиента")
public class UpdateOrderClientRequest {

    @NotBlank
    @Schema(description = "Полное имя клиента", example = "Иван Петров")
    private String name;

    @NotBlank
    @Schema(description = "Телефон клиента", example = "+79991234567")
    private String phone;

    @Email
    @Schema(description = "Email клиента", example = "client@example.com")
    private String email;
}

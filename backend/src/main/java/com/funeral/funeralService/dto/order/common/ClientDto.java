package com.funeral.funeralService.dto.order.common;

import io.swagger.v3.oas.annotations.media.Schema;
import jakarta.validation.constraints.Email;
import jakarta.validation.constraints.NotBlank;
import jakarta.validation.constraints.Pattern;
import lombok.Data;

@Data
@Schema(description = "Контактные данные клиента")
public class ClientDto {

    @NotBlank
    @Pattern(
            regexp = "^[\\p{L}]+(?:[ '\\-][\\p{L}]+)*$",
            message = "Имя может содержать только буквы, пробелы, дефисы и апострофы"
    )
    @Schema(description = "Полное имя клиента", example = "Иван Петров")
    private String name;

    @NotBlank
    @Pattern(regexp = "^\\+?[0-9]{10,15}$", message = "Телефон должен содержать от 10 до 15 цифр")
    @Schema(description = "Телефон клиента для поиска заказов и входа в аккаунт", example = "+79991234567")
    private String phone;

    @Email
    @Schema(description = "Необязательный email клиента", example = "client@example.com")
    private String email;
}

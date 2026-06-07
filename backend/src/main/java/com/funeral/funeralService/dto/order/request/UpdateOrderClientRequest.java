package com.funeral.funeralService.dto.order.request;

import io.swagger.v3.oas.annotations.media.Schema;
import jakarta.validation.constraints.Email;
import jakarta.validation.constraints.NotBlank;
import jakarta.validation.constraints.Pattern;
import lombok.Data;

@Data
@Schema(description = "Запрос изменения контактных данных клиента")
public class UpdateOrderClientRequest {

    @NotBlank
    @Pattern(
            regexp = "^[\\p{L}]+(?:[ '\\-][\\p{L}]+)*$",
            message = "Имя может содержать только буквы, пробелы, дефисы и апострофы"
    )
    @Schema(description = "Полное имя клиента", example = "Иван Петров")
    private String name;

    @NotBlank
    @Pattern(regexp = "^\\+?[0-9]{10,15}$", message = "Телефон должен содержать от 10 до 15 цифр")
    @Schema(description = "Телефон клиента", example = "+79991234567")
    private String phone;

    @Email
    @Schema(description = "Email клиента", example = "client@example.com")
    private String email;
}

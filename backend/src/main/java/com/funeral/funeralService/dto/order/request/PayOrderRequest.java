package com.funeral.funeralService.dto.order.request;

import jakarta.validation.constraints.NotBlank;
import jakarta.validation.constraints.Pattern;
import lombok.Data;

@Data
public class PayOrderRequest {

    @NotBlank
    @Pattern(regexp = "\\d{16}", message = "Номер карты должен содержать 16 цифр")
    private String cardNumber;

    @NotBlank
    @Pattern(regexp = "\\d{3}", message = "CVV должен содержать 3 цифры")
    private String cvv;

    @NotBlank
    @Pattern(
            regexp = "(0[1-9]|1[0-2])/\\d{2}",
            message = "Срок действия должен иметь формат ММ/ГГ"
    )
    private String expiryDate;
}

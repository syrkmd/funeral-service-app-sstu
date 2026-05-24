package com.funeral.funeralService.dto.order.request;

import jakarta.validation.constraints.Email;
import jakarta.validation.constraints.NotBlank;
import lombok.Data;

@Data
public class UpdateOrderClientRequest {

    @NotBlank
    private String name;

    @NotBlank
    private String phone;

    @Email
    private String email;
}

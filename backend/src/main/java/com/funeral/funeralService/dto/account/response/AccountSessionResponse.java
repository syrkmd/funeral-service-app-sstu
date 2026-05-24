package com.funeral.funeralService.dto.account.response;

import lombok.AllArgsConstructor;
import lombok.Data;

@Data
@AllArgsConstructor
public class AccountSessionResponse {

    private Boolean isAuthenticated;

    private String phone;
}

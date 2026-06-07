package com.funeral.funeralService.dto.cemetery.request;

import lombok.AllArgsConstructor;
import lombok.Data;
import lombok.NoArgsConstructor;

@Data
@AllArgsConstructor
@NoArgsConstructor
public class CemeteryLoginRequest {

    private String login;

    private String password;
}
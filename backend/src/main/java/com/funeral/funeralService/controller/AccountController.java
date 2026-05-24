package com.funeral.funeralService.controller;

import com.funeral.funeralService.dto.account.response.AccountSessionResponse;
import com.funeral.funeralService.dto.account.request.AccountVerifyRequest;
import com.funeral.funeralService.service.AccountService;
import io.swagger.v3.oas.annotations.Operation;
import io.swagger.v3.oas.annotations.responses.ApiResponse;
import io.swagger.v3.oas.annotations.tags.Tag;
import jakarta.servlet.http.HttpSession;
import jakarta.validation.Valid;
import lombok.RequiredArgsConstructor;
import org.springframework.http.HttpStatus;
import org.springframework.web.bind.annotation.*;

@RestController
@RequestMapping("/users/account")
@RequiredArgsConstructor
@Tag(name = "Аккаунт клиента", description = "Подтверждение номера клиента и управление пользовательской сессией")
public class AccountController {

    private final AccountService service;

    @PostMapping("/verify")
    @Operation(summary = "Подтвердить код аккаунта", description = "Клиентский вход в аккаунт. Проверяет демо-код из SMS и сохраняет телефон клиента в HTTP-сессии.")
    @ApiResponse(responseCode = "200", description = "Сессия клиента создана")
    @ApiResponse(responseCode = "400", description = "Неверный код подтверждения или ошибка валидации")
    public AccountSessionResponse verify(@Valid @RequestBody AccountVerifyRequest request, HttpSession session) {
        return service.verify(request, session);
    }

    @GetMapping("/session")
    @Operation(summary = "Получить сессию клиента", description = "Возвращает состояние текущей клиентской сессии на основе cookie.")
    @ApiResponse(responseCode = "200", description = "Состояние сессии возвращено")
    public AccountSessionResponse getSession(HttpSession session) {
        return service.getSession(session);
    }

    @PostMapping("/logout")
    @ResponseStatus(HttpStatus.NO_CONTENT)
    @Operation(summary = "Выйти из аккаунта", description = "Удаляет телефон клиента из HTTP-сессии.")
    @ApiResponse(responseCode = "204", description = "Сессия клиента очищена")
    public void logout(HttpSession session) {
        service.logout(session);
    }
}

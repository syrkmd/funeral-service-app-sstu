package com.funeral.funeralService.controller;

import com.funeral.funeralService.dto.admin.request.AdminLoginRequest;
import com.funeral.funeralService.dto.admin.response.AdminSessionResponse;
import com.funeral.funeralService.service.AdminSessionService;
import io.swagger.v3.oas.annotations.Operation;
import io.swagger.v3.oas.annotations.responses.ApiResponse;
import io.swagger.v3.oas.annotations.tags.Tag;
import jakarta.servlet.http.HttpSession;
import jakarta.validation.Valid;
import lombok.RequiredArgsConstructor;
import org.springframework.http.HttpStatus;
import org.springframework.web.bind.annotation.*;

@RestController
@RequestMapping("/admin/session")
@RequiredArgsConstructor
@Tag(name = "Авторизация администратора", description = "Вход администратора и управление сессией через cookie")
public class AdminSessionController {

    private final AdminSessionService service;

    @PostMapping
    @Operation(summary = "Войти как администратор", description = "Создаёт HTTP-сессию администратора по демо-логину и паролю. Сессия хранится через cookie.")
    @ApiResponse(responseCode = "200", description = "Сессия администратора создана")
    @ApiResponse(responseCode = "401", description = "Неверный логин или пароль администратора")
    public AdminSessionResponse login(@Valid @RequestBody AdminLoginRequest request, HttpSession session) {
        return service.login(request, session);
    }

    @GetMapping
    @Operation(summary = "Получить сессию администратора", description = "Возвращает текущее состояние авторизации администратора на основе cookie.")
    @ApiResponse(responseCode = "200", description = "Состояние сессии администратора возвращено")
    public AdminSessionResponse getSession(HttpSession session) {
        return service.getSession(session);
    }

    @DeleteMapping
    @ResponseStatus(HttpStatus.NO_CONTENT)
    @Operation(summary = "Выйти из админки", description = "Очищает HTTP-сессию администратора.")
    @ApiResponse(responseCode = "204", description = "Сессия администратора очищена")
    public void logout(HttpSession session) {
        service.logout(session);
    }
}

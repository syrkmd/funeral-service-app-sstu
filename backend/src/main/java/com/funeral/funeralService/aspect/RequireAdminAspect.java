package com.funeral.funeralService.aspect;

import com.funeral.funeralService.annotation.RequireAdmin;
import com.funeral.funeralService.service.AdminSessionService;
import jakarta.servlet.http.HttpSession;
import lombok.RequiredArgsConstructor;
import org.aspectj.lang.annotation.Aspect;
import org.aspectj.lang.annotation.Before;
import org.springframework.stereotype.Component;

@Component
@Aspect
@RequiredArgsConstructor
public class RequireAdminAspect {

    private final AdminSessionService adminSessionService;
    private final HttpSession session;

    @Before("@annotation(requireAdmin)")
    public void checkAdmin(RequireAdmin requireAdmin) {
        adminSessionService.requireAuthenticated(session);
    }
}

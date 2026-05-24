package com.funeral.funeralService.service;

import com.funeral.funeralService.dto.admin.request.AdminLoginRequest;
import com.funeral.funeralService.dto.admin.response.AdminSessionResponse;
import com.funeral.funeralService.exception.InvalidAdminCredentialsException;
import jakarta.servlet.http.HttpSession;
import org.springframework.stereotype.Service;

@Service
public class AdminSessionService {

    private static final String ADMIN_SESSION_KEY = "adminAuthenticated";

    private static final String ADMIN_LOGIN = "admin";
    private static final String ADMIN_PASSWORD = "admin";

    public AdminSessionResponse login(AdminLoginRequest request, HttpSession session) {
        if (!ADMIN_LOGIN.equals(request.getLogin()) || !ADMIN_PASSWORD.equals(request.getPassword())) {
            throw new InvalidAdminCredentialsException();
        }

        session.setAttribute(ADMIN_SESSION_KEY, true);

        return new AdminSessionResponse(true);
    }

    public AdminSessionResponse getSession(HttpSession session) {
        Boolean authenticated = (Boolean) session.getAttribute(ADMIN_SESSION_KEY);

        return new AdminSessionResponse(authenticated);
    }

    public void logout(HttpSession session) {
        session.removeAttribute(ADMIN_SESSION_KEY);
    }

}

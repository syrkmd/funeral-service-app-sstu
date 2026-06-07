package com.funeral.funeralService.exception;

import com.funeral.funeralService.exception.dto.ApiError;
import org.springframework.http.HttpStatus;
import org.springframework.web.bind.MethodArgumentNotValidException;
import org.springframework.web.bind.annotation.ExceptionHandler;
import org.springframework.web.bind.annotation.ResponseStatus;
import org.springframework.web.bind.annotation.RestControllerAdvice;

@RestControllerAdvice
public class GlobalExceptionHandler {

    @ResponseStatus(HttpStatus.NOT_FOUND)
    @ExceptionHandler(OrderNotFoundException.class)
    public ApiError handleOrderNotFound(OrderNotFoundException exception) {
        return new ApiError("ORDER_NOT_FOUND", exception.getMessage());
    }

    @ResponseStatus(HttpStatus.NOT_FOUND)
    @ExceptionHandler(OrderDocumentNotFoundException.class)
    public ApiError handleOrderDocumentNotFound(OrderDocumentNotFoundException exception) {
        return new ApiError("ORDER_DOCUMENT_NOT_FOUND", exception.getMessage());
    }

    @ResponseStatus(HttpStatus.NOT_FOUND)
    @ExceptionHandler(CatalogProductNotFoundException.class)
    public ApiError handleCatalogProductNotFound(CatalogProductNotFoundException exception) {
        return new ApiError("CATALOG_PRODUCT_NOT_FOUND", exception.getMessage());
    }

    @ResponseStatus(HttpStatus.NOT_FOUND)
    @ExceptionHandler(ProductCategoryNotFoundException.class)
    public ApiError handleProductCategoryNotFound(ProductCategoryNotFoundException exception) {
        return new ApiError("PRODUCT_CATEGORY_NOT_FOUND", exception.getMessage());
    }

    @ResponseStatus(HttpStatus.NOT_FOUND)
    @ExceptionHandler(FuneralServiceNotFoundException.class)
    public ApiError handleFuneralServiceNotFound(FuneralServiceNotFoundException exception) {
        return new ApiError("FUNERAL_SERVICE_NOT_FOUND", exception.getMessage());
    }

    @ResponseStatus(HttpStatus.NOT_FOUND)
    @ExceptionHandler(ServiceCategoryNotFoundException.class)
    public ApiError handleServiceCategoryNotFound(ServiceCategoryNotFoundException exception) {
        return new ApiError("SERVICE_CATEGORY_NOT_FOUND", exception.getMessage());
    }

    @ResponseStatus(HttpStatus.BAD_REQUEST)
    @ExceptionHandler(InvalidVerificationCodeException.class)
    public ApiError handleInvalidVerificationCode(InvalidVerificationCodeException exception) {
        return new ApiError("INVALID_VERIFICATION_CODE", exception.getMessage());
    }

    @ResponseStatus(HttpStatus.UNAUTHORIZED)
    @ExceptionHandler(InvalidAdminCredentialsException.class)
    public ApiError handleInvalidAdminCredentials(InvalidAdminCredentialsException exception) {
        return new ApiError("INVALID_ADMIN_CREDENTIALS", exception.getMessage());
    }

    @ResponseStatus(HttpStatus.BAD_REQUEST)
    @ExceptionHandler(MethodArgumentNotValidException.class)
    public ApiError handleValidation(MethodArgumentNotValidException exception) {
        String message = exception.getBindingResult()
                .getFieldErrors()
                .stream()
                .findFirst()
                .map(err -> err.getField() + ": " + err.getDefaultMessage())
                .orElse("Validation error");

        return new ApiError("VALIDATION_ERROR", message);
    }

    @ResponseStatus(HttpStatus.CONFLICT)
    @ExceptionHandler(OrderAlreadyPaidException.class)
    public ApiError handleOrderAlreadyPaid(OrderAlreadyPaidException exception) {
        return new ApiError("ORDER_ALREADY_PAID", exception.getMessage());
    }

    @ResponseStatus(HttpStatus.BAD_REQUEST)
    @ExceptionHandler(BankPaymentException.class)
    public ApiError handleBankPayment(BankPaymentException exception) {
        return new ApiError("BANK_PAYMENT_FAILED", exception.getMessage());
    }

    @ResponseStatus(HttpStatus.BAD_GATEWAY)
    @ExceptionHandler(BankServiceUnavailableException.class)
    public ApiError handleBankServiceUnavailable(BankServiceUnavailableException exception) {
        return new ApiError("BANK_SERVICE_UNAVAILABLE", exception.getMessage());
    }

    @ResponseStatus(HttpStatus.UNAUTHORIZED)
    @ExceptionHandler(AdminAccessDeniedException.class)
    public ApiError handleAdminAccessDenied(AdminAccessDeniedException exception) {
        return new ApiError("ADMIN_AUTHENTICATION_REQUIRED", exception.getMessage());
    }

    @ResponseStatus(HttpStatus.BAD_REQUEST)
    @ExceptionHandler(InvalidCemeteryReservationException.class)
    public ApiError handleInvalidCemeteryReservation(InvalidCemeteryReservationException exception) {
        return new ApiError("INVALID_CEMETERY_RESERVATION", exception.getMessage());
    }

    @ResponseStatus(HttpStatus.BAD_REQUEST)
    @ExceptionHandler(InvalidDiscountException.class)
    public ApiError handleInvalidDiscount(InvalidDiscountException exception) {
        return new ApiError("INVALID_DISCOUNT", exception.getMessage());
    }

    @ResponseStatus(HttpStatus.BAD_REQUEST)
    @ExceptionHandler(InvalidOrderStatusException.class)
    public ApiError handleInvalidOrderStatus(InvalidOrderStatusException exception) {
        return new ApiError("INVALID_ORDER_STATUS", exception.getMessage());
    }

    @ResponseStatus(HttpStatus.BAD_GATEWAY)
    @ExceptionHandler(CemeteryIntegrationException.class)
    public ApiError handleCemeteryIntegration(CemeteryIntegrationException exception) {
        return new ApiError("CEMETERY_SERVICE_ERROR", exception.getMessage());
    }

}

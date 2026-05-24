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

}

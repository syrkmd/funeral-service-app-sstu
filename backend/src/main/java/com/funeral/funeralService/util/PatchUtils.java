package com.funeral.funeralService.util;

import java.util.function.Consumer;

public final class PatchUtils {

    private PatchUtils() {
    }

    public static <T> void applyIfPresent(T value, Consumer<T> setter) {
        if (value != null) {
            setter.accept(value);
        }
    }
}

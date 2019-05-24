package com.sap.piper.cm

public class ChangeManagementException extends RuntimeException {

    private static final long serialVersionUID = -139169285551665766L

    ChangeManagementException(String message) {
        super(message, null)
    }

    ChangeManagementException(String message, Throwable cause) {
        super(message, cause)
    }
}

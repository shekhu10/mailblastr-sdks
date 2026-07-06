package com.mailblastr;

/**
 * Thrown for every non-2xx API response (and for transport failures, with
 * {@code statusCode == 0} and {@code name == "network_error"}).
 *
 * <p>Mirrors the API's error body: {@code { statusCode, name, message }}.
 */
public class MailblastrException extends RuntimeException {
    private static final long serialVersionUID = 1L;

    private final int statusCode;
    private final String name;

    public MailblastrException(int statusCode, String name, String message) {
        super(message);
        this.statusCode = statusCode;
        this.name = name == null ? "application_error" : name;
    }

    /** HTTP status of the failed request (0 for transport/network failures). */
    public int getStatusCode() { return statusCode; }

    /** Machine-readable error name, e.g. {@code validation_error}. */
    public String getName() { return name; }

    @Override
    public String toString() {
        return "MailblastrException{statusCode=" + statusCode + ", name=" + name
                + ", message=" + getMessage() + "}";
    }
}

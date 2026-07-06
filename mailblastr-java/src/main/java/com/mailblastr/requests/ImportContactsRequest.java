package com.mailblastr.requests;

/**
 * Bulk-import contacts from CSV text ({@code POST /audiences/:id/contacts/import}).
 * Upserts by email. By default every non-builtin CSV column is auto-registered
 * as a custom property; {@code createProperties(false)} switches to strict
 * mode (unknown columns returned in {@code ignored_columns}).
 */
public final class ImportContactsRequest {
    private final String audienceId;
    private final String csv;
    private final String onConflict;
    private final Boolean createProperties;

    private ImportContactsRequest(Builder b) {
        if (b.audienceId == null) throw new IllegalStateException("ImportContactsRequest: audienceId is required");
        if (b.csv == null) throw new IllegalStateException("ImportContactsRequest: csv is required");
        this.audienceId = b.audienceId;
        this.csv = b.csv;
        this.onConflict = b.onConflict;
        this.createProperties = b.createProperties;
    }

    public static Builder builder() { return new Builder(); }

    public String getAudienceId() { return audienceId; }
    public String getCsv() { return csv; }
    public String getOnConflict() { return onConflict; }
    public Boolean getCreateProperties() { return createProperties; }

    public static final class Builder {
        private String audienceId;
        private String csv;
        private String onConflict;
        private Boolean createProperties;

        public Builder audienceId(String audienceId) { this.audienceId = audienceId; return this; }
        /** CSV text (header row optional). */
        public Builder csv(String csv) { this.csv = csv; return this; }
        /** {@code "upsert"} (default) or {@code "skip"}. */
        public Builder onConflict(String onConflict) { this.onConflict = onConflict; return this; }
        public Builder createProperties(boolean createProperties) { this.createProperties = createProperties; return this; }

        public ImportContactsRequest build() { return new ImportContactsRequest(this); }
    }
}

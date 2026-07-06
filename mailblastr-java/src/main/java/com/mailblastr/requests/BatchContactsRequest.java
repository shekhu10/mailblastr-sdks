package com.mailblastr.requests;

import java.util.ArrayList;
import java.util.List;

/**
 * Bulk-import contacts from a list ({@code POST /audiences/:id/contacts/batch}).
 * Upserts by email; max 10,000 per call. {@code onConflict("skip")} leaves
 * existing contacts untouched (default {@code upsert}).
 */
public final class BatchContactsRequest {
    private final String audienceId;
    private final List<ContactInput> contacts;
    private final String onConflict;

    private BatchContactsRequest(Builder b) {
        if (b.audienceId == null) throw new IllegalStateException("BatchContactsRequest: audienceId is required");
        this.audienceId = b.audienceId;
        this.contacts = b.contacts;
        this.onConflict = b.onConflict;
    }

    public static Builder builder() { return new Builder(); }

    public String getAudienceId() { return audienceId; }
    public List<ContactInput> getContacts() { return contacts; }
    public String getOnConflict() { return onConflict; }

    public static final class Builder {
        private String audienceId;
        private final List<ContactInput> contacts = new ArrayList<>();
        private String onConflict;

        public Builder audienceId(String audienceId) { this.audienceId = audienceId; return this; }
        public Builder contact(ContactInput contact) { this.contacts.add(contact); return this; }
        public Builder contacts(List<ContactInput> contacts) { this.contacts.addAll(contacts); return this; }
        /** {@code "upsert"} (default) or {@code "skip"}. */
        public Builder onConflict(String onConflict) { this.onConflict = onConflict; return this; }

        public BatchContactsRequest build() { return new BatchContactsRequest(this); }
    }
}

package com.mailblastr.requests;

/**
 * Params for {@code GET /emails}: cursor pagination plus optional server-side
 * {@code campaign_id} / {@code automation_id} / {@code source} /
 * {@code domain_id} / {@code status} / {@code search} filters.
 */
public final class ListEmailsParams {
    private final Integer limit;
    private final String after;
    private final String before;
    private final String campaignId;
    private final String automationId;
    private final String source;
    private final String domainId;
    private final String status;
    private final String search;

    private ListEmailsParams(Builder b) {
        this.limit = b.limit;
        this.after = b.after;
        this.before = b.before;
        this.campaignId = b.campaignId;
        this.automationId = b.automationId;
        this.source = b.source;
        this.domainId = b.domainId;
        this.status = b.status;
        this.search = b.search;
    }

    public static Builder builder() { return new Builder(); }

    public Integer getLimit() { return limit; }
    public String getAfter() { return after; }
    public String getBefore() { return before; }
    public String getCampaignId() { return campaignId; }
    public String getAutomationId() { return automationId; }
    public String getSource() { return source; }
    public String getDomainId() { return domainId; }
    public String getStatus() { return status; }
    public String getSearch() { return search; }

    public static final class Builder {
        private Integer limit;
        private String after;
        private String before;
        private String campaignId;
        private String automationId;
        private String source;
        private String domainId;
        private String status;
        private String search;

        /** Page size — an integer 1–100 (the API defaults to 20). */
        public Builder limit(int limit) { this.limit = limit; return this; }
        public Builder after(String after) { this.after = after; return this; }
        public Builder before(String before) { this.before = before; return this; }
        /** Only emails sent by this campaign (takes precedence over the other source filters). */
        public Builder campaignId(String campaignId) { this.campaignId = campaignId; return this; }
        /** Only emails sent by this automation. Ignored when {@code campaignId} is set. */
        public Builder automationId(String automationId) { this.automationId = automationId; return this; }
        /**
         * {@code "individual"} restricts to one-off API sends (no campaign/automation).
         * Only honoured when neither {@code campaignId} nor {@code automationId} is set.
         */
        public Builder source(String source) { this.source = source; return this; }
        /** Only emails sent from this sending domain (domain id); composes with the source filters. */
        public Builder domainId(String domainId) { this.domainId = domainId; return this; }
        /**
         * Case-insensitively match the value list rows expose as
         * {@code last_event} — e.g. {@code "delivered"}, {@code "bounced"},
         * {@code "opened"}, {@code "scheduled"}, {@code "failed"}.
         */
        public Builder status(String status) { this.status = status; return this; }
        /** Substring search across recipients, subject and sender. */
        public Builder search(String search) { this.search = search; return this; }

        public ListEmailsParams build() { return new ListEmailsParams(this); }
    }
}

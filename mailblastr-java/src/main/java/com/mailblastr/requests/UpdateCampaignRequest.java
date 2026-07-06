package com.mailblastr.requests;

import com.mailblastr.json.JsonPayload;

import java.util.Arrays;
import java.util.Collections;
import java.util.LinkedHashMap;
import java.util.Map;

/** Body of {@code PATCH /campaigns/:id}. */
public final class UpdateCampaignRequest implements JsonPayload {
    private final Map<String, Object> body;

    private UpdateCampaignRequest(Map<String, Object> body) { this.body = body; }

    public static Builder builder() { return new Builder(); }

    @Override
    public Map<String, Object> toMap() { return Collections.unmodifiableMap(body); }

    public static final class Builder {
        private final Map<String, Object> m = new LinkedHashMap<>();

        public Builder name(String name) { m.put("name", name); return this; }
        public Builder from(String from) { m.put("from", from); return this; }
        public Builder subject(String subject) { m.put("subject", subject); return this; }
        public Builder html(String html) { m.put("html", html); return this; }
        public Builder text(String text) { m.put("text", text); return this; }
        public Builder replyTo(String... replyTo) { m.put("reply_to", Arrays.asList(replyTo)); return this; }
        public Builder previewText(String previewText) { m.put("preview_text", previewText); return this; }
        /** Re-point at another of your domains' contact pools (draft campaigns only). */
        public Builder domain(String domain) { m.put("domain", domain); return this; }
        /** Re-target a segment. */
        public Builder segmentId(String segmentId) { m.put("segment_id", segmentId); return this; }
        /** Clear the segment target (sends {@code segment_id: null}). */
        public Builder clearSegment() { m.put("segment_id", null); return this; }
        /** Re-target a topic gate. */
        public Builder topicId(String topicId) { m.put("topic_id", topicId); return this; }
        /** Clear the topic gate (sends {@code topic_id: null}). */
        public Builder clearTopic() { m.put("topic_id", null); return this; }
        public Builder recurrence(String recurrence) { m.put("recurrence", recurrence); return this; }
        public Builder recurrenceEvery(int recurrenceEvery) { m.put("recurrence_every", recurrenceEvery); return this; }
        public Builder abTest(CampaignAbTest abTest) { m.put("ab_test", abTest); return this; }

        public UpdateCampaignRequest build() { return new UpdateCampaignRequest(new LinkedHashMap<>(m)); }
    }
}

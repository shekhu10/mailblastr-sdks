# frozen_string_literal: true

module Mailblastr
  # Send and inspect emails. Inbound mail lives under Mailblastr::Emails::Receiving.
  module Emails
    class << self
      # Send a single email. POST /emails
      #   Mailblastr::Emails.send({ from: "...", to: ["..."], subject: "...", html: "..." })
      # Pass { idempotency_key: "order-123" } as the second argument to retry safely.
      def send(params, options = {})
        Client.request(:post, "/emails", body: params, options: options)
      end

      # Send up to 100 emails in one request. POST /emails/batch (alias of Batch.send).
      # Batch items reject `attachments` and `scheduled_at` — send those
      # individually via Emails.send.
      def batch(payloads, options = {})
        Client.request(:post, "/emails/batch", body: payloads, options: options)
      end

      # List sent emails (trimmed list items) — cursor pagination plus optional
      # server-side `campaign_id`, `automation_id`, `source` ("individual"),
      # and `domain_id` filters. GET /emails
      def list(params = {})
        query = Client.pagination(params)
        %i[campaign_id automation_id source domain_id].each do |k|
          v = Client.opt(params, k)
          query[k] = v unless v.nil?
        end
        Client.request(:get, "/emails", query: query)
      end

      # Retrieve a sent email and its events. GET /emails/:id
      def get(email_id)
        Client.request(:get, "/emails/#{Client.path_escape(email_id)}")
      end

      # List a sent email's attachments. GET /emails/:id/attachments
      def list_attachments(email_id)
        Client.request(:get, "/emails/#{Client.path_escape(email_id)}/attachments")
      end

      # Retrieve one attachment's metadata. GET /emails/:id/attachments/:attachment_id
      def get_attachment(email_id, attachment_id)
        Client.request(:get, "/emails/#{Client.path_escape(email_id)}/attachments/#{Client.path_escape(attachment_id)}")
      end

      # Reschedule a scheduled email. PATCH /emails/:id
      #   Mailblastr::Emails.update("email_id", { scheduled_at: "2026-08-01T09:00:00Z" })
      def update(email_id, params)
        Client.request(:patch, "/emails/#{Client.path_escape(email_id)}", body: params)
      end

      # Cancel a scheduled email. POST /emails/:id/cancel
      def cancel(email_id)
        Client.request(:post, "/emails/#{Client.path_escape(email_id)}/cancel")
      end
    end

    # Inbound (received) email.
    module Receiving
      class << self
        # List received emails — cursor pagination plus an optional
        # `received_for` filter (only messages received for that address).
        # GET /emails/receiving
        def list(params = {})
          query = Client.pagination(params)
          received_for = Client.opt(params, :received_for)
          query[:received_for] = received_for unless received_for.nil?
          Client.request(:get, "/emails/receiving", query: query)
        end

        # Retrieve a received email. GET /emails/receiving/:id
        def get(email_id)
          Client.request(:get, "/emails/receiving/#{Client.path_escape(email_id)}")
        end

        # List a received email's attachments. GET /emails/receiving/:id/attachments
        def list_attachments(email_id)
          Client.request(:get, "/emails/receiving/#{Client.path_escape(email_id)}/attachments")
        end

        # Download one attachment as raw bytes (binary String).
        # GET /emails/receiving/:id/attachments/:attachment_id
        def get_attachment(email_id, attachment_id)
          Client.request(
            :get,
            "/emails/receiving/#{Client.path_escape(email_id)}/attachments/#{Client.path_escape(attachment_id)}",
            raw: true
          )
        end

        # Download the original RFC822/MIME message as raw bytes (binary String).
        # GET /emails/receiving/:id/raw
        def raw(email_id)
          Client.request(:get, "/emails/receiving/#{Client.path_escape(email_id)}/raw", raw: true)
        end

        # Forward a received email. POST /emails/receiving/:id/forward
        #   Receiving.forward(id, { from: "you@yourdomain.com", to: "team@you.com" })
        def forward(email_id, params)
          Client.request(:post, "/emails/receiving/#{Client.path_escape(email_id)}/forward", body: params)
        end

        # Reply to a received email's sender, threaded into the conversation.
        # POST /emails/receiving/:id/reply
        def reply(email_id, params)
          Client.request(:post, "/emails/receiving/#{Client.path_escape(email_id)}/reply", body: params)
        end

        # Delete a received email. DELETE /emails/receiving/:id
        def delete(email_id)
          Client.request(:delete, "/emails/receiving/#{Client.path_escape(email_id)}")
        end
      end
    end
  end

  # Batch send — Mailblastr::Batch.send([...]). POST /emails/batch
  # Batch items reject `attachments` and `scheduled_at` — send those
  # individually via Mailblastr::Emails.send.
  module Batch
    class << self
      def send(payloads, options = {})
        Client.request(:post, "/emails/batch", body: payloads, options: options)
      end
    end
  end
end

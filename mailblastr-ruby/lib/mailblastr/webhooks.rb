# frozen_string_literal: true

require "openssl"

module Mailblastr
  module Webhooks
    class << self
      # Create a webhook. The plaintext signing secret is returned ONCE, only here.
      # POST /webhooks — params: { endpoint:, events: [...], secret: }
      def create(params)
        Client.request(:post, "/webhooks", body: params)
      end

      # GET /webhooks/:id
      def get(webhook_id)
        Client.request(:get, "/webhooks/#{Client.path_escape(webhook_id)}")
      end

      # GET /webhooks
      def list(params = {})
        Client.request(:get, "/webhooks", query: Client.pagination(params))
      end

      # PATCH /webhooks/:id — params: { endpoint:, events:, status: }
      def update(webhook_id, params)
        Client.request(:patch, "/webhooks/#{Client.path_escape(webhook_id)}", body: params)
      end

      # Rotate the signing secret; the new plaintext secret is returned once and
      # the old one stops verifying immediately. POST /webhooks/:id/rotate
      def rotate(webhook_id)
        Client.request(:post, "/webhooks/#{Client.path_escape(webhook_id)}/rotate")
      end

      # Send a synchronous test delivery and return the endpoint's live result.
      # POST /webhooks/:id/test
      def test(webhook_id)
        Client.request(:post, "/webhooks/#{Client.path_escape(webhook_id)}/test")
      end

      # DELETE /webhooks/:id
      def delete(webhook_id)
        Client.request(:delete, "/webhooks/#{Client.path_escape(webhook_id)}")
      end

      # Verify a webhook delivery's Svix-style signature against your endpoint's
      # signing secret. Pure local computation (OpenSSL HMAC-SHA256) — no HTTP.
      #
      # `payload` MUST be the exact raw request body string (do not re-serialize
      # parsed JSON). `headers` is a Hash carrying svix-id / svix-timestamp /
      # svix-signature (read case-insensitively; array values use the first
      # element). The signature header may carry multiple space-separated
      # `v1,<base64>` entries — any one match makes the delivery valid.
      #
      # Returns { valid: true } or { valid: false, reason: "..." }.
      #
      #   result = Mailblastr::Webhooks.verify_signature(request.raw_post, request.headers.to_h, secret)
      #   head :unauthorized unless result[:valid]
      def verify_signature(payload, headers, secret, tolerance: 300)
        id = read_header(headers, "svix-id")
        timestamp = read_header(headers, "svix-timestamp")
        sig_header = read_header(headers, "svix-signature")
        return { valid: false, reason: "missing_headers" } if id.nil? || timestamp.nil? || sig_header.nil?
        return { valid: false, reason: "missing_secret" } if secret.nil? || secret.to_s.empty?

        # Optional timestamp freshness check (default 5 minutes; 0 disables).
        if tolerance && tolerance.positive?
          ts = begin
            Integer(timestamp, 10)
          rescue ArgumentError, TypeError
            nil
          end
          return { valid: false, reason: "invalid_timestamp" } if ts.nil?
          return { valid: false, reason: "timestamp_out_of_tolerance" } if (Time.now.to_i - ts).abs > tolerance
        end

        signed = "#{id}.#{timestamp}.#{payload}"
        digest = OpenSSL::HMAC.digest("SHA256", secret_to_key(secret), signed)
        expected = [digest].pack("m0") # strict base64, no newline

        sig_header.split(" ").each do |part|
          part = part.strip
          next if part.empty?

          sig = part.start_with?("v1,") ? part[3..] : part
          return { valid: true } if secure_compare(sig, expected)
        end
        { valid: false, reason: "no_match" }
      end

      private

      # Case-insensitively read one header value (first element if an array).
      def read_header(headers, name)
        return nil unless headers.is_a?(Hash)

        lower = name.downcase
        headers.each do |k, v|
          next unless k.to_s.downcase == lower

          return v.is_a?(Array) ? v.first : v
        end
        nil
      end

      # Derive the HMAC key from a `whsec_`-prefixed secret (base64-decode the
      # suffix); a secret without the prefix is used as raw UTF-8 bytes.
      def secret_to_key(secret)
        s = secret.to_s
        if s.start_with?("whsec_")
          decoded = s["whsec_".length..].unpack1("m") # lenient base64 decode
          return decoded if decoded && !decoded.empty?
        end
        s
      end

      # Constant-time compare of two signature strings.
      def secure_compare(a, b)
        a = a.to_s
        b = b.to_s
        return false unless a.bytesize == b.bytesize

        l = a.unpack("C*")
        res = 0
        b.each_byte.with_index { |byte, i| res |= byte ^ l[i] }
        res.zero?
      end
    end
  end
end

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
      #
      # A FAILED delivery is still HTTP 200, so this does NOT raise when your
      # endpoint rejects the test. The outcome is the "ok" key:
      #
      #   { "object" => "webhook_test", "id" => "<id>",
      #     "ok" => true,  "status" => 200 }          # endpoint accepted it
      #   { "object" => "webhook_test", "id" => "<id>",
      #     "ok" => false, "error" => "lookup_failed" } # it did not
      #
      #   result = Mailblastr::Webhooks.test(id)
      #   warn "test delivery failed: #{result['error']}" unless result["ok"]
      #
      # "status" is your endpoint's HTTP status when it responded at all;
      # "error" says why the delivery failed (e.g. "lookup_failed",
      # "webhook missing or disabled"). It is a single attempt — no retries
      # are scheduled.
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
      # parsed JSON). `headers` may be a Hash OR any pair-yielding header
      # container (e.g. Rails' `request.headers`), carrying svix-id /
      # svix-timestamp / svix-signature (read case-insensitively, rack
      # `HTTP_SVIX_ID` spellings included; array values use the first element).
      # The signature header may carry multiple space-separated `v1,<base64>`
      # entries — any one match makes the delivery valid.
      #
      # Returns { valid: true } or { valid: false, reason: "..." }.
      #
      #   # Rails: pass request.headers straight through — NOT .to_h, which
      #   # hands you the rack env spellings instead of the header names.
      #   result = Mailblastr::Webhooks.verify(request.raw_post, request.headers, secret)
      #   head :unauthorized unless result[:valid]
      def verify(payload, headers, secret, tolerance: 300)
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
      #
      # Accepts a Hash AND any pair-yielding container, because Rails'
      # `request.headers` is an ActionDispatch::Http::Headers — Enumerable, not
      # a Hash — and the old `is_a?(Hash)` guard rejected it before any lookup,
      # so the most natural call answered `missing_headers` for every genuinely
      # signed delivery. Rack/WSGI spellings (`HTTP_SVIX_ID`) are matched too,
      # since ActionDispatch's #each delegates to the raw rack env: that is what
      # `request.headers.to_h` and a bare rack `env` actually contain, and
      # downcased `http_svix_id` never equals `svix-id`.
      #
      # Enumerating is guarded: an exotic container whose #each demands a block
      # would otherwise raise LocalJumpError out of `verify`, turning a 401 into
      # a 500 inside a webhook controller. Caller-supplied input must never
      # raise here — an unreadable container is just `missing_headers`.
      def read_header(headers, name)
        return nil if headers.nil?

        lower = name.downcase
        rack = "http_#{lower.tr('-', '_')}"
        pairs =
          begin
            if headers.respond_to?(:each_pair) then headers.each_pair.to_a
            elsif headers.respond_to?(:each) then headers.each.to_a
            else return nil
            end
          rescue StandardError
            return nil
          end

        pairs.each do |k, v|
          key = k.to_s.downcase
          next unless key == lower || key == rack

          return v.is_a?(Array) ? v.first : v
        end
        nil
      end

      # Derive the HMAC key from a `whsec_`-prefixed secret (base64-decode the
      # suffix); a secret without the prefix is used as raw UTF-8 bytes.
      #
      # `tr` FIRST: the signer decodes with Node's Buffer.from(suffix, 'base64')
      # (lib/crypto.ts secretToKey), which reads "-"/"_" as the URL-safe
      # spellings of "+"/"/". `unpack1("m")` tolerates missing padding but
      # DROPS those two characters instead of translating them, yielding a
      # SHORTER, different key than the signer derived — so a caller-supplied
      # URL-safe secret (POST /webhooks accepts `secret` verbatim, unvalidated)
      # made every genuinely signed delivery come back `no_match`. Standard,
      # padded and unpadded suffixes are unaffected by the `tr`, which is why
      # the failure looked arbitrary and secret-specific.
      def secret_to_key(secret)
        s = secret.to_s
        if s.start_with?("whsec_")
          decoded = s["whsec_".length..].tr("-_", "+/").unpack1("m")
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

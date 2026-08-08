# frozen_string_literal: true

require "net/http"
require "json"
require "uri"
require "cgi"
require "time"

module Mailblastr
  # Internal HTTP layer. Every resource funnels through Client.request;
  # tests stub Client.deliver to avoid the network.
  module Client
    module_function

    VERBS = {
      get: Net::HTTP::Get,
      post: Net::HTTP::Post,
      patch: Net::HTTP::Patch,
      delete: Net::HTTP::Delete
    }.freeze

    # Only 429 and 503 are safe to retry automatically: the server guarantees
    # it did NOT apply the request, so a retry cannot duplicate a non-idempotent
    # side effect (e.g. sending an email twice). Everything else propagates.
    RETRYABLE_STATUSES = [429, 503].freeze

    # Upper bound (seconds) on any single backoff wait.
    MAX_BACKOFF_SECONDS = 30.0

    # `Idempotency-Key` is stored in a VARCHAR(255) column, so the API accepts
    # 1-255 characters measured after it trims the value — 255, not 256 — and
    # answers anything else with 400 invalid_idempotency_key. Only
    # POST /emails and POST /emails/batch read the header; every other endpoint
    # ignores it, so a retry there creates a second resource.
    #
    # Exposed for discoverability only: the SDK sends the key as given and lets
    # the server be the authority.
    IDEMPOTENCY_KEY_MAX_LENGTH = 255

    # Perform an API request and return the parsed JSON body (a Hash/Array),
    # or the raw body String when `raw: true` (binary download endpoints).
    # Raises Mailblastr::Error on any non-2xx response.
    def request(verb, path, body: nil, query: nil, options: {}, raw: false)
      key = Mailblastr.api_key
      if key.nil? || key.to_s.strip.empty?
        raise Mailblastr::Error.new(
          'Mailblastr.api_key is not set. Configure it first: Mailblastr.api_key = "mb_xxxxxxxxx"',
          error_name: "missing_api_key"
        )
      end

      uri = URI.parse("#{Mailblastr.base_url.to_s.sub(%r{/+\z}, '')}#{path}")
      if query && !query.empty?
        uri.query = [uri.query, URI.encode_www_form(query)].compact.reject(&:empty?).join("&")
      end

      req = build_request(verb, uri, body, options, key)
      handle_response(deliver_with_retries(req, uri), raw: raw)
    end

    # The single request chokepoint: both the JSON and raw/binary paths funnel
    # through here, so both get the timeout (applied inside deliver) and the
    # bounded 429/503 retry loop. Non-retryable responses return immediately;
    # network/timeout errors are NOT caught here (they propagate as before).
    def deliver_with_retries(req, uri)
      max = Mailblastr.max_retries.to_i
      attempt = 0
      loop do
        resp = deliver(req, uri)
        code = resp.code.to_i
        return resp unless RETRYABLE_STATUSES.include?(code) && attempt < max

        wait = retry_after_seconds(response_header(resp, "Retry-After"))
        wait ||= [MAX_BACKOFF_SECONDS, 0.5 * (2**attempt)].min
        backoff_sleep(wait) if wait.positive?
        attempt += 1
      end
    end

    # Wraps Kernel#sleep so tests can stub out the wait. Extracted so the
    # retry loop stays deterministic under test.
    def backoff_sleep(seconds)
      sleep(seconds)
    end

    # Read a response header without assuming the response object shape
    # (Net::HTTPResponse supports #[]; test doubles may not).
    def response_header(resp, name)
      return nil unless resp.respond_to?(:[])

      resp[name]
    rescue StandardError
      nil
    end

    # Parse a Retry-After header into a number of seconds to wait:
    # a numeric delta-seconds value, or an HTTP-date to wait until.
    # Negative is treated as 0, and the wait is capped at 30 seconds.
    # Returns nil when the header is absent or unparseable.
    def retry_after_seconds(value)
      return nil if value.nil?

      str = value.to_s.strip
      return nil if str.empty?

      seconds =
        begin
          Float(str)
        rescue ArgumentError, TypeError
          begin
            Time.httpdate(str) - Time.now
          rescue ArgumentError
            return nil
          end
        end

      seconds = 0.0 if seconds.negative?
      [seconds.to_f, MAX_BACKOFF_SECONDS].min
    end

    def build_request(verb, uri, body, options, key)
      klass = VERBS.fetch(verb) { raise ArgumentError, "unsupported HTTP verb: #{verb.inspect}" }
      req = klass.new(uri)
      req["Authorization"] = "Bearer #{key}"
      req["User-Agent"] = "mailblastr-ruby/#{Mailblastr::VERSION}"
      req["Accept"] = "application/json"
      idem = idempotency_key(opt(options, :idempotency_key))
      req["Idempotency-Key"] = idem if idem
      unless body.nil?
        req["Content-Type"] = "application/json"
        req.body = JSON.generate(body)
      end
      req
    end

    # The single seam that touches the network (stub me in tests).
    # Applies the configured open/read timeout; 0 or nil means "no timeout".
    def deliver(req, uri)
      opts = { use_ssl: uri.scheme == "https" }
      t = Mailblastr.timeout
      if !t.nil? && t.to_f > 0
        opts[:open_timeout] = t
        opts[:read_timeout] = t
      end
      Net::HTTP.start(uri.host, uri.port, **opts) do |http|
        http.request(req)
      end
    end

    def handle_response(resp, raw: false)
      code = resp.code.to_i
      body = resp.body

      if code >= 200 && code < 300
        return body if raw
        return nil if body.nil? || body.empty?

        begin
          JSON.parse(body)
        rescue JSON::ParserError
          body
        end
      else
        parsed = begin
          JSON.parse(body.to_s)
        rescue JSON::ParserError, TypeError
          nil
        end
        parsed = {} unless parsed.is_a?(Hash)
        # The whole body rides along: plan/quota errors add `limit`, reputation
        # gates add `reputation`, and a partial batch failure adds
        # `sent`/`sent_count` (see Mailblastr::Error).
        raise Mailblastr::Error.new(
          parsed["message"] || "Request failed with status #{code}",
          status_code: parsed["statusCode"] || code,
          error_name: parsed["name"] || "application_error",
          body: parsed
        )
      end
    end

    # Percent-encode one path segment so an id like "../api-keys" cannot
    # traverse the URL path (spaces become %20, "/" becomes %2F).
    def path_escape(value)
      CGI.escape(value.to_s).gsub("+", "%20")
    end

    # Normalize an `idempotency_key` option into the header value. nil/absent or
    # an empty string means "no header"; anything else is sent VERBATIM.
    #
    # The 1-255 bound (IDEMPOTENCY_KEY_MAX_LENGTH) is the server's to enforce —
    # it trims the value and answers an out-of-range key with
    # 400 invalid_idempotency_key. Checking here would only risk drifting from
    # the server, and would disagree with the other MailBlastr SDKs.
    def idempotency_key(value)
      return nil if value.nil?

      key = value.to_s
      key.empty? ? nil : key
    end

    # Read a hash param by symbol or string key.
    def opt(params, key)
      return nil unless params.is_a?(Hash)

      params.key?(key) ? params[key] : params[key.to_s]
    end

    # A copy of `params` without the given keys (symbol or string forms).
    def without(params, *keys)
      strs = keys.map(&:to_s)
      params.reject { |k, _| strs.include?(k.to_s) }
    end

    # Extract the cursor-pagination params ({ limit, after, before }).
    def pagination(params)
      q = {}
      %i[limit after before].each do |k|
        v = opt(params, k)
        q[k] = v unless v.nil?
      end
      q
    end

    # Copy the given keys out of `params` into a query hash, skipping the
    # ones the caller left out. Used to expose an endpoint's server-side
    # filters without forwarding unrelated params.
    def filters(params, *keys)
      keys.each_with_object({}) do |k, q|
        v = opt(params, k)
        q[k] = v unless v.nil?
      end
    end

    # Domain-first guard: several resources require the sending domain.
    def require_domain!(params, context)
      v = opt(params, :domain)
      if v.nil? || v.to_s.strip.empty?
        raise ArgumentError,
              "#{context} requires `domain` — the sending domain whose contact pool it targets, " \
              'e.g. { domain: "yourdomain.com", ... }'
      end
      v
    end
  end
end

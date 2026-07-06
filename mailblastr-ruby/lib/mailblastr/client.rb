# frozen_string_literal: true

require "net/http"
require "json"
require "uri"
require "cgi"

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
      handle_response(deliver(req, uri), raw: raw)
    end

    def build_request(verb, uri, body, options, key)
      klass = VERBS.fetch(verb) { raise ArgumentError, "unsupported HTTP verb: #{verb.inspect}" }
      req = klass.new(uri)
      req["Authorization"] = "Bearer #{key}"
      req["User-Agent"] = "mailblastr-ruby/#{Mailblastr::VERSION}"
      req["Accept"] = "application/json"
      idem = opt(options, :idempotency_key)
      req["Idempotency-Key"] = idem if idem
      unless body.nil?
        req["Content-Type"] = "application/json"
        req.body = JSON.generate(body)
      end
      req
    end

    # The single seam that touches the network (stub me in tests).
    def deliver(req, uri)
      Net::HTTP.start(uri.host, uri.port, use_ssl: uri.scheme == "https") do |http|
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
        raise Mailblastr::Error.new(
          parsed["message"] || "Request failed with status #{code}",
          status_code: parsed["statusCode"] || code,
          error_name: parsed["name"] || "application_error"
        )
      end
    end

    # Percent-encode one path segment so an id like "../api-keys" cannot
    # traverse the URL path (spaces become %20, "/" becomes %2F).
    def path_escape(value)
      CGI.escape(value.to_s).gsub("+", "%20")
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

# frozen_string_literal: true

require "test_helper"
require "openssl"

class WebhooksTest < Minitest::Test
  include ClientStubHelper

  def test_rest_surface
    Mailblastr::Webhooks.create({ endpoint: "https://yourapp.com/hooks", events: ["email.delivered"] })
    assert_request :post, "/webhooks"
    assert_equal ["email.delivered"], last_body["events"]

    Mailblastr::Webhooks.get("wh_1")
    assert_request :get, "/webhooks/wh_1"

    Mailblastr::Webhooks.list
    assert_request :get, "/webhooks"

    Mailblastr::Webhooks.update("wh_1", { status: "disabled" })
    assert_request :patch, "/webhooks/wh_1"

    Mailblastr::Webhooks.rotate("wh_1")
    assert_request :post, "/webhooks/wh_1/rotate"

    Mailblastr::Webhooks.test("wh_1")
    assert_request :post, "/webhooks/wh_1/test"

    Mailblastr::Webhooks.delete("wh_1")
    assert_request :delete, "/webhooks/wh_1"
  end

  # A failed test delivery is still HTTP 200 — the real outcome is "ok", so the
  # call must not raise and the caller must not read success from the status.
  def test_failed_test_delivery_is_reported_in_ok_not_by_raising
    stub_response!(200, { "object" => "webhook_test", "id" => "wh_1",
                          "ok" => false, "error" => "lookup_failed" })
    result = Mailblastr::Webhooks.test("wh_1")
    assert_request :post, "/webhooks/wh_1/test"
    assert_equal false, result["ok"]
    assert_equal "lookup_failed", result["error"]
    assert_nil result["status"]
  end

  def test_successful_test_delivery_carries_the_endpoint_status
    stub_response!(200, { "object" => "webhook_test", "id" => "wh_1", "ok" => true, "status" => 200 })
    result = Mailblastr::Webhooks.test("wh_1")
    assert_equal true, result["ok"]
    assert_equal 200, result["status"]
    assert_nil result["error"]
  end

  # --- verify (pure local HMAC; no HTTP) ---

  SECRET_BYTES = "super-secret-signing-key"
  WHSEC = "whsec_#{[SECRET_BYTES].pack('m0')}".freeze

  # Key bytes chosen so their base64 actually contains "+" and "/", which is the
  # only way the URL-safe spelling below differs from the standard one at all.
  URLSAFE_KEY_BYTES = [0xFB, 0xFF, 0xBE, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13].pack("C*").freeze
  WHSEC_STD     = "whsec_#{[URLSAFE_KEY_BYTES].pack('m0')}".freeze
  WHSEC_URLSAFE = "whsec_#{[URLSAFE_KEY_BYTES].pack('m0').tr('+/', '-_')}".freeze

  # Faithful structural stand-in for Rails' ActionDispatch::Http::Headers: an
  # Enumerable (NOT a Hash) whose #each delegates to Rack::Request::Env#each_header,
  # i.e. yields raw rack env pairs. actionpack is not a test dependency, so the
  # shape is reproduced rather than required.
  class RailsHeadersStandIn
    include Enumerable

    def initialize(env)
      @env = env
    end

    def each(&block)
      @env.each_pair(&block)
    end
  end

  def sign(payload, id: "msg_1", timestamp: Time.now.to_i.to_s, key: SECRET_BYTES)
    digest = OpenSSL::HMAC.digest("SHA256", key, "#{id}.#{timestamp}.#{payload}")
    ["v1,#{[digest].pack('m0')}", id, timestamp]
  end

  def test_valid_signature_with_whsec_secret
    payload = '{"type":"email.delivered","data":{"id":"email_1"}}'
    sig, id, ts = sign(payload)
    result = Mailblastr::Webhooks.verify(
      payload,
      { "svix-id" => id, "svix-timestamp" => ts, "svix-signature" => sig },
      WHSEC
    )
    assert_equal({ valid: true }, result)
    assert_empty @requests # pure local computation
  end

  def test_headers_are_read_case_insensitively_and_multi_sig_any_match_wins
    payload = "{}"
    sig, id, ts = sign(payload)
    result = Mailblastr::Webhooks.verify(
      payload,
      { "Svix-Id" => id, "SVIX-TIMESTAMP" => ts, "Svix-Signature" => "v1,bogus #{sig}" },
      WHSEC
    )
    assert result[:valid]
  end

  def test_raw_secret_without_whsec_prefix
    payload = "{}"
    sig, id, ts = sign(payload, key: "rawsecret")
    result = Mailblastr::Webhooks.verify(
      payload,
      { "svix-id" => id, "svix-timestamp" => ts, "svix-signature" => sig },
      "rawsecret"
    )
    assert result[:valid]
  end

  def test_tampered_payload_fails_with_no_match
    sig, id, ts = sign('{"amount":100}')
    result = Mailblastr::Webhooks.verify(
      '{"amount":999}',
      { "svix-id" => id, "svix-timestamp" => ts, "svix-signature" => sig },
      WHSEC
    )
    assert_equal({ valid: false, reason: "no_match" }, result)
  end

  def test_missing_headers_and_secret
    assert_equal "missing_headers",
                 Mailblastr::Webhooks.verify("{}", { "svix-id" => "msg_1" }, WHSEC)[:reason]
    sig, id, ts = sign("{}")
    assert_equal "missing_secret",
                 Mailblastr::Webhooks.verify(
                   "{}", { "svix-id" => id, "svix-timestamp" => ts, "svix-signature" => sig }, ""
                 )[:reason]
  end

  def test_stale_timestamp_rejected_unless_tolerance_disabled
    old_ts = (Time.now.to_i - 3600).to_s
    payload = "{}"
    sig, id, = sign(payload, timestamp: old_ts)
    headers = { "svix-id" => id, "svix-timestamp" => old_ts, "svix-signature" => sig }

    assert_equal "timestamp_out_of_tolerance",
                 Mailblastr::Webhooks.verify(payload, headers, WHSEC)[:reason]
    assert Mailblastr::Webhooks.verify(payload, headers, WHSEC, tolerance: 0)[:valid]
  end

  # The signer decodes the secret with Node's Buffer.from(suffix, 'base64')
  # (lib/crypto.ts secretToKey), which reads "-"/"_" as "+"/"/". `unpack1("m")`
  # alone DROPS those characters, deriving a shorter key than the signer, so a
  # caller-supplied URL-safe `whsec_` secret rejected every genuine delivery as
  # forged. Both spellings must derive the same key and verify the same bytes.
  def test_url_safe_whsec_secret_derives_the_same_key_as_the_standard_spelling
    refute_equal WHSEC_STD, WHSEC_URLSAFE, "test vector must actually contain + and /"

    payload = '{"type":"email.delivered","data":{"id":"email_1"}}'
    sig, id, ts = sign(payload, key: URLSAFE_KEY_BYTES)
    headers = { "svix-id" => id, "svix-timestamp" => ts, "svix-signature" => sig }

    assert Mailblastr::Webhooks.verify(payload, headers, WHSEC_STD)[:valid]
    assert_equal({ valid: true }, Mailblastr::Webhooks.verify(payload, headers, WHSEC_URLSAFE))
  end

  # Rails' `request.headers` is an ActionDispatch::Http::Headers — Enumerable,
  # not a Hash — and its #each yields rack env pairs, so the obvious Rails call
  # and the `.to_h` form the docstring used to show BOTH answered
  # `missing_headers` for deliveries MailBlastr really had signed.
  def test_rails_header_container_and_rack_env_spellings_verify
    payload = "{}"
    sig, id, ts = sign(payload)
    env = { "REQUEST_METHOD" => "POST", "HTTP_SVIX_ID" => id,
            "HTTP_SVIX_TIMESTAMP" => ts, "HTTP_SVIX_SIGNATURE" => sig }
    rails_headers = RailsHeadersStandIn.new(env)

    refute_kind_of Hash, rails_headers
    assert Mailblastr::Webhooks.verify(payload, rails_headers, WHSEC)[:valid]
    assert Mailblastr::Webhooks.verify(payload, rails_headers.to_h, WHSEC)[:valid]
    assert Mailblastr::Webhooks.verify(payload, env, WHSEC)[:valid]
  end

  # Widening the reader must not let caller-supplied input raise out of verify:
  # a LocalJumpError inside a webhook controller turns a 401 into a 500.
  def test_unreadable_header_containers_answer_missing_headers_without_raising
    block_only = Class.new do
      def each
        raise LocalJumpError, "no block given (yield)" unless block_given?
      end
    end.new

    [nil, "not-headers", block_only].each do |bad|
      assert_equal "missing_headers", Mailblastr::Webhooks.verify("{}", bad, WHSEC)[:reason]
    end
  end

  def test_invalid_timestamp
    payload = "{}"
    sig, id, = sign(payload, timestamp: "not-a-number")
    result = Mailblastr::Webhooks.verify(
      payload,
      { "svix-id" => id, "svix-timestamp" => "not-a-number", "svix-signature" => sig },
      WHSEC
    )
    assert_equal "invalid_timestamp", result[:reason]
  end
end

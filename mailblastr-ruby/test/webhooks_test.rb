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

  # --- verify_signature (pure local HMAC; no HTTP) ---

  SECRET_BYTES = "super-secret-signing-key"
  WHSEC = "whsec_#{[SECRET_BYTES].pack('m0')}".freeze

  def sign(payload, id: "msg_1", timestamp: Time.now.to_i.to_s, key: SECRET_BYTES)
    digest = OpenSSL::HMAC.digest("SHA256", key, "#{id}.#{timestamp}.#{payload}")
    ["v1,#{[digest].pack('m0')}", id, timestamp]
  end

  def test_valid_signature_with_whsec_secret
    payload = '{"type":"email.delivered","data":{"id":"email_1"}}'
    sig, id, ts = sign(payload)
    result = Mailblastr::Webhooks.verify_signature(
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
    result = Mailblastr::Webhooks.verify_signature(
      payload,
      { "Svix-Id" => id, "SVIX-TIMESTAMP" => ts, "Svix-Signature" => "v1,bogus #{sig}" },
      WHSEC
    )
    assert result[:valid]
  end

  def test_raw_secret_without_whsec_prefix
    payload = "{}"
    sig, id, ts = sign(payload, key: "rawsecret")
    result = Mailblastr::Webhooks.verify_signature(
      payload,
      { "svix-id" => id, "svix-timestamp" => ts, "svix-signature" => sig },
      "rawsecret"
    )
    assert result[:valid]
  end

  def test_tampered_payload_fails_with_no_match
    sig, id, ts = sign('{"amount":100}')
    result = Mailblastr::Webhooks.verify_signature(
      '{"amount":999}',
      { "svix-id" => id, "svix-timestamp" => ts, "svix-signature" => sig },
      WHSEC
    )
    assert_equal({ valid: false, reason: "no_match" }, result)
  end

  def test_missing_headers_and_secret
    assert_equal "missing_headers",
                 Mailblastr::Webhooks.verify_signature("{}", { "svix-id" => "msg_1" }, WHSEC)[:reason]
    sig, id, ts = sign("{}")
    assert_equal "missing_secret",
                 Mailblastr::Webhooks.verify_signature(
                   "{}", { "svix-id" => id, "svix-timestamp" => ts, "svix-signature" => sig }, ""
                 )[:reason]
  end

  def test_stale_timestamp_rejected_unless_tolerance_disabled
    old_ts = (Time.now.to_i - 3600).to_s
    payload = "{}"
    sig, id, = sign(payload, timestamp: old_ts)
    headers = { "svix-id" => id, "svix-timestamp" => old_ts, "svix-signature" => sig }

    assert_equal "timestamp_out_of_tolerance",
                 Mailblastr::Webhooks.verify_signature(payload, headers, WHSEC)[:reason]
    assert Mailblastr::Webhooks.verify_signature(payload, headers, WHSEC, tolerance: 0)[:valid]
  end

  def test_invalid_timestamp
    payload = "{}"
    sig, id, = sign(payload, timestamp: "not-a-number")
    result = Mailblastr::Webhooks.verify_signature(
      payload,
      { "svix-id" => id, "svix-timestamp" => "not-a-number", "svix-signature" => sig },
      WHSEC
    )
    assert_equal "invalid_timestamp", result[:reason]
  end
end

# frozen_string_literal: true

require "test_helper"

class EmailsTest < Minitest::Test
  include ClientStubHelper

  def test_send_posts_json_body_with_auth
    stub_response!(200, { "id" => "email_1" })
    result = Mailblastr::Emails.send({
                                       from: "Acme <hello@yourdomain.com>",
                                       to: ["user@example.com"],
                                       subject: "Hello",
                                       html: "<p>Hi</p>"
                                     })

    assert_request :post, "/emails"
    assert_equal "application/json", last_request["Content-Type"]
    assert_equal "Acme <hello@yourdomain.com>", last_body["from"]
    assert_equal ["user@example.com"], last_body["to"]
    assert_equal "Hello", last_body["subject"]
    assert_equal({ "id" => "email_1" }, result)
  end

  def test_send_sets_idempotency_key_header
    Mailblastr::Emails.send({ from: "a@b.com", to: "c@d.com", subject: "s" }, { idempotency_key: "order-123" })
    assert_equal "order-123", last_request["Idempotency-Key"]
  end

  def test_batch_send_posts_array
    Mailblastr::Batch.send([{ from: "a@b.com", to: "c@d.com", subject: "one" }])
    assert_request :post, "/emails/batch"
    assert_instance_of Array, last_body
    assert_equal "one", last_body.first["subject"]
  end

  def test_list_with_pagination
    Mailblastr::Emails.list({ limit: 20, after: "email_9" })
    assert_request :get, "/emails?limit=20&after=email_9"
    assert_nil last_request.body
  end

  def test_get_update_cancel_and_attachments
    Mailblastr::Emails.get("email_1")
    assert_request :get, "/emails/email_1"

    Mailblastr::Emails.update("email_1", { scheduled_at: "2026-08-01T09:00:00Z" })
    assert_request :patch, "/emails/email_1"
    assert_equal "2026-08-01T09:00:00Z", last_body["scheduled_at"]

    Mailblastr::Emails.cancel("email_1")
    assert_request :post, "/emails/email_1/cancel"

    Mailblastr::Emails.list_attachments("email_1")
    assert_request :get, "/emails/email_1/attachments"

    Mailblastr::Emails.get_attachment("email_1", "att_1")
    assert_request :get, "/emails/email_1/attachments/att_1"
  end

  def test_path_traversal_ids_are_percent_encoded
    Mailblastr::Emails.get("../api-keys")
    assert_request :get, "/emails/..%2Fapi-keys"
  end

  def test_receiving_surface
    Mailblastr::Emails::Receiving.list({ limit: 5 })
    assert_request :get, "/emails/receiving?limit=5"

    Mailblastr::Emails::Receiving.get("recv_1")
    assert_request :get, "/emails/receiving/recv_1"

    Mailblastr::Emails::Receiving.list_attachments("recv_1")
    assert_request :get, "/emails/receiving/recv_1/attachments"

    Mailblastr::Emails::Receiving.forward("recv_1", { from: "you@yourdomain.com", to: "team@you.com" })
    assert_request :post, "/emails/receiving/recv_1/forward"
    assert_equal "team@you.com", last_body["to"]

    Mailblastr::Emails::Receiving.reply("recv_1", { from: "you@yourdomain.com", html: "<p>Thanks</p>" })
    assert_request :post, "/emails/receiving/recv_1/reply"

    Mailblastr::Emails::Receiving.delete("recv_1")
    assert_request :delete, "/emails/receiving/recv_1"
  end

  def test_receiving_raw_downloads_return_body_string_unparsed
    stub_response!(200, "BINARY-BYTES-NOT-JSON")
    data = Mailblastr::Emails::Receiving.raw("recv_1")
    assert_request :get, "/emails/receiving/recv_1/raw"
    assert_equal "BINARY-BYTES-NOT-JSON", data

    stub_response!(200, "%PDF-1.4 ...")
    data = Mailblastr::Emails::Receiving.get_attachment("recv_1", "att_1")
    assert_request :get, "/emails/receiving/recv_1/attachments/att_1"
    assert_equal "%PDF-1.4 ...", data
  end

  def test_non_2xx_raises_mailblastr_error_with_api_shape
    stub_response!(422, { "statusCode" => 422, "name" => "validation_error", "message" => "from must be verified" })
    err = assert_raises(Mailblastr::Error) do
      Mailblastr::Emails.send({ from: "nope@unverified.com", to: "a@b.com", subject: "s" })
    end
    assert_equal 422, err.status_code
    assert_equal "validation_error", err.name
    assert_equal "from must be verified", err.message
  end

  def test_missing_api_key_raises_before_any_request
    Mailblastr.api_key = nil
    err = assert_raises(Mailblastr::Error) { Mailblastr::Emails.list }
    assert_equal "missing_api_key", err.name
    assert_empty @requests
  end

  def test_base_url_override
    Mailblastr.base_url = "http://localhost:3000/"
    Mailblastr::Emails.list
    assert_equal "localhost", last_uri.host
    assert_equal 3000, last_uri.port
    assert_equal "/emails", last_request.path
  end
end

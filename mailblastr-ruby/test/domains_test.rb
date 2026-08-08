# frozen_string_literal: true

require "test_helper"

class DomainsTest < Minitest::Test
  include ClientStubHelper

  def test_crud_and_verify
    Mailblastr::Domains.create({ name: "yourdomain.com", region: "us-east-1" })
    assert_request :post, "/domains"
    assert_equal "yourdomain.com", last_body["name"]

    Mailblastr::Domains.get("dom_1")
    assert_request :get, "/domains/dom_1"

    Mailblastr::Domains.list({ limit: 5 })
    assert_request :get, "/domains?limit=5"

    Mailblastr::Domains.update("dom_1", { click_tracking: true })
    assert_request :patch, "/domains/dom_1"
    assert_equal true, last_body["click_tracking"]

    Mailblastr::Domains.verify("dom_1")
    assert_request :post, "/domains/dom_1/verify"

    Mailblastr::Domains.delete("dom_1")
    assert_request :delete, "/domains/dom_1"
  end

  def test_claim_flow
    Mailblastr::Domains.claim({ name: "yourdomain.com" })
    assert_request :post, "/domains/claim"
    assert_equal "yourdomain.com", last_body["name"]

    Mailblastr::Domains.get_claim("dom_1")
    assert_request :get, "/domains/dom_1/claim"

    Mailblastr::Domains.verify_claim("dom_1")
    assert_request :post, "/domains/dom_1/claim/verify"
  end

  def test_dns_helpers
    Mailblastr::Domains.detect_dns("dom_1")
    assert_request :get, "/domains/dom_1/dns/detect"

    Mailblastr::Domains.apply_cloudflare_dns("dom_1", { token: "cf_token" })
    assert_request :post, "/domains/dom_1/dns/cloudflare"
    assert_equal "cf_token", last_body["token"]

    Mailblastr::Domains.apply_godaddy_dns("dom_1", { key: "k", secret: "s" })
    assert_request :post, "/domains/dom_1/dns/godaddy"
    assert_equal "s", last_body["secret"]

    Mailblastr::Domains.apply_namecheap_dns("dom_1", { apiUser: "u", apiKey: "k" })
    assert_request :post, "/domains/dom_1/dns/namecheap"
    assert_equal "u", last_body["apiUser"]
  end

  def test_mx_check_and_records_csv
    Mailblastr::Domains.mx_check("yourdomain.com")
    assert_request :get, "/domains/mx-check?name=yourdomain.com"

    # records.csv is CSV text, not JSON — it must come back as a raw String.
    stub_response!(200, "Type,Host,Full name,Value,Priority,TTL,Purpose,Status\r\nTXT,@,...")
    csv = Mailblastr::Domains.records_csv("dom_1")
    assert_request :get, "/domains/dom_1/records.csv"
    assert csv.start_with?("Type,Host,Full name")
  end

  def test_domain_id_is_escaped_in_paths
    Mailblastr::Domains.get("dom x/../../admin")
    assert_request :get, "/domains/dom%20x%2F..%2F..%2Fadmin"
  end
end

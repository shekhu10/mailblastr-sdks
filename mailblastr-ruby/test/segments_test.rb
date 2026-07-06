# frozen_string_literal: true

require "test_helper"

class SegmentsTest < Minitest::Test
  include ClientStubHelper

  def test_create_requires_domain
    err = assert_raises(ArgumentError) { Mailblastr::Segments.create({ name: "VIP" }) }
    assert_match(/domain/, err.message)
    assert_empty @requests
  end

  def test_create_posts_domain_and_filter
    Mailblastr::Segments.create({ domain: "yourdomain.com", name: "VIP", filter: { status: "subscribed" } })
    assert_request :post, "/segments"
    assert_equal "yourdomain.com", last_body["domain"]
    assert_equal "VIP", last_body["name"]
    assert_equal "subscribed", last_body.dig("filter", "status")
  end

  def test_list_requires_domain_and_paginates
    assert_raises(ArgumentError) { Mailblastr::Segments.list({}) }

    Mailblastr::Segments.list({ domain: "yourdomain.com", limit: 10 })
    assert_request :get, "/segments?domain=yourdomain.com&limit=10"
  end

  def test_get_contacts_update_delete
    Mailblastr::Segments.get("seg_1")
    assert_request :get, "/segments/seg_1"

    Mailblastr::Segments.contacts("seg_1")
    assert_request :get, "/segments/seg_1/contacts"

    Mailblastr::Segments.update("seg_1", { name: "VIP customers" })
    assert_request :patch, "/segments/seg_1"
    assert_equal "VIP customers", last_body["name"]

    Mailblastr::Segments.delete("seg_1")
    assert_request :delete, "/segments/seg_1"
  end
end

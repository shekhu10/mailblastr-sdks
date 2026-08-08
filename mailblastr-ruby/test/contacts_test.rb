# frozen_string_literal: true

require "test_helper"

class ContactsTest < Minitest::Test
  include ClientStubHelper

  def test_create_flat_requires_domain
    assert_raises(ArgumentError) { Mailblastr::Contacts.create({ email: "a@b.com" }) }
    assert_empty @requests
  end

  def test_create_flat_sends_domain_in_body
    Mailblastr::Contacts.create({ domain: "yourdomain.com", email: "a@b.com", first_name: "Ada" })
    assert_request :post, "/contacts"
    assert_equal "yourdomain.com", last_body["domain"]
    assert_equal "a@b.com", last_body["email"]
    assert_equal "Ada", last_body["first_name"]
  end

  def test_create_nested_audience_strips_routing_keys_from_body
    Mailblastr::Contacts.create({ audience_id: "aud_1", email: "a@b.com" })
    assert_request :post, "/audiences/aud_1/contacts"
    refute last_body.key?("audience_id")
    assert_equal "a@b.com", last_body["email"]
  end

  def test_get_by_id_and_by_email_with_domain
    Mailblastr::Contacts.get({ id: "cont_1" })
    assert_request :get, "/contacts/cont_1"

    Mailblastr::Contacts.get({ id: "a@b.com", domain: "yourdomain.com" })
    assert_request :get, "/contacts/a%40b.com?domain=yourdomain.com"

    Mailblastr::Contacts.get({ id: "cont_1", audience_id: "aud_1" })
    assert_request :get, "/audiences/aud_1/contacts/cont_1"
  end

  def test_list_flat_requires_domain_and_supports_filters
    assert_raises(ArgumentError) { Mailblastr::Contacts.list }

    Mailblastr::Contacts.list({ domain: "yourdomain.com", limit: 50, segment_id: "seg_1" })
    assert_request :get, "/contacts?domain=yourdomain.com&limit=50&segment_id=seg_1"
  end

  def test_list_nested_audience
    Mailblastr::Contacts.list({ audience_id: "aud_1", after: "cont_9" })
    assert_request :get, "/audiences/aud_1/contacts?after=cont_9"
  end

  def test_update_flat_keeps_domain_in_body_but_not_id
    Mailblastr::Contacts.update({ id: "a@b.com", domain: "yourdomain.com", unsubscribed: true })
    assert_request :patch, "/contacts/a%40b.com"
    assert_equal true, last_body["unsubscribed"]
    assert_equal "yourdomain.com", last_body["domain"]
    refute last_body.key?("id")
  end

  def test_delete_flat_with_domain_query_and_nested
    Mailblastr::Contacts.delete({ id: "a@b.com", domain: "yourdomain.com" })
    assert_request :delete, "/contacts/a%40b.com?domain=yourdomain.com"

    Mailblastr::Contacts.delete({ id: "cont_1", audience_id: "aud_1" })
    assert_request :delete, "/audiences/aud_1/contacts/cont_1"
  end

  def test_batch_import
    Mailblastr::Contacts.batch({ audience_id: "aud_1", contacts: [{ email: "a@b.com" }], on_conflict: "skip" })
    assert_request :post, "/audiences/aud_1/contacts/batch?on_conflict=skip"
    assert_equal [{ "email" => "a@b.com" }], last_body["contacts"]
  end

  def test_csv_import_with_strict_properties
    Mailblastr::Contacts.import({ audience_id: "aud_1", csv: "email\na@b.com", create_properties: false })
    assert_request :post, "/audiences/aud_1/contacts/import?create_properties=false"
    assert_equal "email\na@b.com", last_body["csv"]
  end

  def test_csv_import_forwards_segment_id_and_file_name
    Mailblastr::Contacts.import({
                                  audience_id: "aud_1",
                                  csv: "email\na@b.com",
                                  file_name: "june.csv",
                                  on_conflict: "skip",
                                  segment_id: "seg_1"
                                })
    assert_request :post, "/audiences/aud_1/contacts/import?on_conflict=skip&segment_id=seg_1"
    assert_equal "june.csv", last_body["file_name"]
  end

  def test_csv_import_by_storage_key_does_not_send_a_null_csv
    Mailblastr::Contacts.import({ audience_id: "aud_1", storage_key: "imports/abc.csv" })
    assert_request :post, "/audiences/aud_1/contacts/import"
    assert_equal({ "storage_key" => "imports/abc.csv" }, last_body)
  end

  def test_import_upload_mints_a_presigned_slot
    Mailblastr::Contacts.import_upload({ audience_id: "aud_1", filename: "list.csv", size: 2048 })
    assert_request :post, "/audiences/aud_1/contacts/import/upload"
    assert_equal "list.csv", last_body["filename"]
    assert_equal 2048, last_body["size"]
    refute last_body.key?("audience_id")
  end

  def test_segment_membership_and_topics
    Mailblastr::Contacts.add_to_segment("cont_1", "seg_1")
    assert_request :post, "/contacts/cont_1/segments/seg_1"

    Mailblastr::Contacts.remove_from_segment("cont_1", "seg_1")
    assert_request :delete, "/contacts/cont_1/segments/seg_1"

    Mailblastr::Contacts.list_segments("cont_1")
    assert_request :get, "/contacts/cont_1/segments"

    Mailblastr::Contacts.list_segments("cont_1", { limit: 10 })
    assert_request :get, "/contacts/cont_1/segments?limit=10"

    Mailblastr::Contacts.get_topics("cont_1")
    assert_request :get, "/contacts/cont_1/topics"

    Mailblastr::Contacts.get_topics("cont_1", { limit: 10 })
    assert_request :get, "/contacts/cont_1/topics?limit=10"

    Mailblastr::Contacts.update_topics("cont_1", { topics: [{ id: "top_1", subscription: "opt_in" }] })
    assert_request :patch, "/contacts/cont_1/topics"
    assert_equal "opt_in", last_body["topics"].first["subscription"]
  end

  def test_string_keys_work_like_symbol_keys
    Mailblastr::Contacts.create({ "domain" => "yourdomain.com", "email" => "a@b.com" })
    assert_request :post, "/contacts"
    assert_equal "yourdomain.com", last_body["domain"]
  end
end

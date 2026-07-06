# frozen_string_literal: true

require "mailblastr/version"
require "mailblastr/error"
require "mailblastr/client"
require "mailblastr/emails"
require "mailblastr/domains"
require "mailblastr/audiences"
require "mailblastr/contacts"
require "mailblastr/contact_properties"
require "mailblastr/segments"
require "mailblastr/topics"
require "mailblastr/campaigns"
require "mailblastr/templates"
require "mailblastr/automations"
require "mailblastr/webhooks"
require "mailblastr/events"
require "mailblastr/api_keys"
require "mailblastr/logs"
require "mailblastr/polls"

# The MailBlastr API client.
#
#   Mailblastr.api_key = "mb_xxxxxxxxx"
#
#   sent = Mailblastr::Emails.send({
#     from: "Acme <hello@yourdomain.com>",
#     to: ["user@example.com"],
#     subject: "Hello from MailBlastr",
#     html: "<p>Your first email</p>"
#   })
#   sent["id"] # => "..."
module Mailblastr
  DEFAULT_BASE_URL = "https://api.mailblastr.com"

  class << self
    # Your API key, e.g. "mb_xxxxxxxxx". Required before any call.
    attr_accessor :api_key

    # Override the API host (defaults to https://api.mailblastr.com).
    attr_writer :base_url

    def base_url
      @base_url || DEFAULT_BASE_URL
    end

    #   Mailblastr.configure do |config|
    #     config.api_key = ENV["MAILBLASTR_API_KEY"]
    #   end
    def configure
      yield self
    end
  end
end

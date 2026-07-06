# frozen_string_literal: true

module Mailblastr
  # Raised for every non-2xx API response. Mirrors the API error shape
  # { statusCode, name, message }:
  #
  #   begin
  #     Mailblastr::Emails.send(params)
  #   rescue Mailblastr::Error => e
  #     e.status_code # => 422
  #     e.name        # => "validation_error"
  #     e.message     # => "The `from` address must use a verified domain."
  #   end
  class Error < StandardError
    attr_reader :status_code, :error_name

    def initialize(message = nil, status_code: nil, error_name: nil)
      super(message)
      @status_code = status_code
      @error_name = error_name
    end

    # The API error `name` (e.g. "validation_error", "not_found").
    def name
      @error_name
    end
  end
end

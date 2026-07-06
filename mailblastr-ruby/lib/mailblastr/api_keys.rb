# frozen_string_literal: true

module Mailblastr
  module ApiKeys
    class << self
      # Create an API key — the full token is returned only once, here.
      # POST /api-keys — params: { name:, permission: "full_access"|"sending_access", domain_id: }
      def create(params)
        Client.request(:post, "/api-keys", body: params)
      end

      # GET /api-keys
      def list
        Client.request(:get, "/api-keys")
      end

      # DELETE /api-keys/:id
      def delete(api_key_id)
        Client.request(:delete, "/api-keys/#{Client.path_escape(api_key_id)}")
      end
    end
  end
end

# frozen_string_literal: true

module Mailblastr
  # Campaigns are DOMAIN-FIRST: `domain` (required on create) picks the
  # contact pool the campaign targets; `from` may be any verified domain.
  module Campaigns
    class << self
      # POST /campaigns
      #   Mailblastr::Campaigns.create({ domain: "yourdomain.com", from: "you@yourdomain.com",
      #                                  subject: "Hello", html: "<p>Hi</p>", segment_id: "seg_1" })
      def create(params)
        Client.require_domain!(params, "Campaigns.create")
        Client.request(:post, "/campaigns", body: params)
      end

      # GET /campaigns/:id
      def get(campaign_id)
        Client.request(:get, "/campaigns/#{Client.path_escape(campaign_id)}")
      end

      # GET /campaigns
      def list(params = {})
        Client.request(:get, "/campaigns", query: Client.pagination(params))
      end

      # PATCH /campaigns/:id
      def update(campaign_id, params)
        Client.request(:patch, "/campaigns/#{Client.path_escape(campaign_id)}", body: params)
      end

      # Send now, or schedule with { scheduled_at: "..." }. POST /campaigns/:id/send
      def send(campaign_id, params = {})
        Client.request(:post, "/campaigns/#{Client.path_escape(campaign_id)}/send", body: params)
      end

      # Cancel a scheduled campaign (returns it to draft). POST /campaigns/:id/cancel
      def cancel(campaign_id)
        Client.request(:post, "/campaigns/#{Client.path_escape(campaign_id)}/cancel")
      end

      # Per-campaign analytics (counts, engagement rates, top links). GET /campaigns/:id/stats
      def stats(campaign_id)
        Client.request(:get, "/campaigns/#{Client.path_escape(campaign_id)}/stats")
      end

      # A/B winner evaluation for an A/B campaign. GET /campaigns/:id/ab
      def ab(campaign_id)
        Client.request(:get, "/campaigns/#{Client.path_escape(campaign_id)}/ab")
      end

      # DELETE /campaigns/:id
      def delete(campaign_id)
        Client.request(:delete, "/campaigns/#{Client.path_escape(campaign_id)}")
      end
    end
  end
end

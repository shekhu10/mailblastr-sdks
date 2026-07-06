# frozen_string_literal: true

module Mailblastr
  # Custom events that automations trigger on. Events are DOMAIN-FIRST:
  # `domain` is required on send, and only that domain's automations fire.
  module Events
    class << self
      # Send a custom event. POST /events/send
      #   Mailblastr::Events.send({ event: "signup.completed", domain: "yourdomain.com",
      #                             email: "user@example.com", payload: { plan: "pro" } })
      # Identify the contact by `contact_id` OR `email`. Pass
      # { idempotency_key: "..." } as the second argument to retry safely.
      def send(params, options = {})
        Client.require_domain!(params, "Events.send")
        Client.request(:post, "/events/send", body: params, options: options)
      end

      # Create a custom-event definition (name + optional payload schema). POST /events
      #   Mailblastr::Events.create({ name: "signup.completed", schema: { plan: "string" } })
      def create(params, options = {})
        Client.request(:post, "/events", body: params, options: options)
      end

      # List custom-event definitions. GET /events
      def list(params = {})
        Client.request(:get, "/events", query: Client.pagination(params))
      end

      # Delete a custom-event definition. DELETE /events/:id
      def delete(event_id)
        Client.request(:delete, "/events/#{Client.path_escape(event_id)}")
      end
    end
  end
end

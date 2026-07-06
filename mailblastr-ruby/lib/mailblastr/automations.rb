# frozen_string_literal: true

module Mailblastr
  # Automations are DOMAIN-FIRST: `domain` is required on create, and only
  # Events.send calls naming the same domain trigger them.
  module Automations
    class << self
      # POST /automations
      #   Mailblastr::Automations.create({ name: "Welcome series", domain: "yourdomain.com",
      #                                    trigger: "contact.created" })
      def create(params)
        Client.require_domain!(params, "Automations.create")
        Client.request(:post, "/automations", body: params)
      end

      # GET /automations/:id
      def get(automation_id)
        Client.request(:get, "/automations/#{Client.path_escape(automation_id)}")
      end

      # GET /automations
      def list(params = {})
        Client.request(:get, "/automations", query: Client.pagination(params))
      end

      # PATCH /automations/:id — params: { name:, status: "enabled"|"disabled", ... }
      def update(automation_id, params)
        Client.request(:patch, "/automations/#{Client.path_escape(automation_id)}", body: params)
      end

      # Append a step. POST /automations/:id/steps — params: { type:, config:, key: }
      def add_step(automation_id, params)
        Client.request(:post, "/automations/#{Client.path_escape(automation_id)}/steps", body: params)
      end

      # Delete a step. DELETE /automations/:id/steps/:step_id
      def delete_step(automation_id, step_id)
        Client.request(:delete, "/automations/#{Client.path_escape(automation_id)}/steps/#{Client.path_escape(step_id)}")
      end

      # List an automation's runs. GET /automations/:id/runs
      def runs(automation_id, params = {})
        Client.request(:get, "/automations/#{Client.path_escape(automation_id)}/runs", query: Client.pagination(params))
      end

      # Retrieve a single run with its step trace. GET /automations/:id/runs/:run_id
      def get_run(automation_id, run_id)
        Client.request(:get, "/automations/#{Client.path_escape(automation_id)}/runs/#{Client.path_escape(run_id)}")
      end

      # Stop an automation — no new runs; in-progress runs finish. POST /automations/:id/stop
      def stop(automation_id)
        Client.request(:post, "/automations/#{Client.path_escape(automation_id)}/stop")
      end

      # DELETE /automations/:id
      def delete(automation_id)
        Client.request(:delete, "/automations/#{Client.path_escape(automation_id)}")
      end
    end
  end
end

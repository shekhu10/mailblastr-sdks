# frozen_string_literal: true

module Mailblastr
  module Domains
    class << self
      # Register a sending domain. POST /domains
      def create(params)
        Client.request(:post, "/domains", body: params)
      end

      # GET /domains/:id
      def get(domain_id)
        Client.request(:get, "/domains/#{Client.path_escape(domain_id)}")
      end

      # GET /domains
      def list(params = {})
        Client.request(:get, "/domains", query: Client.pagination(params))
      end

      # PATCH /domains/:id (returns the slim ack { object: "domain", id }).
      def update(domain_id, params)
        Client.request(:patch, "/domains/#{Client.path_escape(domain_id)}", body: params)
      end

      # Trigger DNS verification. POST /domains/:id/verify
      def verify(domain_id)
        Client.request(:post, "/domains/#{Client.path_escape(domain_id)}/verify")
      end

      # Claim a domain already verified in another account. POST /domains/claim
      def claim(params)
        Client.request(:post, "/domains/claim", body: params)
      end

      # Retrieve a domain's claim record. GET /domains/:id/claim
      def get_claim(domain_id)
        Client.request(:get, "/domains/#{Client.path_escape(domain_id)}/claim")
      end

      # Verify a domain claim's TXT record. POST /domains/:id/claim/verify
      def verify_claim(domain_id)
        Client.request(:post, "/domains/#{Client.path_escape(domain_id)}/claim/verify")
      end

      # Detect the DNS provider and available one-click apply methods.
      # GET /domains/:id/dns/detect
      def detect_dns(domain_id)
        Client.request(:get, "/domains/#{Client.path_escape(domain_id)}/dns/detect")
      end

      # Apply DNS records via the Cloudflare API, then auto-verify.
      # POST /domains/:id/dns/cloudflare — params: { token: "..." }
      def apply_cloudflare_dns(domain_id, params)
        Client.request(:post, "/domains/#{Client.path_escape(domain_id)}/dns/cloudflare", body: params)
      end

      # Apply DNS records via the GoDaddy API, then auto-verify.
      # POST /domains/:id/dns/godaddy — params: { key: "...", secret: "..." }
      def apply_godaddy_dns(domain_id, params)
        Client.request(:post, "/domains/#{Client.path_escape(domain_id)}/dns/godaddy", body: params)
      end

      # Apply DNS records via the Namecheap API (existing records preserved),
      # then auto-verify. POST /domains/:id/dns/namecheap
      # params: { api_user: "...", api_key: "...", user_name: "..." } — note the
      # backend expects camelCase keys (apiUser/apiKey/userName), so pass those.
      def apply_namecheap_dns(domain_id, params)
        Client.request(:post, "/domains/#{Client.path_escape(domain_id)}/dns/namecheap", body: params)
      end

      # DELETE /domains/:id
      def delete(domain_id)
        Client.request(:delete, "/domains/#{Client.path_escape(domain_id)}")
      end
    end
  end
end

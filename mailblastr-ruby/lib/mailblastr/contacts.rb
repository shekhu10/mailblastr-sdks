# frozen_string_literal: true

module Mailblastr
  # Contacts are DOMAIN-FIRST: each sending domain has its own contact pool,
  # so the flat /contacts API takes `domain` (required on create/list). The
  # nested audience variants (`audience_id:`) derive the pool from the path.
  module Contacts
    class << self
      # Create a contact. POST /contacts (flat, `domain` required) or
      # POST /audiences/:id/contacts when `audience_id` is given.
      #   Mailblastr::Contacts.create({ domain: "yourdomain.com", email: "a@b.com" })
      #   Mailblastr::Contacts.create({ audience_id: "aud_1", email: "a@b.com" })
      def create(params)
        audience_id = Client.opt(params, :audience_id)
        if audience_id
          body = Client.without(params, :audience_id, :domain)
          Client.request(:post, "/audiences/#{Client.path_escape(audience_id)}/contacts", body: body)
        else
          Client.require_domain!(params, "Contacts.create (flat /contacts API)")
          Client.request(:post, "/contacts", body: Client.without(params, :audience_id))
        end
      end

      # Retrieve a contact by id (exact) or by email. An email can exist in
      # several domains' pools, so pass `domain` to pick the pool.
      #   Contacts.get({ id: "cont_1" })
      #   Contacts.get({ id: "a@b.com", domain: "yourdomain.com" })
      #   Contacts.get({ id: "cont_1", audience_id: "aud_1" })
      def get(params)
        id = Client.path_escape(Client.opt(params, :id))
        audience_id = Client.opt(params, :audience_id)
        return Client.request(:get, "/audiences/#{Client.path_escape(audience_id)}/contacts/#{id}") if audience_id

        domain = Client.opt(params, :domain)
        query = domain ? { domain: domain } : nil
        Client.request(:get, "/contacts/#{id}", query: query)
      end

      # List contacts. Flat /contacts requires `domain` (names the pool);
      # pass `audience_id` instead to use the nested API. `segment_id`
      # filters either variant; limit/after/before paginate.
      def list(params = {})
        audience_id = Client.opt(params, :audience_id)
        query = Client.pagination(params)
        segment_id = Client.opt(params, :segment_id)
        query[:segment_id] = segment_id if segment_id

        if audience_id
          Client.request(:get, "/audiences/#{Client.path_escape(audience_id)}/contacts", query: query)
        else
          query = { domain: Client.require_domain!(params, "Contacts.list (flat /contacts API)") }.merge(query)
          Client.request(:get, "/contacts", query: query)
        end
      end

      # Update a contact (id or email). PATCH /contacts/:id, or the nested
      # route when `audience_id` is given. On the flat API pass `domain` when
      # `id` is an email (disambiguates across pools).
      #   Contacts.update({ id: "cont_1", unsubscribed: true })
      def update(params)
        id = Client.path_escape(Client.opt(params, :id))
        audience_id = Client.opt(params, :audience_id)
        if audience_id
          body = Client.without(params, :audience_id, :domain, :id)
          Client.request(:patch, "/audiences/#{Client.path_escape(audience_id)}/contacts/#{id}", body: body)
        else
          Client.request(:patch, "/contacts/#{id}", body: Client.without(params, :audience_id, :id))
        end
      end

      # Delete a contact. DELETE /contacts/:id (pass `domain` when `id` is an
      # email), or the nested route when `audience_id` is given.
      def delete(params)
        id = Client.path_escape(Client.opt(params, :id))
        audience_id = Client.opt(params, :audience_id)
        return Client.request(:delete, "/audiences/#{Client.path_escape(audience_id)}/contacts/#{id}") if audience_id

        domain = Client.opt(params, :domain)
        query = domain ? { domain: domain } : nil
        Client.request(:delete, "/contacts/#{id}", query: query)
      end

      # Bulk-import contacts from an array (upsert by email; max 10,000).
      # POST /audiences/:id/contacts/batch
      #   Contacts.batch({ audience_id: "aud_1", contacts: [{ email: "a@b.com" }], on_conflict: "skip" })
      def batch(params)
        audience_id = Client.opt(params, :audience_id)
        query = {}
        on_conflict = Client.opt(params, :on_conflict)
        query[:on_conflict] = on_conflict if on_conflict
        Client.request(
          :post,
          "/audiences/#{Client.path_escape(audience_id)}/contacts/batch",
          body: { contacts: Client.opt(params, :contacts) },
          query: query
        )
      end

      # Bulk-import contacts from CSV text (header row optional; upsert by
      # email). Non-builtin columns auto-register as custom properties unless
      # `create_properties: false`. POST /audiences/:id/contacts/import
      #   Contacts.import({ audience_id: "aud_1", csv: "email\na@b.com" })
      def import(params)
        audience_id = Client.opt(params, :audience_id)
        query = {}
        on_conflict = Client.opt(params, :on_conflict)
        query[:on_conflict] = on_conflict if on_conflict
        query[:create_properties] = "false" if Client.opt(params, :create_properties) == false
        Client.request(
          :post,
          "/audiences/#{Client.path_escape(audience_id)}/contacts/import",
          body: { csv: Client.opt(params, :csv) },
          query: query
        )
      end

      # Add a contact to a segment. POST /contacts/:id/segments/:segment_id
      def add_to_segment(contact_id, segment_id)
        Client.request(:post, "/contacts/#{Client.path_escape(contact_id)}/segments/#{Client.path_escape(segment_id)}")
      end

      # Remove a contact from a segment. DELETE /contacts/:id/segments/:segment_id
      def remove_from_segment(contact_id, segment_id)
        Client.request(:delete, "/contacts/#{Client.path_escape(contact_id)}/segments/#{Client.path_escape(segment_id)}")
      end

      # List the segments a contact belongs to. GET /contacts/:id/segments
      def list_segments(contact_id)
        Client.request(:get, "/contacts/#{Client.path_escape(contact_id)}/segments")
      end

      # Get a contact's topic subscriptions. GET /contacts/:id/topics
      def get_topics(contact_id)
        Client.request(:get, "/contacts/#{Client.path_escape(contact_id)}/topics")
      end

      # Update a contact's topic subscriptions. PATCH /contacts/:id/topics
      #   Contacts.update_topics("cont_1", { topics: [{ id: "top_1", subscription: "opt_in" }] })
      def update_topics(contact_id, params)
        Client.request(:patch, "/contacts/#{Client.path_escape(contact_id)}/topics", body: params)
      end
    end
  end
end

package server

import (
	"github.com/rs/cors"
	"github.com/rykov/paperboy/config"

	"net/http"
)

func GraphQLHandler(cfg *config.AConfig) http.Handler {
	// CORS allows central preview
	c := cors.New(cors.Options{
		AllowedOrigins: []string{
			"http://www.paperboy.email",
			"http://paperboy.email",
			"http://localhost:*",
		},
	})

	// GraphQL handler for exposed API
	handler := MustSchemaHandler(schemaText, &Resolver{cfg: cfg})
	return c.Handler(handler)
}

const schemaText = `
  schema {
    query: Query
    mutation: Mutation
  }

  # The Query type, represents all of the entry points
  type Query {
    campaigns: [Campaign]!
    lists: [RecipientList]!
    renderOne(content: String!, recipient: String!): RenderedEmail
    paperboyInfo: PaperboyInfo!
  }

  # All mutations
  type Mutation {
    sendBeta(content: String!, recipients: [RecipientInput!]!): Int!
    sendCampaign(campaign: String!, list: String!): Boolean!
  }

  # A single rendered email information
  type RenderedEmail {
    rawMessage: String!
    text: String!
    html: String
    # html: HTML
  }

  # Build/version information
  type PaperboyInfo {
    version: String!
    buildDate: String!
  }

  # Campaign metadata
  type Campaign {
    param: String!
    subject: String!
  }

  # Recipient list metadata
  type RecipientList {
    param: String!
    name: String!
  }

  # Recipient metadata
  input RecipientInput {
    email: String!
    params: JSON
  }

  # HTML (same as string)
  scalar HTML

  # JSON (freeform object)
  scalar JSON
`

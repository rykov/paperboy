package server

import (
	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	"github.com/rs/cors"

	"net/http"
)

func GraphQLHandler() http.Handler {
	// CORS allows central preview
	c := cors.New(cors.Options{
		AllowedOrigins: []string{
			"http://www.paperboy.email",
			"http://paperboy.email",
			"http://localhost:*",
		},
	})

	schema := graphql.MustParseSchema(schemaText, &Resolver{})
	return c.Handler(&relay.Handler{Schema: schema})
}

const schemaText = `
  schema {
    query: Query
  }

  # The Query type, represents all of the entry points
  type Query {
    campaigns: [Campaign]!
    lists: [RecipientList]!
    renderOne(content: String!, recipient: String!): RenderedEmail
    paperboyInfo: PaperboyInfo!
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

  # HTML (same as string)
  scalar HTML
`

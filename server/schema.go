package server

import (
	"github.com/neelance/graphql-go"
	"github.com/neelance/graphql-go/relay"
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
    renderOne(content: String!, recipient: String!): RenderedEmail
  }

  # A single rendered email information
  type RenderedEmail {
    rawMessage: String!
    text: String!
    html: String
    # html: HTML
  }

  # HTML (same as string)
  scalar HTML
`

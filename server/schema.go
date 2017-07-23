package server

import (
	"github.com/neelance/graphql-go"
	"github.com/neelance/graphql-go/relay"

	"net/http"
)

func AddGraphQLRoutes() {
	schema := graphql.MustParseSchema(schemaText, &Resolver{})
	http.Handle("/graphql", &relay.Handler{Schema: schema})
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

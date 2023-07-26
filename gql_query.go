package graphvent
import (
  "github.com/graphql-go/graphql"
)

var GQLQuerySelf = &graphql.Field{
  Type: GQLTypeBaseThread.Type,
  Resolve: func(p graphql.ResolveParams) (interface{}, error) {
    _, ctx, err := PrepResolve(p)
    if err != nil {
      return nil, err
    }

    return ctx.Server, nil
  },
}

var GQLQueryUser = &graphql.Field{
  Type: GQLTypeBaseNode.Type,
  Resolve: func(p graphql.ResolveParams) (interface{}, error) {
    _, ctx, err := PrepResolve(p)
    if err != nil {
      return nil, err
    }

    return ctx.User, nil
  },
}

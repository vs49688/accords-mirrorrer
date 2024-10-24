# gqlclient

[![godocs.io](https://godocs.io/git.sr.ht/~emersion/gqlclient?status.svg)](https://godocs.io/git.sr.ht/~emersion/gqlclient)
[![builds.sr.ht status](https://builds.sr.ht/~emersion/gqlclient/commits.svg)](https://builds.sr.ht/~emersion/gqlclient/commits?)

A GraphQL client and code generator for Go.

## Usage

gqlclient can be used as a thin GraphQL client, and can be augmented with code
generation. See the [GoDoc examples] for direct usage.

### GraphQL schema code generation

The code generator can parse a GraphQL schema and generate Go types. For
instance, the following schema:

```graphqls
type Train {
	name: String!
	maxSpeed: Int!
	weight: Int!
	linesServed: [String!]!
}
```

and the following `gqlclientgen` invocation:

```sh
gqlclientgen -s schema.graphqls -o gql.go -n rail
```

will generate the following Go type:

```go
type Train struct {
	Name string
	MaxSpeed int32
	Weight int32
	LinesServed []string
}
```

which can then be used in a GraphQL query:

```go
op := gqlclient.NewOperation(`query {
	train(name: "Shinkansen E5") {
		name
		maxSpeed
		linesServed
	}
}`)

var data struct {
	Train rail.Train
}
if err := c.Execute(ctx, op, &data); err != nil {
	log.Fatal(err)
}
log.Print(data.Train)
```

### GraphQL query code generation

The code generator can also parse a GraphQL query document and generate Go
functions. For instance, the following query document:

```graphql
query fetchTrain($name: String!) {
	train(name: $name) {
		maxSpeed
		linesServed
	}
}
```

and the following `gqlclientgen` invocation:

```sh
gqlclientgen -s schema.graphqls -q queries.graphql -o gql.go -n rail
```

will generate the following function:

```go
func FetchTrain(client *gqlclient.Client, ctx context.Context, name string) (Train, error)
```

which can then be used to execute the query:

```go
train, err := rail.FetchTrain(c, ctx, "Shinkansen E5")
if err != nil {
	log.Fatal(err)
}
log.Print(train)
```

### GraphQL schema introspection

gqlclient also supports fetching GraphQL schemas through GraphQL introspection.
For instance, the following `gqlintrospect` invocation will fetch the GraphQL
schema of the `https://example.com/query` GraphQL endpoint:

```sh
gqlintrospect https://example.com/query > schema.graphqls
```

## Contributing

Send patches on the [mailing list]. Discuss in [#emersion on Libera Chat][IRC channel].

## License

MIT

[GoDoc examples]: https://godocs.io/git.sr.ht/~emersion/gqlclient#example-Client-Execute
[mailing list]: https://lists.sr.ht/~emersion/gqlclient-dev
[IRC channel]: ircs://irc.libera.chat/#emersion

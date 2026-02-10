// Package playcamp provides a Go client for the PlayCamp SDK API.
//
// Two client types are available:
//
//   - [Client] for read-only operations using a CLIENT API key
//   - [Server] for read/write operations using a SERVER API key
//
// # Quick Start
//
//	client, err := playcamp.NewClient("keyId:secret",
//	    playcamp.WithEnvironment(playcamp.EnvironmentSandbox),
//	)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	campaigns, err := client.Campaigns.List(ctx, nil)
package playcamp

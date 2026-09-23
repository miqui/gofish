# gofish

A minimal, stdlib-only Go client for HPE iLO 7 Redfish APIs.

This initial implementation focuses on two resource groups from the iLO 7
v1.23 Redfish resource map:

- **Core compute**: `ComputerSystem`, `Processor`, `Memory`,
  `EthernetInterface`, `Bios`
- **Chassis & hardware**: `Chassis`, `Power`, `Thermal`, `Manager`

It is not affiliated with or a drop-in replacement for
[`github.com/hewlettpackard/gofish`](https://github.com/hewlettpackard/gofish);
it's a small, purpose-built client with no dependencies beyond the Go
standard library.

## Install

```sh
go get github.com/miqui/gofish
```

## Usage

```go
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/miqui/gofish"
)

func main() {
	ctx := context.Background()

	client, err := gofish.NewClient(gofish.ClientConfig{
		Endpoint:           "https://ilo.example.com",
		Username:           "admin",
		Password:           "password",
		InsecureSkipVerify: true, // iLO often uses self-signed certs
	})
	if err != nil {
		log.Fatal(err)
	}

	// Enumerate systems and print basic health.
	systemIDs, err := gofish.ListSystems(ctx, client)
	if err != nil {
		log.Fatal(err)
	}
	for _, id := range systemIDs {
		sys, err := gofish.GetSystem(ctx, client, id)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("%s: power=%s health=%s\n", sys.Name, sys.PowerState, sys.Status.Health)
	}

	// Inspect chassis power and thermal state.
	chassisIDs, err := gofish.ListChassis(ctx, client)
	if err != nil {
		log.Fatal(err)
	}
	for _, id := range chassisIDs {
		ch, err := gofish.GetChassis(ctx, client, id)
		if err != nil {
			log.Fatal(err)
		}
		thermal, err := ch.GetThermal(ctx, client)
		if err != nil {
			log.Fatal(err)
		}
		for _, fan := range thermal.Fans {
			fmt.Printf("%s: %d%%\n", fan.Name, fan.Reading)
		}
	}

	// Reboot a system gracefully.
	sys, err := gofish.GetSystem(ctx, client, "1")
	if err != nil {
		log.Fatal(err)
	}
	if err := sys.Reset(ctx, client, gofish.ResetTypeGracefulRestart); err != nil {
		log.Fatal(err)
	}
}
```

## Resource coverage

| Group          | Resource            | Read | Actions       |
|----------------|----------------------|------|---------------|
| Core compute   | ComputerSystem       | Yes  | Reset         |
| Core compute   | Processor            | Yes  | -             |
| Core compute   | Memory               | Yes  | -             |
| Core compute   | EthernetInterface    | Yes  | -             |
| Core compute   | Bios                 | Yes  | -             |
| Chassis & HW   | Chassis              | Yes  | -             |
| Chassis & HW   | Power                | Yes  | -             |
| Chassis & HW   | Thermal              | Yes  | -             |
| Chassis & HW   | Manager              | Yes  | -             |

Additional resource groups and write actions (BIOS settings, manager
resets, virtual media, etc.) can be layered on top of the same `Client`
primitives (`Get`/`Post`/`Patch`/`Delete`) in follow-up changes.

## Development

```sh
make build   # go build ./...
make test    # go test ./... -race -cover
make vet     # go vet ./...
make fmt     # gofmt -w .
make ci      # fmt-check + vet + test (what CI runs, minus golangci-lint)
```

CI (`.github/workflows/ci.yml`) additionally runs `golangci-lint`.

package main

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/elliottpolk/bodega"
	"github.com/elliottpolk/bodega/config"

	log "github.com/sirupsen/logrus"
	cli "github.com/urfave/cli/v2"
	altsrc "github.com/urfave/cli/v2/altsrc"
)

var (
	version  string
	compiled string = fmt.Sprint(time.Now().Unix())
	githash  string

	RpcPortFlag = altsrc.NewStringFlag(&cli.StringFlag{
		Name:    "server.rpc-port",
		Value:   "7000",
		Usage:   "RPC port to listen on",
		EnvVars: []string{"BODEGA_RPC_PORT"},
	})

	HttpPortFlag = altsrc.NewStringFlag(&cli.StringFlag{
		Name:    "server.http-port",
		Value:   "8080",
		Usage:   "HTTP port to listen on",
		EnvVars: []string{"BODEGA_HTTP_PORT"},
	})

	HttpsPortFlag = altsrc.NewStringFlag(&cli.StringFlag{
		Name:    "server.tls-port",
		Value:   "8443",
		Usage:   "HTTPS port to listen on",
		EnvVars: []string{"BODEGA_HTTPS_PORT"},
	})

	TlsCertFlag = altsrc.NewStringFlag(&cli.StringFlag{
		Name:    "server.tls-cert",
		Usage:   "TLS certificate file for HTTPS",
		EnvVars: []string{"BODEGA_TLS_CERT"},
	})

	TlsKeyFlag = altsrc.NewStringFlag(&cli.StringFlag{
		Name:    "server.tls-key",
		Usage:   "TLS key file for HTTPS",
		EnvVars: []string{"BODEGA_TLS_KEY"},
	})

	DatastoreAddrFlag = altsrc.NewStringFlag(&cli.StringFlag{
		Name:    "datastore.addr",
		Aliases: []string{"ds.addr", "dsa"},
		Usage:   "Database address",
	})

	DatastorePortFlag = altsrc.NewStringFlag(&cli.StringFlag{
		Name:    "datastore.port",
		Aliases: []string{"ds.port", "dsp"},
		Value:   "27017",
		Usage:   "Database port",
	})

	DatastoreNameFlag = altsrc.NewStringFlag(&cli.StringFlag{
		Name:    "datastore.name",
		Aliases: []string{"ds.name", "dsn"},
		Value:   "bodega",
		Usage:   "Database name",
	})

	DatastoreUserFlag = altsrc.NewStringFlag(&cli.StringFlag{
		Name:    "datastore.user",
		Aliases: []string{"ds.user", "dsu"},
		Usage:   "Database user",
	})

	DatastorePasswordFlag = altsrc.NewStringFlag(&cli.StringFlag{
		Name:    "datastore.password",
		Aliases: []string{"ds.password", "dspwd"},
		Usage:   "Database password",
	})

	CfgFlag = altsrc.NewStringFlag(&cli.StringFlag{
		Name:    "config",
		Aliases: []string{"c", "cfg", "confg"},
		Usage:   "optional path to config file",
	})

	flags = []cli.Flag{
		CfgFlag,
		RpcPortFlag,
		HttpPortFlag,
		HttpsPortFlag,
		TlsCertFlag,
		TlsKeyFlag,
		DatastoreAddrFlag,
		DatastorePortFlag,
		DatastoreNameFlag,
		DatastoreUserFlag,
		DatastorePasswordFlag,
	}
)

func main() {
	ct, err := strconv.ParseInt(compiled, 0, 0)
	if err != nil {
		panic(err)
	}

	app := cli.App{
		Name:      "bodega",
		Copyright: fmt.Sprintf("Copyright © 2018-%s Elliott Polk", time.Now().Format("2006")),
		Version:   fmt.Sprintf("%s | compiled %s | commit %s", version, time.Unix(ct, -1).Format(time.RFC3339), githash),
		Compiled:  time.Unix(ct, -1),
		Flags:     flags,
		Before: func(ctx *cli.Context) error {
			if len(ctx.String(CfgFlag.Name)) > 0 {
				return altsrc.InitInputSourceWithContext(flags, altsrc.NewYamlSourceFromFlagFunc(CfgFlag.Name))(ctx)
			}
			return nil
		},
		Action: func(ctx *cli.Context) error {
			// read in the configuration
			comp := &config.Composition{
				Server: &config.Server{
					RpcPort:   ctx.String(RpcPortFlag.Name),
					HttpPort:  ctx.String(HttpPortFlag.Name),
					HttpsPort: ctx.String(HttpsPortFlag.Name),
					TlsCert:   ctx.String(TlsCertFlag.Name),
					TlsKey:    ctx.String(TlsKeyFlag.Name),
				},
				Db: &config.Db{
					Addr:     ctx.String(DatastoreAddrFlag.Name),
					Port:     ctx.String(DatastorePortFlag.Name),
					Name:     ctx.String(DatastoreNameFlag.Name),
					User:     ctx.String(DatastoreUserFlag.Name),
					Password: ctx.String(DatastorePasswordFlag.Name),
				},
			}

			// run in a non-blocking goroutine since it is blocking
			go func() {
				if err := bodega.ServeREST(context.Background(), comp); err != nil {
					log.Fatal(err)
				}
			}()

			// use this one to block and prevent exiting
			if err := bodega.ServeGRPC(context.Background(), comp); err != nil {
				return cli.Exit(err, 1)
			}

			return nil
		},
	}

	app.Run(os.Args)
}

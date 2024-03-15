package main

import (
	"context"
	"fmt"
	"os"

	"github.com/alecthomas/kong"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"github.com/hideckies/hermit/pkg/client/console"
	"github.com/hideckies/hermit/pkg/client/rpc"
	"github.com/hideckies/hermit/pkg/client/state"
	"github.com/hideckies/hermit/pkg/common/config"
	"github.com/hideckies/hermit/pkg/common/meta"
	metafs "github.com/hideckies/hermit/pkg/common/meta/fs"
	"github.com/hideckies/hermit/pkg/common/stdout"
	"github.com/hideckies/hermit/pkg/protobuf/rpcpb"
)

type Context struct {
	Debug bool
}

type StartCmd struct {
	Config string `short:"c" required:"" help:"The config file generated by the C2 server." type:"path"`
}

func run(configPath string) error {
	stdout.PrintBannerClient()

	// Set a log file
	logFile, err := metafs.OpenLogFile(true)
	if err != nil {
		return err
	}
	defer logFile.Close()

	// Read a client config
	clientConfig, err := config.ReadClientConfigJson(configPath, true)
	if err != nil {
		return err
	}

	addr := fmt.Sprintf("%s:%d", clientConfig.Server.Host, clientConfig.Server.Port)
	stdout.LogInfo(fmt.Sprintf("Connecting to the C2 server at %v", addr))

	// Initialize a client state
	clientState := state.NewClientState(clientConfig)

	var opts []grpc.DialOption

	// Temporarily create a ca cert file.
	tmpDir, err := metafs.GetTempDir(true)
	if err != nil {
		return err
	}
	caCertPath := fmt.Sprintf("%s/ca-cert.pem", tmpDir)
	err = os.WriteFile(caCertPath, []byte(clientConfig.CaCertificate), 0644)
	if err != nil {
		return err
	}

	// Read a client certificate for connecting to RPC server
	creds, err := credentials.NewClientTLSFromFile(caCertPath, "")
	if err != nil {
		return err
	}
	opts = append(opts, grpc.WithTransportCredentials(creds))

	// Connect to the C2 server
	conn, err := grpc.Dial(addr, opts...)
	if err != nil {
		return err
	}
	defer conn.Close()

	c := rpcpb.NewHermitRPCClient(conn)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Add the operator to the database.
	res, err := rpc.RequestOperatorRegister(c, ctx, *clientConfig)
	if err != nil {
		return err
	}
	stdout.LogSuccess(res)

	if err := console.Readline(c, ctx, clientState); err != nil {
		return err
	}

	// Send a request to the C2 server for deleting the current operator
	_, err = rpc.RequestOperatorDeleteByUuid(c, ctx, clientConfig.Uuid)
	if err != nil {
		return err
	}
	stdout.LogSuccess("The operator deleted from the C2 server successfully.")
	stdout.LogSuccess("See you again.")

	return nil
}

func (s *StartCmd) Run(ctx *Context) error {
	if err := run(s.Config); err != nil {
		stdout.LogFailed(fmt.Sprintf("%v", err))
		os.Exit(1)
	}
	return nil
}

type VersionCmd struct{}

func (v *VersionCmd) Run(ctx *Context) error {
	meta.PrintVersion()
	return nil
}

var cli struct {
	Debug bool `help:"Enable debug mode."`

	Start   StartCmd   `cmd:"" default:"withargs"`
	Version VersionCmd `cmd:"" help:"Print the version of Hermit"`
}

func main() {
	if err := metafs.MakeAppDirs(true); err != nil {
		stdout.LogFailed(fmt.Sprintf("failed to create app directories: %v", err))
		os.Exit(1)
	}

	ctx := kong.Parse(
		&cli,
		kong.Name("hermit-client"),
		kong.Description("Hermit C2 Client"),
		kong.UsageOnError(),
		kong.ConfigureHelp(kong.HelpOptions{
			Compact: true,
			Summary: false,
		}))

	err := ctx.Run(&Context{Debug: cli.Debug})
	ctx.FatalIfErrorf(err)
}

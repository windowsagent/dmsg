// Package commands cmd/dmsgpty-cli/commands/root.go
package commands

import (
	"context"
	"encoding/json"
	"log"
	"os"

	"github.com/skycoin/skywire-utilities/pkg/buildinfo"
	"github.com/skycoin/skywire-utilities/pkg/cmdutil"
	"github.com/spf13/cobra"

	dmsg "github.com/skycoin/dmsg/pkg/dmsg"
	"github.com/skycoin/dmsg/pkg/dmsgpty"
)

var cli = dmsgpty.DefaultCLI()

// path for config file ( required for whitelists )
var (
	defaultConfPath = "config.json"
	confPath        string
)

// conf to update whitelists
var conf dmsgpty.Config

var remoteAddr dmsg.Addr
var cmdName = dmsgpty.DefaultCmd
var cmdArgs []string

func init() {
	RootCmd.PersistentFlags().StringVar(&cli.Net, "clinet", cli.Net,
		"network to use for dialing to dmsgpty-host")

	RootCmd.PersistentFlags().StringVar(&cli.Addr, "cliaddr", cli.Addr,
		"address to use for dialing to dmsgpty-host")

	RootCmd.PersistentFlags().StringVarP(&confPath, "confpath", confPath,
		defaultConfPath, "config path")

	cobra.OnInitialize(initConfig)
	RootCmd.Flags().Var(&remoteAddr, "addr",
		"remote dmsg address of format 'pk:port'\n If unspecified, the pty will start locally\n")

	RootCmd.Flags().StringVarP(&cmdName, "cmd", "c", cmdName,
		"name of command to run\n")

	RootCmd.Flags().StringSliceVarP(&cmdArgs, "args", "a", cmdArgs,
		"command arguments")

}

// initConfig sources whitelist from config file
// by default : it will look for config
//
// case 1 : config file is new (does not contain a "wl" key)
// - create a "wl" key within the config file
//
// case 2 : config file is old (already contains "wl" key)
// - load config file into memory to manipulate whitelists
// - writes changes back to config file
func initConfig() {

	println(confPath)

	if _, err := os.Stat(confPath); err != nil {
		cli.Log.Fatalf("Config file %s not found.", confPath)
	}

	// read file using ioutil
	file, err := os.ReadFile(confPath) //nolint:gosec
	if err != nil {
		cli.Log.Fatalln("Unable to read ", confPath, err)
	}

	// store config.json into conf to manipulate whitelists
	err = json.Unmarshal(file, &conf)
	if err != nil {
		cli.Log.Errorln(err)
		// ignoring this error
		b, err := json.MarshalIndent(conf, "", "  ")
		if err != nil {
			cli.Log.Fatalln("Unable to marshal conf")
		}

		// write to config.json
		err = os.WriteFile(confPath, b, 0600)
		if err != nil {
			cli.Log.Fatalln("Unable to write", confPath, err)
		}
	}
	conf.CLIAddr = dmsgpty.ParseWindowsEnv(conf.CLIAddr)
	if conf.CLIAddr != "" {
		cli.Addr = conf.CLIAddr
	}
	if conf.CLINet != "" {
		cli.Net = conf.CLINet
	}
}

// RootCmd contains commands for dmsgpty-cli; which interacts with the dmsgpty-host instance (i.e. skywire-visor)
var RootCmd = &cobra.Command{
	Use:   "dmsgpty-cli",
	Short: "Run commands over dmsg",
	PreRun: func(*cobra.Command, []string) {
		if remoteAddr.Port == 0 {
			remoteAddr.Port = dmsgpty.DefaultPort
		}
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		if _, err := buildinfo.Get().WriteTo(log.Writer()); err != nil {
			log.Printf("Failed to output build info: %v", err)
		}

		ctx, cancel := cmdutil.SignalContext(context.Background(), nil)
		defer cancel()

		if remoteAddr.PK.Null() {
			// Local pty.
			return cli.StartLocalPty(ctx, cmdName, cmdArgs...)
		}
		// Remote pty.
		return cli.StartRemotePty(ctx, remoteAddr.PK, remoteAddr.Port, cmdName, cmdArgs...)
	},
}

// Execute executes the root command.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

// Command lifx allows performing basic operations on LIFX devices over the LAN
package main

import (
	"bytes"
	"fmt"
	"net"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"

	"github.com/pdf/golifx"
	"github.com/pdf/golifx/common"
	"github.com/pdf/golifx/protocol"
)

var (
	client *golifx.Client

	flagTimeout  time.Duration
	flagLogLevel string
	flagIP       net.IP
	flagPort     int

	logger = logrus.New()
	app    = &cobra.Command{
		Use: `lifx`,
		PersistentPreRun: func(c *cobra.Command, args []string) {
			setLogger()
		},
	}

	cmdGenerateBashComp = &cobra.Command{
		Use:   `bashcomp <filename>`,
		Short: "generate bash completion at <file>",
		Run:   generateBashComp,
	}

	cmdGenerateDocs = &cobra.Command{
		Use:   `docs <path>`,
		Short: "generate markdown documentation at <path>",
		Run:   generateDocs,
	}

	cmdVersion = &cobra.Command{
		Use:   `version`,
		Short: "output the lifx version",
		Run:   version,
	}

	cmdWatch = &cobra.Command{
		Use:    `watch`,
		Short:  "watch for events (hint: set log level to 'debug'), end with Ctrl+C",
		PreRun: setupClient,
		Run:    watch,
	}
)

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	golifx.SetLogger(logger)

	app.PersistentFlags().DurationVarP(&flagTimeout, `timeout`, `t`, common.DefaultTimeout, `timeout for all operations`)
	app.PersistentFlags().StringVarP(&flagLogLevel, `log-level`, `L`, `info`, `log level, one of: [debug,info,warn,error]`)
	app.PersistentFlags().IPVarP(&flagIP, `ip-address`, `I`, net.IPv4(0, 0, 0, 0), `UDP listen address`)
	app.PersistentFlags().IntVarP(&flagPort, `port`, `p`, 56700, `UDP listen port`)

	app.AddCommand(cmdLight)
	app.AddCommand(cmdGroup)
	app.AddCommand(cmdGenerateBashComp)
	app.AddCommand(cmdGenerateDocs)
	app.AddCommand(cmdVersion)
	app.AddCommand(cmdWatch)
}

func main() {
	if err := app.Execute(); err != nil {
		logger.WithField(`error`, err).Fatalln(`Failed starting app`)
	}
}

func setupClient(c *cobra.Command, args []string) {
	var err error

	client, err = golifx.NewClient(&protocol.V2{Reliable: true, Port: flagPort})
	if err != nil {
		logger.WithField(`error`, err).Fatalln(`Failed initializing client`)
	}
}

func closeClient(c *cobra.Command, args []string) {
	err := client.Close()
	if err != nil {
		logger.WithField(`error`, err).Fatalln(`Failed closing client`)
	}
}

func generateBashComp(c *cobra.Command, args []string) {
	if len(args) != 1 {
		if err := c.Usage(); err != nil {
			logger.WithField(`error`, err).Fatalln(`Failed to print usage`)
		}
		fmt.Println()
		logger.Fatalln(`Missing filename`)
	}

	buf := new(bytes.Buffer)
	f, err := os.Create(args[0])
	if err != nil {
		logger.WithFields(logrus.Fields{
			`filename`: args[0],
			`error`:    err,
		}).Fatalln(`Could not open file`)
	}
	err = app.GenBashCompletion(buf)
	if err != nil {
		logger.WithFields(logrus.Fields{
			`filename`: args[0],
			`error`:    err,
		}).Fatalln(`Could not generate completion`)
	}
	if _, err := buf.WriteTo(f); err != nil {
		logger.WithField(`error`, err).Fatalln(`Failed writing to file`)
	}
}

func generateDocs(c *cobra.Command, args []string) {
	if len(args) != 1 {
		if err := c.Usage(); err != nil {
			logger.WithField(`error`, err).Fatalln(`Failed to print usage`)
		}
		fmt.Println()
		logger.Fatalln(`Missing output path`)
	}

	path := args[0]
	if path[len(path)-1] != os.PathSeparator {
		path += string(os.PathSeparator)
	}
	if err := doc.GenMarkdownTree(app, path); err != nil {
		logger.WithFields(logrus.Fields{
			`filename`: args[0],
			`error`:    err,
		}).Fatalln(`Could no generate markdown`)
	}
}

func version(c *cobra.Command, args []string) {
	fmt.Printf("lifx version %s\n", golifx.VERSION)
}

func usage(c *cobra.Command, args []string) {
	if err := c.Usage(); err != nil {
		logger.WithField(`error`, err).Fatalln(`Failed to print usage`)
	}
}

func watch(c *cobra.Command, args []string) {
	sig := make(chan os.Signal, 1)

	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	<-sig
}

func setLogger() {
	switch flagLogLevel {
	case `debug`:
		logger.Level = logrus.DebugLevel
	case `info`:
		logger.Level = logrus.InfoLevel
	case `warn`:
		logger.Level = logrus.WarnLevel
	case `error`:
		logger.Level = logrus.ErrorLevel
	default:
		logger.Level = logrus.InfoLevel
	}
}

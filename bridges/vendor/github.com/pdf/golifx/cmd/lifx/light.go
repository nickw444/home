package main

import (
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/pdf/golifx/common"
	"github.com/spf13/cobra"
)

var (
	flagLightIDs        []int
	flagLightLabels     []string
	flagLightHue        uint16
	flagLightSaturation uint16
	flagLightBrightness uint16
	flagLightKelvin     uint16
	flagLightDuration   time.Duration

	cmdLightList = &cobra.Command{
		Use:     `list`,
		Short:   `list available lights`,
		PreRun:  setupClient,
		Run:     lightList,
		PostRun: closeClient,
	}

	cmdLightColor = &cobra.Command{
		Use:     `color`,
		Short:   `set light color`,
		PreRun:  setupClient,
		Run:     lightColor,
		PostRun: closeClient,
	}

	cmdLightPower = &cobra.Command{
		Use:       `power`,
		Short:     `[on|off]`,
		Long:      `lifx light power [on|off]`,
		ValidArgs: []string{`on`, `off`},
		PreRun:    setupClient,
		Run:       lightPower,
		PostRun:   closeClient,
	}

	cmdLight = &cobra.Command{
		Use:   `light`,
		Short: `interact with lights`,
		Long: `Interact with lights.
Acts on all lights by default, however you may restrict the lights that a command applies to by specifying IDs or labels via the flags listed below.`,
		Run: usage,
	}
)

func init() {
	cmdLightColor.Flags().Uint16VarP(&flagLightHue, `hue`, `H`, 0, `hue component of the HSBK color (0-65535)`)
	cmdLightColor.Flags().Uint16VarP(&flagLightSaturation, `saturation`, `S`, 0, `saturation component of the HSBK color (0-65535)`)
	cmdLightColor.Flags().Uint16VarP(&flagLightBrightness, `brightness`, `B`, 0, `brightness component of the HSBK color (0-65535)`)
	cmdLightColor.Flags().Uint16VarP(&flagLightKelvin, `kelvin`, `K`, 0, `kelvin component of the HSBK color, the color temperature of whites (2500-9000)`)
	if err := cmdLightColor.MarkFlagRequired(`hue`); err != nil {
		logger.WithField(`error`, err).Panicln(`Failed initializing application`)
	}
	if err := cmdLightColor.MarkFlagRequired(`saturation`); err != nil {
		logger.WithField(`error`, err).Panicln(`Failed initializing application`)
	}
	if err := cmdLightColor.MarkFlagRequired(`brightness`); err != nil {
		logger.WithField(`error`, err).Panicln(`Failed initializing application`)
	}
	if err := cmdLightColor.MarkFlagRequired(`kelvin`); err != nil {
		logger.WithField(`error`, err).Panicln(`Failed initializing application`)
	}
	cmdLight.AddCommand(cmdLightList)
	cmdLight.AddCommand(cmdLightColor)
	cmdLight.AddCommand(cmdLightPower)

	cmdLight.PersistentFlags().IntSliceVarP(&flagLightIDs, `id`, `i`, make([]int, 0), `ID of the light(s) to manage, comma-seprated.  Defaults to all lights`)
	cmdLight.PersistentFlags().StringSliceVarP(&flagLightLabels, `label`, `l`, make([]string, 0), `label of the light(s) to manage, comma-separated.  Defaults to all lights.`)
	cmdLight.PersistentFlags().DurationVarP(&flagLightDuration, `duration`, `d`, 0*time.Second, `duration of the power/color transition`)
}

func lightList(c *cobra.Command, args []string) {
	var (
		err     error
		lights  []common.Light
		timeout <-chan time.Time
	)

	if flagTimeout == 0 {
		logger.Fatalln(`Can not list with a timeout of zero`)
	}

	if len(flagLightIDs) > 0 || len(flagLightLabels) > 0 {
		lights = getLights()
	} else {
		timeout = time.After(flagTimeout)
		<-timeout

		lights, err = client.GetLights()
		if err == common.ErrNotFound {
			logger.Fatalln(`No lights found`)
		} else if err != nil {
			logger.WithField(`error`, err).Fatalln(`Could not find lights`)
		}
	}

	table := new(tabwriter.Writer)
	table.Init(os.Stdout, 0, 4, 4, ' ', 0)
	fmt.Fprintf(table, fmt.Sprintf("%s\t%s\t%s\t%s\t%s\t%s\n", `ID`, `Label`, `Power`, `Color`, `Firmware`, `Product`))

	for _, l := range lights {
		label, err := l.GetLabel()
		if err != nil {
			logger.WithField(`light-id`, l.ID()).Warnln(`Couldn't get color for light`)
			continue
		}
		power, err := l.GetPower()
		if err != nil {
			logger.WithField(`light-id`, l.ID()).Warnln(`Couldn't get color for light`)
			continue
		}
		color, err := l.GetColor()
		if err != nil {
			logger.WithField(`light-id`, l.ID()).Warnln(`Couldn't get color for light`)
			continue
		}
		firmwareVersion, err := l.GetFirmwareVersion()
		if err != nil {
			logger.WithField(`light-id`, l.ID()).Warnln(`Couldn't get firmware version for light`)
			continue
		}
		productName, err := l.GetProductName()
		if err != nil {
			logger.WithField(`light-id`, l.ID()).Warnln(`Couldn't get product name for light`)
			continue
		}
		fmt.Fprintf(table, "%v\t%s\t%v\t%+v\t%s\t%s\n", l.ID(), label, power, color, firmwareVersion, productName)
	}
	fmt.Fprintln(table)
	if err := table.Flush(); err != nil {
		logger.WithField(`error`, err).Fatalln(`Failed outputting results`)
	}
}

func getLights() []common.Light {
	var lights []common.Light

	logger.WithField(`ids`, flagLightIDs).Debug(`Requested IDs`)
	logger.WithField(`labels`, flagLightLabels).Debug(`Requested labels`)

	if len(flagLightIDs) > 0 {
		for _, id := range flagLightIDs {
			light, err := client.GetLightByID(uint64(id))
			if err != nil {
				logger.WithFields(logrus.Fields{
					`error`: err,
					`ID`:    id,
				}).Fatalln(`Could not find light with requested ID`)
			}
			lights = append(lights, light)
		}
	}
	if len(flagLightLabels) > 0 {
		for _, label := range flagLightLabels {
			light, err := client.GetLightByLabel(label)
			if err != nil {
				logger.WithFields(logrus.Fields{
					`error`: err,
					`label`: label,
				}).Fatalln(`Could not find light with requested label`)
			}
			lights = append(lights, light)
		}
	}

	return lights
}

func lightPower(c *cobra.Command, args []string) {
	if len(args) < 1 {
		if err := c.Usage(); err != nil {
			logger.WithField(`error`, err).Fatalln(`Failed to print usage`)
		}
		fmt.Println()
		logger.Fatalln(`Missing state (on|off)`)
	}

	var state bool

	switch args[0] {
	case `on`:
		state = true
	case `off`:
		state = false
	default:
		if err := c.Usage(); err != nil {
			logger.WithField(`error`, err).Fatalln(`Failed to print usage`)
		}
		fmt.Println()
		logger.WithField(`state`, args[0]).Fatalln(`Invalid power state requested, should be one of [on|off]`)
	}

	lights := getLights()

	if len(lights) > 0 {
		for _, light := range lights {
			if err := light.SetPowerDuration(state, flagLightDuration); err != nil {
				logger.WithFields(logrus.Fields{
					`light-id`: light.ID(),
					`error`:    err,
				}).Fatalln(`Failed setting power for light`)
			}
		}
	} else {
		if err := client.SetPowerDuration(state, flagLightDuration); err != nil {
			logger.WithField(`error`, err).Fatalln(`Failed setting power for lights`)
		}
	}
}

func lightColor(c *cobra.Command, args []string) {
	if flagLightHue == 0 && flagLightSaturation == 0 && flagLightBrightness == 0 && flagLightKelvin == 0 {
		if err := c.Usage(); err != nil {
			logger.WithField(`error`, err).Fatalln(`Failed to print usage`)
		}
		fmt.Println()
		logger.Fatalln(`Missing color definition`)
	}

	lights := getLights()

	color := common.Color{
		Hue:        flagLightHue,
		Saturation: flagLightSaturation,
		Brightness: flagLightBrightness,
		Kelvin:     flagLightKelvin,
	}

	if len(lights) > 0 {
		for _, light := range lights {
			if err := light.SetColor(color, flagLightDuration); err != nil {
				logger.WithFields(logrus.Fields{
					`light-id`: light.ID(),
					`error`:    err,
				}).Fatalln(`Failed setting color for light`)
			}
		}
	} else {
		if err := client.SetColor(color, flagLightDuration); err != nil {
			logger.WithField(`error`, err).Fatalln(`Failed setting color for lights`)
		}
	}
}

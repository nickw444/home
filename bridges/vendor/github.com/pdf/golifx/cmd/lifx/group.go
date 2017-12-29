package main

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/pdf/golifx/common"
	"github.com/spf13/cobra"
)

var (
	flagGroupIDs        []string
	flagGroupLabels     []string
	flagGroupHue        uint16
	flagGroupSaturation uint16
	flagGroupBrightness uint16
	flagGroupKelvin     uint16
	flagGroupDuration   time.Duration

	cmdGroupList = &cobra.Command{
		Use:     `list`,
		Short:   `list available groups`,
		PreRun:  setupClient,
		Run:     groupList,
		PostRun: closeClient,
	}

	cmdGroupColor = &cobra.Command{
		Use:     `color`,
		Short:   `set group color`,
		PreRun:  setupClient,
		Run:     groupColor,
		PostRun: closeClient,
	}

	cmdGroupPower = &cobra.Command{
		Use:       `power`,
		Short:     `[on|off]`,
		Long:      `lifx group power [on|off]`,
		ValidArgs: []string{`on`, `off`},
		PreRun:    setupClient,
		Run:       groupPower,
		PostRun:   closeClient,
	}

	cmdGroup = &cobra.Command{
		Use:   `group`,
		Short: `interact with groups`,
		Long: `Interact with groups.
Acts on all groups by default, however you may restrict the groups that a command applies to by specifying IDs or labels via the flags listed below.`,
		Run: usage,
	}
)

func init() {
	cmdGroupColor.Flags().Uint16VarP(&flagGroupHue, `hue`, `H`, 0, `hue component of the HSBK color (0-65535)`)
	cmdGroupColor.Flags().Uint16VarP(&flagGroupSaturation, `saturation`, `S`, 0, `saturation component of the HSBK color (0-65535)`)
	cmdGroupColor.Flags().Uint16VarP(&flagGroupBrightness, `brightness`, `B`, 0, `brightness component of the HSBK color (0-65535)`)
	cmdGroupColor.Flags().Uint16VarP(&flagGroupKelvin, `kelvin`, `K`, 0, `kelvin component of the HSBK color, the color temperature of whites (2500-9000)`)
	if err := cmdGroupColor.MarkFlagRequired(`hue`); err != nil {
		logger.WithField(`error`, err).Panicln(`Failed initializing application`)
	}
	if err := cmdGroupColor.MarkFlagRequired(`saturation`); err != nil {
		logger.WithField(`error`, err).Panicln(`Failed initializing application`)
	}
	if err := cmdGroupColor.MarkFlagRequired(`brightness`); err != nil {
		logger.WithField(`error`, err).Panicln(`Failed initializing application`)
	}
	if err := cmdGroupColor.MarkFlagRequired(`kelvin`); err != nil {
		logger.WithField(`error`, err).Panicln(`Failed initializing application`)
	}
	cmdGroup.AddCommand(cmdGroupList)
	cmdGroup.AddCommand(cmdGroupColor)
	cmdGroup.AddCommand(cmdGroupPower)

	cmdGroup.PersistentFlags().StringSliceVarP(&flagGroupIDs, `id`, `i`, make([]string, 0), `ID of the group(s) to manage, comma-seprated.  Defaults to all groups`)
	cmdGroup.PersistentFlags().StringSliceVarP(&flagGroupLabels, `label`, `l`, make([]string, 0), `label of the group(s) to manage, comma-separated.  Defaults to all groups.`)
	cmdGroup.PersistentFlags().DurationVarP(&flagGroupDuration, `duration`, `d`, 0*time.Second, `duration of the power/color transition`)
}

func groupList(c *cobra.Command, args []string) {
	var (
		err     error
		groups  []common.Group
		timeout <-chan time.Time
	)

	if flagTimeout == 0 {
		logger.Fatalln(`Can not list with a timeout of zero`)
	}

	if len(flagGroupIDs) > 0 || len(flagGroupLabels) > 0 {
		groups = getGroups()
	} else {
		// Because groups may appear before they're populated with lights, group
		// commands must all wait on the timeout
		timeout = time.After(flagTimeout)
		<-timeout

		groups, err = client.GetGroups()
		if err == common.ErrNotFound {
			logger.Fatalln(`No groups found`)
		} else if err != nil {
			logger.WithField(`error`, err).Fatalln(`Could not find groups`)
		}
	}

	table := new(tabwriter.Writer)
	table.Init(os.Stdout, 0, 4, 4, ' ', 0)
	fmt.Fprintf(table, fmt.Sprintf("%s\t%s\t%s\t%s\t%s\n", `ID`, `Label`, `Power`, `Color`, `Devices`))

	for _, g := range groups {
		var deviceLabels []string

		label := g.GetLabel()
		if err != nil {
			logger.WithFields(logrus.Fields{
				`group-id`: g.ID(),
				`error`:    err,
			}).Warnln(`Couldn't get color for group`)
			continue
		}
		power, err := g.GetPower()
		if err != nil {
			logger.WithFields(logrus.Fields{
				`group-id`: g.ID(),
				`error`:    err,
			}).Warnln(`Couldn't get color for group`)
			continue
		}
		color, err := g.GetColor()
		if err != nil {
			logger.WithFields(logrus.Fields{
				`group-id`: g.ID(),
				`error`:    err,
			}).Warnln(`Couldn't get color for group`)
			continue
		}
		for _, dev := range g.Devices() {
			l, err := dev.GetLabel()
			if err != nil {
				continue
			}
			deviceLabels = append(deviceLabels, l)
		}
		fmt.Fprintf(table, "%v\t%s\t%v\t%+v\t%+v\n", g.ID(), label, power, color, fmt.Sprintf("[%v]", strings.Join(deviceLabels, `, `)))
	}
	fmt.Fprintln(table)
	if err := table.Flush(); err != nil {
		logger.WithField(`error`, err).Fatalln(`Failed outputting results`)
	}
}

func getGroups() []common.Group {
	var (
		groups  []common.Group
		timeout <-chan time.Time
	)

	logger.WithField(`ids`, flagGroupLabels).Debug(`Requested IDs`)
	logger.WithField(`labels`, flagGroupLabels).Debug(`Requested labels`)

	// Because groups may appear before they're populated with lights, group
	// commands must all wait on the timeout
	timeout = time.After(flagTimeout)
	<-timeout

	if len(flagGroupIDs) > 0 {
		for _, id := range flagGroupIDs {
			group, err := client.GetGroupByID(id)
			if err != nil {
				logger.WithFields(logrus.Fields{
					`error`: err,
					`ID`:    id,
				}).Fatalln(`Could not find group with requested ID`)
			}
			groups = append(groups, group)
		}
	}
	if len(flagGroupLabels) > 0 {
		for _, label := range flagGroupLabels {
			group, err := client.GetGroupByLabel(label)
			if err != nil {
				logger.WithFields(logrus.Fields{
					`error`: err,
					`label`: label,
				}).Fatalln(`Could not find group with requested label`)
			}
			groups = append(groups, group)
		}
	}

	return groups
}

func groupPower(c *cobra.Command, args []string) {
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

	groups := getGroups()

	if len(groups) > 0 {
		for _, group := range groups {
			if err := group.SetPowerDuration(state, flagGroupDuration); err != nil {
				logger.WithFields(logrus.Fields{
					`group-id`: group.ID(),
					`error`:    err,
				}).Fatalln(`Failed setting power for group`)
			}
		}
	} else {
		if err := client.SetPowerDuration(state, flagGroupDuration); err != nil {
			logger.WithField(`error`, err).Fatalln(`Failed setting power for groups`)
		}
	}
}

func groupColor(c *cobra.Command, args []string) {
	if flagGroupHue == 0 && flagGroupSaturation == 0 && flagGroupBrightness == 0 && flagGroupKelvin == 0 {
		if err := c.Usage(); err != nil {
			logger.WithField(`error`, err).Fatalln(`Failed to print usage`)
		}
		fmt.Println()
		logger.Fatalln(`Missing color definition`)
	}

	groups := getGroups()

	color := common.Color{
		Hue:        flagGroupHue,
		Saturation: flagGroupSaturation,
		Brightness: flagGroupBrightness,
		Kelvin:     flagGroupKelvin,
	}

	if len(groups) > 0 {
		for _, group := range groups {
			if err := group.SetColor(color, flagGroupDuration); err != nil {
				logger.WithFields(logrus.Fields{
					`group-id`: group.ID(),
					`error`:    err,
				}).Fatalln(`Failed setting color for group`)
			}
		}
	} else {
		if err := client.SetColor(color, flagGroupDuration); err != nil {
			logger.WithField(`error`, err).Fatalln(`Failed setting color for groups`)
		}
	}
}

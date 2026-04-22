package cmd

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

func newGuideCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "guide",
		Short: "Render live guide packets for guided planning stages",
	}

	var currentFormat string
	current := &cobra.Command{
		Use:   "current",
		Short: "Render the guide packet for the last-active guided session",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			packet, err := planningManager().CurrentGuidePacket()
			if err != nil {
				return err
			}
			return writeGuidePacket(cmd, currentFormat, packet)
		},
	}
	current.Flags().StringVar(&currentFormat, "format", "json", "output format: json")

	var (
		showFormat     string
		showChain      string
		showStage      string
		showCheckpoint string
	)
	show := &cobra.Command{
		Use:   "show",
		Short: "Render the guide packet for an explicit guided session chain",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			packet, err := planningManager().GuidePacketForChain(showChain, showStage, showCheckpoint)
			if err != nil {
				return err
			}
			return writeGuidePacket(cmd, showFormat, packet)
		},
	}
	show.Flags().StringVar(&showChain, "chain", "", "guided session chain id, such as brainstorm/my-topic")
	show.Flags().StringVar(&showStage, "stage", "", "explicit stage override")
	show.Flags().StringVar(&showCheckpoint, "checkpoint", "", "explicit checkpoint override")
	show.Flags().StringVar(&showFormat, "format", "json", "output format: json")
	_ = show.MarkFlagRequired("chain")

	cmd.AddCommand(current, show)
	return cmd
}

func writeGuidePacket(cmd *cobra.Command, format string, packet any) error {
	if strings.TrimSpace(format) != "json" {
		return fmt.Errorf("unsupported guide output format %q; only json is supported in v1", format)
	}
	encoder := json.NewEncoder(cmd.OutOrStdout())
	encoder.SetIndent("", "  ")
	return encoder.Encode(packet)
}

package cmd

import (
	"encoding/json"
	"fmt"
	"strings"

	brainapp "github.com/JimmyMcBride/brain/planning/application"
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
			if handled, err := withSharedPlanning(cmd, "guide current", true, func(service *brainapp.Service) error {
				packet, err := service.CurrentGuidePacket(cmd.Context())
				if err != nil {
					return err
				}
				return writeGuidePacket(cmd, currentFormat, packet)
			}); handled {
				return err
			}
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
		showBrainstorm string
		showDiscussion string
		showStage      string
		showCheckpoint string
	)
	show := &cobra.Command{
		Use:   "show",
		Short: "Render a guide packet for a guided brainstorm chain or collaboration source",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			switch {
			case strings.TrimSpace(showChain) != "":
				if strings.TrimSpace(showBrainstorm) != "" || strings.TrimSpace(showDiscussion) != "" {
					return fmt.Errorf("choose either --chain or one collaboration source flag")
				}
				if (strings.TrimSpace(showStage) == "" || strings.TrimSpace(showStage) == "brainstorm") && sharedGuideCheckpoint(showCheckpoint) {
					if handled, err := withSharedPlanning(cmd, "guide show", true, func(service *brainapp.Service) error {
						packet, err := service.GuidePacketForChain(cmd.Context(), showChain, showCheckpoint)
						if err != nil {
							return err
						}
						return writeGuidePacket(cmd, showFormat, packet)
					}); handled {
						return err
					}
				}
				packet, err := planningManager().GuidePacketForChain(showChain, showStage, showCheckpoint)
				if err != nil {
					return err
				}
				return writeGuidePacket(cmd, showFormat, packet)
			case strings.TrimSpace(showBrainstorm) != "" || strings.TrimSpace(showDiscussion) != "":
				if strings.TrimSpace(showCheckpoint) != "" {
					return fmt.Errorf("--checkpoint only applies to --chain guide previews")
				}
				if strings.TrimSpace(showBrainstorm) != "" && strings.TrimSpace(showDiscussion) == "" && (strings.TrimSpace(showStage) == "" || strings.TrimSpace(showStage) == "brainstorm") {
					if handled, err := withSharedPlanning(cmd, "guide show", true, func(service *brainapp.Service) error {
						packet, err := service.GuidePacketForChain(cmd.Context(), "brainstorm/"+strings.TrimSpace(showBrainstorm), "")
						if err != nil {
							return err
						}
						return writeGuidePacket(cmd, showFormat, packet)
					}); handled {
						return err
					}
				}
				packet, err := planningManager().GuidePacketForCollaborationSource(showBrainstorm, showDiscussion, showStage)
				if err != nil {
					return err
				}
				return writeGuidePacket(cmd, showFormat, packet)
			default:
				return fmt.Errorf("guide show requires --chain, --brainstorm, or --discussion")
			}
		},
	}
	show.Flags().StringVar(&showChain, "chain", "", "guided session chain id, such as brainstorm/my-topic")
	show.Flags().StringVar(&showBrainstorm, "brainstorm", "", "local brainstorm slug for collaboration-stage guide packets")
	show.Flags().StringVar(&showDiscussion, "discussion", "", "GitHub Discussion number or URL for collaboration-stage guide packets")
	show.Flags().StringVar(&showStage, "stage", "", "explicit stage override")
	show.Flags().StringVar(&showCheckpoint, "checkpoint", "", "explicit checkpoint override for --chain previews")
	show.Flags().StringVar(&showFormat, "format", "json", "output format: json")

	cmd.AddCommand(current, show)
	return cmd
}

func sharedGuideCheckpoint(checkpoint string) bool {
	switch strings.TrimSpace(checkpoint) {
	case "", "vision-intake", "clarify-problem-user-value", "clarify-constraints-appetite", "clarify-open-approaches", "handoff-epic":
		return true
	default:
		return false
	}
}

func writeGuidePacket(cmd *cobra.Command, format string, packet any) error {
	if strings.TrimSpace(format) != "json" {
		return fmt.Errorf("unsupported guide output format %q; only json is supported", format)
	}
	encoder := json.NewEncoder(cmd.OutOrStdout())
	encoder.SetIndent("", "  ")
	return encoder.Encode(packet)
}

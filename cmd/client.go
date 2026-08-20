package cmd

import (
	"io"
	"strings"

	"github.com/spf13/cobra"

	"github.com/lemonade-command/lemonade/client"
	"github.com/lemonade-command/lemonade/lemon"
)

func newCopyCmd(cfg *lemon.Config) *cobra.Command {
	return &cobra.Command{
		Use:   "copy [text]",
		Short: "Copy text.",
		Args:  cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			text := strings.Join(args, " ")
			if text == "" {
				b, err := io.ReadAll(cmd.InOrStdin())
				if err != nil {
					return err
				}
				text = string(b)
			}

			c := client.New(client.Options{
				Host:       cfg.Host,
				Port:       cfg.Port,
				LineEnding: cfg.LineEnding,
				Logger:     lemon.NewLogger(*cfg, cmd.ErrOrStderr()),
			})
			return c.Copy(text)
		},
	}
}

func newPasteCmd(cfg *lemon.Config) *cobra.Command {
	return &cobra.Command{
		Use:   "paste",
		Short: "Paste text.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			c := client.New(client.Options{
				Host:       cfg.Host,
				Port:       cfg.Port,
				LineEnding: cfg.LineEnding,
				Logger:     lemon.NewLogger(*cfg, cmd.ErrOrStderr()),
			})
			text, err := c.Paste()
			if err != nil {
				return err
			}
			_, err = io.WriteString(cmd.OutOrStdout(), text)
			return err
		},
	}
}

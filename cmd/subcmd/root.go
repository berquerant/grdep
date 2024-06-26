package subcmd

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/signal"

	"github.com/berquerant/grdep"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.PersistentFlags().Bool("debug", false, "Enable debug logs")
	rootCmd.PersistentFlags().Bool("metrics", false, "Show metrics")
}

func getDebug(cmd *cobra.Command) bool {
	debug, _ := cmd.Flags().GetBool("debug")
	return debug
}

func getMetrics(cmd *cobra.Command) bool {
	enableMetrics, _ := cmd.Flags().GetBool("metrics")
	return enableMetrics
}

var rootCmd = &cobra.Command{
	Use:   "grdep",
	Short: "Find dependencies by grep.",
	PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
		if getMetrics(cmd) {
			grdep.InitMetrics()
		}
		return nil
	},
	PersistentPostRunE: func(cmd *cobra.Command, _ []string) error {
		if getMetrics(cmd) {
			showMetrics()
		}
		return nil
	},
}

func Execute() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	if err := rootCmd.ExecuteContext(ctx); err != nil {
		slog.Error("execute", "err", err)
		return err
	}
	return nil
}

func showMetrics() {
	grdep.GetMetrics().Close()
	grdep.GetMetrics().Walk(func(key string, value *grdep.GaugeMapElement) {
		grdep.L().Info("metrics", "key", key, "count", value.Count, "duration_ms", value.Duration.Milliseconds())
	})
}

func getLogger(cmd *cobra.Command, w io.Writer) *slog.Logger {
	debug := getDebug(cmd)
	logLevel := slog.LevelInfo
	if debug {
		logLevel = slog.LevelDebug
	}
	return grdep.NewLogger(w, logLevel)
}

var (
	errInvalidArgument = errors.New("InvalidArgument")
	errNoConfigFiles   = errors.New("NoConfigFiles")
)

func parseConfigs(configs []string) (*grdep.Config, error) {
	if len(configs) == 0 {
		return nil, errNoConfigFiles
	}

	var result grdep.Config
	for i, config := range configs {
		c, err := grdep.ParseConfig(config)
		if err != nil {
			return nil, fmt.Errorf("%w: config[%d] %s", err, i, config)
		}
		result = result.Add(*c)
	}
	return &result, nil
}

package logs

import (
	"io"
	"os"
	"strings"

	"github.com/Polad20/urlshortener/config"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func InitLogger(cfg config.Logger) {
	var level zerolog.Level
	switch strings.ToLower(cfg.Level) {
	case "debug":
		level = zerolog.DebugLevel
	case "info":
		level = zerolog.InfoLevel
	case "warn":
		level = zerolog.WarnLevel
	case "error":
		level = zerolog.ErrorLevel
	case "fatal":
		level = zerolog.FatalLevel
	case "panic":
		level = zerolog.PanicLevel
	}
	zerolog.SetGlobalLevel(level)
	var output io.Writer = os.Stdout
	if cfg.LogPath != "" {
		file, err := os.OpenFile(cfg.LogPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			log.Err(err).Msg("Failed to open log file, writing to stdout")
		} else {
			output = file
		}
		defer file.Close()
	}
	var writer zerolog.Logger
	if cfg.Format == "json" {
		writer = zerolog.New(output).With().Timestamp().Logger()
	} else {
		consoleWriter := zerolog.ConsoleWriter{Out: output, TimeFormat: zerolog.TimeFormatUnix}
		writer = zerolog.New(consoleWriter).With().Timestamp().Logger()
	}

	log.Logger = writer
}

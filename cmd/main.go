package main

import (
	"fastcopy/internal/fastcopy"
	"flag"
	"log/slog"
	"os"
)

func main() {
	var from, to, filter string
	flag.StringVar(&from, "from", "./test/from", "specify source path")
	flag.StringVar(&to, "to", "./test/to", "specify destination path")
	flag.StringVar(&filter, "filter", ".*\\.txt$", "specify file filter using regex") //.*\.dll$
	flag.Parse()

	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})))

	res, err := fastcopy.Copy(from, to, filter)
	if err == nil {
		slog.Info("Files copied with NO errors", slog.Int("number of files", res))
	}
}

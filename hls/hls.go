package hls

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
	"time"
)

var (
	src      string
	duration int
	output   string
)

var HlsCmd = &cobra.Command{
	Use:   "hls",
	Short: "Measure latency of hls protocol",
	Long:  "Measure latency of hls protocol",
	Run: func(cmd *cobra.Command, args []string) {
		downloader := NewDownloader(src)
		traceWriter := NewTraceWriter()

		defer func() {
			if len(traceWriter.Entries) == 0 {
				return
			}

			traceWriter.CalculateSummaries()
			if payload, err := traceWriter.Serialize(); err == nil {
				err := os.WriteFile(output, payload, 0644)
				if err != nil {
					fmt.Println("Failed to write trace writer's result to file")
				}
			} else {
				fmt.Println("Failed to serialize the trace writer")
			}
		}()

		renditions, trace, err := downloader.DownloadMastermanifest()
		if err != nil {
			fmt.Println(err)
			return
		}

		downloader.SetRenditions(renditions)

		playlist, trace, _ := downloader.DownloadPlaylistByName(renditions[0].Name)

		startTime := time.Now()

		for int(time.Now().Sub(startTime).Seconds()) < duration {
			nextSeqNum := playlist.NextSeqNum()
			playlist, trace, err = downloader.RefreshPlaylistWithBlockingReload(playlist)
			if err != nil {
				fmt.Printf("Error while performing blocking playlist reload %+v\n", err)
				return
			}
			traceWriter.Write("manifest", trace)

			lastMediaItem := playlist.LastMediaItem()
			if item, ok := lastMediaItem.(*PartSegment); ok {
				if item.SequenceNumbers.GreaterOrEqual(&nextSeqNum) {
					_, trace, err = downloader.DownloadMedia(playlist, item)
					if err != nil {
						fmt.Printf("Error while downloading media segment %+v\n", err)
						return
					}
					traceWriter.Write("mediaSegment", trace)
				} else {
					fmt.Errorf("Latest media segment doesn't match the expected sequence numbers", trace)
					return
				}
			}
		}
	},
}

func init() {
	HlsCmd.Flags().StringVarP(&src, "src", "s", "", "source url of the target stream")
	HlsCmd.MarkFlagRequired("src")
	HlsCmd.Flags().IntVarP(&duration, "duration", "d", 10, "duration of the latency test")
	HlsCmd.MarkFlagRequired("duration")
	HlsCmd.Flags().StringVarP(&output, "output", "o", "", "output file for latency measurements")
}

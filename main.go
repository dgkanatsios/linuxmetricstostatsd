package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/load"
	"github.com/shirou/gopsutil/mem"
	"github.com/shirou/gopsutil/net"

	statsd "github.com/cactus/go-statsd-client/statsd"
)

var interval int
var statsdhost string
var statsdport int
var statsdClient statsd.Statter

func main() {
	interval = *flag.Int("interval", 3, "Metrics polling interval in seconds. Default is 3")
	statsdhost = *flag.String("statsdhost", "127.0.0.1", "StatsD host")
	statsdport = *flag.Int("statsdport", 8125, "StatsD port")

	config := &statsd.ClientConfig{
		Address:     fmt.Sprintf("%s:%d", statsdhost, statsdport),
		Prefix:      "linuxmetricstostatsd",
		UseBuffered: true,
		// interval to force flush buffer. full buffers will flush on their own,
		// but for data not frequently sent, a max threshold is useful
		FlushInterval: 1 * time.Second,
	}
	var err error
	statsdClient, err = statsd.NewClientWithConfig(config)
	if err != nil {
		log.Fatal(err)
	}

	c1, cancel := context.WithCancel(context.Background())
	exitCh := make(chan struct{})
	go func(ctx context.Context) {
		for {
			fmt.Println("In loop. Press ^C to stop.")
			err = collectAndSendMetrics()
			if err != nil {
				log.Printf("Cannot collect and send metrics due to %#v at %s\n", err, time.Now())
			}
			time.Sleep(time.Duration(interval) * (time.Second))
			select {
			case <-ctx.Done():
				fmt.Println("received cltr-c, exiting")
				if statsdClient != nil {
					statsdClient.Close()
				}
				exitCh <- struct{}{}
				return
			default:
			}
		}
	}(c1)

	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Interrupt)
	go func() {
		select {
		case <-signalCh:
			cancel()
			return
		}
	}()
	<-exitCh
}

func collectAndSendMetrics() error {

	virtualMemory, loadAvg, cpuPercentage, ioNetStats, err := getMetrics()
	if err != nil {
		return err
	}

	err = statsdClient.Gauge("VirtualMemory", int64(virtualMemory.Used), float32(interval))
	if err != nil {
		return err
	}
	err = statsdClient.Gauge("LoadAverage1", int64(loadAvg.Load1*100), float32(interval))
	if err != nil {
		return err
	}
	err = statsdClient.Gauge("LoadAverage5", int64(loadAvg.Load5*100), float32(interval))
	if err != nil {
		return err
	}
	err = statsdClient.Gauge("LoadAverage15", int64(loadAvg.Load15*100), float32(interval))
	if err != nil {
		return err
	}
	err = statsdClient.Gauge("CPUPercentage", int64(cpuPercentage), float32(interval))
	if err != nil {
		return err
	}
	err = statsdClient.Gauge("NetReceive", int64(ioNetStats[0].BytesRecv), float32(interval))
	if err != nil {
		return err
	}
	err = statsdClient.Gauge("NetSend", int64(ioNetStats[0].BytesSent), float32(interval))
	if err != nil {
		return err
	}

	return nil
}

func getMetrics() (*mem.VirtualMemoryStat, *load.AvgStat, float64, []net.IOCountersStat, error) {
	v, _ := mem.VirtualMemory()
	// almost every return value is a struct
	fmt.Printf("Total: %v, Free:%v, UsedPercent:%f%%\n", v.Total, v.Free, v.UsedPercent)

	c, _ := cpu.Percent(time.Duration(interval)*time.Second, false)
	// we're getting a single CPU average
	fmt.Printf("CPU %f \n", c[0])

	l, _ := load.Avg()
	fmt.Printf("Load average %v\n", l)

	ioNetStats, _ := net.IOCounters(false)
	for _, ioNetStat := range ioNetStats {
		fmt.Printf("stats:%v\n", ioNetStat)
	}

	return v, l, c[0], ioNetStats, nil
}

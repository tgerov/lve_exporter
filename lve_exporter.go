// LVE Exporter - A Prometheus exporter which scrapes metrics from CloudLinux LVE Stats 2
// Copyright (C) 2021 Tsvetan Gerov <tsvetan@gerov.eu>
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"gopkg.in/alecthomas/kingpin.v2"
)

var version = "1.0"

type LVEStats struct {
	Result    string  `json:"result"`
	Timestamp float64 `json:"timestamp"`
	Users     []struct {
		Usage struct {
			CPU struct {
				Lve float64 `json:"lve"`
			} `json:"cpu"`
			Ep struct {
				Lve float64 `json:"lve"`
			} `json:"ep"`
			Vmem struct {
				Lve float64 `json:"lve"`
			} `json:"vmem"`
			Pmem struct {
				Lve float64 `json:"lve"`
			} `json:"pmem"`
			Nproc struct {
				Lve float64 `json:"lve"`
			} `json:"nproc"`
			Io struct {
				Lve float64 `json:"lve"`
			} `json:"io"`
			Iops struct {
				Lve float64 `json:"lve"`
			} `json:"iops"`
		} `json:"usage"`
		Limits struct {
			CPU struct {
				Lve float64 `json:"lve"`
			} `json:"cpu"`
			Ep struct {
				Lve float64 `json:"lve"`
			} `json:"ep"`
			Vmem struct {
				Lve float64 `json:"lve"`
			} `json:"vmem"`
			Pmem struct {
				Lve float64 `json:"lve"`
			} `json:"pmem"`
			Nproc struct {
				Lve float64 `json:"lve"`
			} `json:"nproc"`
			Io struct {
				Lve float64 `json:"lve"`
			} `json:"io"`
			Iops struct {
				Lve float64 `json:"lve"`
			} `json:"iops"`
		} `json:"limits"`
		Faults struct {
			CPU struct {
				Lve float64 `json:"lve"`
			} `json:"cpu"`
			Ep struct {
				Lve float64 `json:"lve"`
			} `json:"ep"`
			Vmem struct {
				Lve float64 `json:"lve"`
			} `json:"vmem"`
			Pmem struct {
				Lve float64 `json:"lve"`
			} `json:"pmem"`
			Nproc struct {
				Lve float64 `json:"lve"`
			} `json:"nproc"`
			Io struct {
				Lve float64 `json:"lve"`
			} `json:"io"`
			Iops struct {
				Lve float64 `json:"lve"`
			} `json:"iops"`
		} `json:"faults"`
		ID       float64 `json:"id"`
		Username string  `json:"username"`
		Domain   string  `json:"domain"`
		Reseller string  `json:"reseller"`
	} `json:"users"`
	Resellers []interface{} `json:"resellers"`
	MySQLGov  string        `json:"mySqlGov"`
}

type lveCollector struct {
	cpuUsageMetric  *prometheus.Desc
	cpuLimitMetric  *prometheus.Desc
	cpuFaultsMetric *prometheus.Desc

	pmemUsageMetric  *prometheus.Desc
	pmemLimitMetric  *prometheus.Desc
	pmemFaultsMetric *prometheus.Desc

	vmemUsageMetric  *prometheus.Desc
	vmemLimitMetric  *prometheus.Desc
	vmemFaultsMetric *prometheus.Desc

	nprocUsageMetric  *prometheus.Desc
	nprocLimitMetric  *prometheus.Desc
	nprocFaultsMetric *prometheus.Desc

	ioUsageMetric  *prometheus.Desc
	ioLimitMetric  *prometheus.Desc
	ioFaultsMetric *prometheus.Desc

	iopsUsageMetric  *prometheus.Desc
	iopsLimitMetric  *prometheus.Desc
	iopsFaultsMetric *prometheus.Desc

	epUsageMetric  *prometheus.Desc
	epLimitMetric  *prometheus.Desc
	epFaultsMetric *prometheus.Desc
}

func newLveCollector() *lveCollector {
	return &lveCollector{
		cpuUsageMetric:    prometheus.NewDesc("LVE_CPU_USAGE", "CPU usage per user in LVE", []string{"username"}, nil),
		cpuLimitMetric:    prometheus.NewDesc("LVE_CPU_LIMIT", "CPU limit per user in LVE", []string{"username"}, nil),
		cpuFaultsMetric:   prometheus.NewDesc("LVE_CPU_FAULTS", "CPU limit per user in LVE", []string{"username"}, nil),
		pmemUsageMetric:   prometheus.NewDesc("LVE_PMEM_USAGE", "Physical memory usage per user in LVE", []string{"username"}, nil),
		pmemLimitMetric:   prometheus.NewDesc("LVE_PMEM_LIMIT", "Phisical memory limit per user in LVE", []string{"username"}, nil),
		pmemFaultsMetric:  prometheus.NewDesc("LVE_PMEM_FAULTS", "Phisical memory faults per user in LVE", []string{"username"}, nil),
		vmemUsageMetric:   prometheus.NewDesc("LVE_VMEM_USAGE", "Virtual memory usage per user in LVE", []string{"username"}, nil),
		vmemLimitMetric:   prometheus.NewDesc("LVE_VMEM_LIMIT", "Virtual memory limit per user in LVE", []string{"username"}, nil),
		vmemFaultsMetric:  prometheus.NewDesc("LVE_VMEM_FAULTS", "Virtual memory faults per user in LVE", []string{"username"}, nil),
		nprocUsageMetric:  prometheus.NewDesc("LVE_NPROC_USAGE", "Nummber of processes per LVE user", []string{"username"}, nil),
		nprocLimitMetric:  prometheus.NewDesc("LVE_NPROC_LIMIT", "Limit for Nummber of processes per LVE user", []string{"username"}, nil),
		nprocFaultsMetric: prometheus.NewDesc("LVE_NPROC_FAULTS", "Faults for Nummber of processes per LVE user", []string{"username"}, nil),
		ioUsageMetric:     prometheus.NewDesc("LVE_IO_USAGE", "IO per LVE user", []string{"username"}, nil),
		ioLimitMetric:     prometheus.NewDesc("LVE_IO_LIMIT", "IO Limit per LVE user", []string{"username"}, nil),
		ioFaultsMetric:    prometheus.NewDesc("LVE_IO_FAULTS", "IO Faults per LVE user", []string{"username"}, nil),
		iopsUsageMetric:   prometheus.NewDesc("LVE_IOPS_USAGE", "IOPS per LVE user", []string{"username"}, nil),
		iopsLimitMetric:   prometheus.NewDesc("LVE_IOPS_LIMIT", "IOPS Limit per LVE user", []string{"username"}, nil),
		iopsFaultsMetric:  prometheus.NewDesc("LVE_IOPS_FAULTS", "IOPS Faults per LVE user", []string{"username"}, nil),
		epUsageMetric:     prometheus.NewDesc("LVE_EP_USAGE", "Entry Processes per LVE user", []string{"username"}, nil),
		epLimitMetric:     prometheus.NewDesc("LVE_EP_LIMIT", "Entry Processes Limit per LVE user", []string{"username"}, nil),
		epFaultsMetric:    prometheus.NewDesc("LVE_EP_FAULTS", "Entry Processes Faults per LVE user", []string{"username"}, nil),
	}
}

func (collector *lveCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- collector.cpuUsageMetric
	ch <- collector.cpuLimitMetric
	ch <- collector.cpuFaultsMetric
	ch <- collector.pmemUsageMetric
	ch <- collector.pmemLimitMetric
	ch <- collector.pmemFaultsMetric
	ch <- collector.vmemUsageMetric
	ch <- collector.vmemLimitMetric
	ch <- collector.vmemFaultsMetric
	ch <- collector.nprocUsageMetric
	ch <- collector.nprocLimitMetric
	ch <- collector.nprocFaultsMetric
	ch <- collector.ioUsageMetric
	ch <- collector.ioLimitMetric
	ch <- collector.ioFaultsMetric
	ch <- collector.iopsUsageMetric
	ch <- collector.iopsLimitMetric
	ch <- collector.iopsFaultsMetric
	ch <- collector.epUsageMetric
	ch <- collector.epLimitMetric
	ch <- collector.epFaultsMetric
}

func (collector *lveCollector) Collect(ch chan<- prometheus.Metric) {
	jsonFile, err := exec.Command("sudo", "/usr/sbin/cloudlinux-statistics", "--json").Output()
	if err != nil {
		fmt.Println(err)
	}
	var lvestats LVEStats
	json.Unmarshal(jsonFile, &lvestats)

	for i := 0; i < len(lvestats.Users); i++ {
		var username = lvestats.Users[i].Username
		// LVE: CPU
		ch <- prometheus.MustNewConstMetric(collector.cpuUsageMetric, prometheus.GaugeValue, lvestats.Users[i].Usage.CPU.Lve, username)
		ch <- prometheus.MustNewConstMetric(collector.cpuLimitMetric, prometheus.GaugeValue, lvestats.Users[i].Limits.CPU.Lve, username)
		ch <- prometheus.MustNewConstMetric(collector.cpuFaultsMetric, prometheus.GaugeValue, lvestats.Users[i].Faults.CPU.Lve, username)
		// LVE: Physical Memory
		ch <- prometheus.MustNewConstMetric(collector.pmemUsageMetric, prometheus.GaugeValue, lvestats.Users[i].Usage.Pmem.Lve, username)
		ch <- prometheus.MustNewConstMetric(collector.pmemLimitMetric, prometheus.GaugeValue, lvestats.Users[i].Limits.Pmem.Lve, username)
		ch <- prometheus.MustNewConstMetric(collector.pmemFaultsMetric, prometheus.GaugeValue, lvestats.Users[i].Faults.Pmem.Lve, username)
		// LVE: Virtual Memory
		ch <- prometheus.MustNewConstMetric(collector.vmemUsageMetric, prometheus.GaugeValue, lvestats.Users[i].Usage.Vmem.Lve, username)
		ch <- prometheus.MustNewConstMetric(collector.vmemLimitMetric, prometheus.GaugeValue, lvestats.Users[i].Limits.Vmem.Lve, username)
		ch <- prometheus.MustNewConstMetric(collector.vmemFaultsMetric, prometheus.GaugeValue, lvestats.Users[i].Faults.Vmem.Lve, username)
		// LVE: Number of processes
		ch <- prometheus.MustNewConstMetric(collector.nprocUsageMetric, prometheus.GaugeValue, lvestats.Users[i].Usage.Nproc.Lve, username)
		ch <- prometheus.MustNewConstMetric(collector.nprocLimitMetric, prometheus.GaugeValue, lvestats.Users[i].Limits.Nproc.Lve, username)
		ch <- prometheus.MustNewConstMetric(collector.nprocFaultsMetric, prometheus.GaugeValue, lvestats.Users[i].Faults.Nproc.Lve, username)
		// LVE: IO
		ch <- prometheus.MustNewConstMetric(collector.ioUsageMetric, prometheus.GaugeValue, lvestats.Users[i].Usage.Io.Lve, username)
		ch <- prometheus.MustNewConstMetric(collector.ioLimitMetric, prometheus.GaugeValue, lvestats.Users[i].Limits.Io.Lve, username)
		ch <- prometheus.MustNewConstMetric(collector.ioFaultsMetric, prometheus.GaugeValue, lvestats.Users[i].Faults.Io.Lve, username)
		// LVE: IOPS
		ch <- prometheus.MustNewConstMetric(collector.iopsUsageMetric, prometheus.GaugeValue, lvestats.Users[i].Usage.Iops.Lve, username)
		ch <- prometheus.MustNewConstMetric(collector.iopsLimitMetric, prometheus.GaugeValue, lvestats.Users[i].Limits.Iops.Lve, username)
		ch <- prometheus.MustNewConstMetric(collector.iopsFaultsMetric, prometheus.GaugeValue, lvestats.Users[i].Faults.Iops.Lve, username)
		// LVE: Entry Proccesses
		ch <- prometheus.MustNewConstMetric(collector.epUsageMetric, prometheus.GaugeValue, lvestats.Users[i].Usage.Ep.Lve, username)
		ch <- prometheus.MustNewConstMetric(collector.epLimitMetric, prometheus.GaugeValue, lvestats.Users[i].Limits.Ep.Lve, username)
		ch <- prometheus.MustNewConstMetric(collector.epFaultsMetric, prometheus.GaugeValue, lvestats.Users[i].Faults.Ep.Lve, username)
	}
}

func main() {
	var (
		listenAddress = kingpin.Flag("web.listen-address",
			"Address to listen on for web interface and telemetry",
		).Default(":9119").String()
	)
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()

	lvestats := newLveCollector()
	prometheus.MustRegister(lvestats)

	fmt.Println("Starting lve_exporter version", version)
	fmt.Println("Serving requests on port", *listenAddress)
	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`<html>
			 <head><title>LVE Exporter</title></head>
			 <body>
			 <h1>LVE Exporter</h1>
			 <p><a href='` + "/metrics" + `'>Metrics</a></p>
			 </body>
			 </html>`))
	})
	http.ListenAndServe(*listenAddress, nil)

}

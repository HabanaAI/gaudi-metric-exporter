// Copyright (C) 2025 Intel Corporation

// This program is free software; you can redistribute it and/or modify it
// under the terms of the GNU General Public License version 2 or later, as published
// by the Free Software Foundation.

// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU General Public License for more details.

// You should have received a copy of the GNU General Public License
// along with this program; if not, see <http://www.gnu.org/licenses/>.

// SPDX-License-Identifier: GPL-2.0-or-later

package main

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

const (
	socClock = iota // = 0
	socClockMax
	icClockMax
	mmeClockMax
	tpcClockMax
	freeMem
	usedMem
	totalMem
	boardTemp
	chipTemp
	temperatureThresholdShutdown
	temperatureThresholdSlowdown
	temperatureThresholdMemory
	temperatureThresholdGPU
	power
	powerLimit
	utilization
	eccMode
	pcieTx
	pcieRx
	energy
	pcieReplay
	pciLinkSpeed
	pciLinkWidth
	pcieTX
	pcieRX
	doubleBitErrRows
	singleBitErrRows
	pendingRowsState
	numSamples
)

type SampleMetricMetadata struct {
	Name string
	Desc string
	Uid  uint64
	Type uint64
}

type Sample struct {
	Timestamp time.Time
	Value     float64
	UID       uint64
}

func SampleMetadata() []SampleMetricMetadata {
	meta := make([]SampleMetricMetadata, 0, numSamples)
	meta = append(
		meta,
		SampleMetricMetadata{
			Uid:  socClock,
			Name: "habanalabs_clock_soc_mhz",
			Desc: "Current frequency of the SOC",
			Type: 0, // Type 0 == sample
		},
		SampleMetricMetadata{
			Uid:  socClockMax,
			Name: "habanalabs_clock_soc_max_mhz",
			Desc: "Maximum frequency of the SOC",
			Type: 0, // Type 0 == sample
		},
		SampleMetricMetadata{
			Uid:  icClockMax,
			Name: "habanalabs_clock_ic_max_mhz",
			Desc: "Maximum frequency of the IC",
			Type: 0, // Type 0 == sample
		},
		SampleMetricMetadata{
			Uid:  mmeClockMax,
			Name: "habanalabs_clock_mme_max_mhz",
			Desc: "Maximum frequency of the MME",
			Type: 0, // Type 0 == sample
		},
		SampleMetricMetadata{
			Uid:  tpcClockMax,
			Name: "habanalabs_clock_tpc_max_mhz",
			Desc: "Maximum frequency of the TPC",
			Type: 0, // Type 0 == sample
		},
		SampleMetricMetadata{
			Uid:  freeMem,
			Name: "habanalabs_memory_free_bytes",
			Desc: "Current free bytes of memory",
			Type: 0,
		},
		SampleMetricMetadata{
			Uid:  usedMem,
			Name: "habanalabs_memory_used_bytes",
			Desc: "Current used bytes of memory",
			Type: 0,
		},
		SampleMetricMetadata{
			Uid:  totalMem,
			Name: "habanalabs_memory_total_bytes",
			Desc: "Current total bytes of memory",
			Type: 0,
		},
		SampleMetricMetadata{
			Uid:  boardTemp,
			Name: "habanalabs_temperature_onboard",
			Desc: "Board temperature in celsius",
			Type: 0,
		},
		SampleMetricMetadata{
			Uid:  chipTemp,
			Name: "habanalabs_temperature_onchip",
			Desc: "Chip temperature in celsius",
			Type: 0,
		},
		SampleMetricMetadata{
			Uid:  temperatureThresholdShutdown,
			Name: "habanalabs_temperature_threshold_shutdown",
			Desc: "Temperature threshold for shutdown",
			Type: 0,
		},
		SampleMetricMetadata{
			Uid:  temperatureThresholdSlowdown,
			Name: "habanalabs_temperature_threshold_slowdown",
			Desc: "Temperature threshold for slowdown",
			Type: 0,
		},
		SampleMetricMetadata{
			Uid:  temperatureThresholdMemory,
			Name: "habanalabs_temperature_threshold_memory",
			Desc: "Temperature threshold for memory",
			Type: 0,
		},
		SampleMetricMetadata{
			Uid:  temperatureThresholdGPU,
			Name: "habanalabs_temperature_threshold_gpu",
			Desc: "Temperature threshold for GPU",
			Type: 0,
		},
		SampleMetricMetadata{
			Uid:  power,
			Name: "habanalabs_power_mW",
			Desc: "Power usage in milli-watts",
			Type: 0,
		},
		SampleMetricMetadata{
			Uid:  powerLimit,
			Name: "habanalabs_power_default_limit_mW",
			Desc: "Default Power usage limit in milli-watts",
			Type: 0,
		},
		SampleMetricMetadata{
			Uid:  utilization,
			Name: "habanalabs_utilization",
			Desc: "Utilization of the device",
			Type: 0,
		},
		SampleMetricMetadata{
			Uid:  eccMode,
			Name: "habanalabs_ecc_feature_mode",
			Desc: "ECC feature mode",
			Type: 0,
		},
		SampleMetricMetadata{
			Uid:  pcieTx,
			Name: "habanalabs_pcie_tx",
			Desc: "PCIe transmit traffic",
			Type: 0,
		},
		SampleMetricMetadata{
			Uid:  pcieRx,
			Name: "habanalabs_pcie_rx",
			Desc: "PCIe receive traffic",
			Type: 0,
		},
		SampleMetricMetadata{
			Uid:  energy,
			Name: "habanalabs_energy",
			Desc: "Device Energy Usage",
			Type: 0, // Type 1 == counter
		},
		SampleMetricMetadata{
			Uid:  pcieReplay,
			Name: "habanalabs_pcie_replay_count",
			Desc: "PCIe replay count",
			Type: 0,
		},
		SampleMetricMetadata{
			Uid:  pciLinkSpeed,
			Name: "habanalabs_pci_link_speed",
			Desc: "Current PCI Link Speed",
			Type: 0,
		},
		SampleMetricMetadata{
			Uid:  pciLinkWidth,
			Name: "habanalabs_pci_link_width",
			Desc: "Current PCI link width",
			Type: 0,
		},
		SampleMetricMetadata{
			Uid:  pcieTX,
			Name: "habanalabs_pcie_transmit_throughput",
			Desc: "PCIe transmit throughput",
			Type: 0,
		},
		SampleMetricMetadata{
			Uid:  pcieRX,
			Name: "habanalabs_pcie_receive_throughput",
			Desc: "PCIe receive throughput",
			Type: 0,
		},
		SampleMetricMetadata{
			Uid:  doubleBitErrRows,
			Name: "habanalabs_pending_rows_with_double_bit_ecc_errors",
			Desc: "Counter for rows with Double-Bit ECC errors",
			Type: 0,
		},
		SampleMetricMetadata{
			Uid:  singleBitErrRows,
			Name: "habanalabs_pending_rows_with_single_bit_ecc_errors",
			Desc: "Counter for rows with Single-Bit ECC errors",
			Type: 0,
		},
		SampleMetricMetadata{
			Uid:  pendingRowsState,
			Name: "habanalabs_pending_rows_state",
			Desc: "Whether or not there are rows waiting for replacement",
			Type: 0,
		},
	)
	return meta
}

func SamplesDesc() []*prometheus.Desc {
	desc := make([]*prometheus.Desc, numSamples)
	for i, sample := range SampleMetadata() {
		desc[i] = prometheus.NewDesc(
			sample.Name,
			sample.Desc,
			getMetricLabels(),
			nil,
		)
	}
	return desc
}

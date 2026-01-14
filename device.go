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
	"github.com/prometheus/client_golang/prometheus"
)

type DeviceInfo struct {
	UUID               string // device UUID
	KernelVersion      string // device kernel firmware version
	UbootVersion       string // device FIT firmware version
	BusID              string // busID of the device
	DriverVersion      string // driver version
	SerialNumber       string // serial number
	Pcb                string // PCB version
	PcbAssembly        string // PCB assembly version
	PcieLinkGeneration uint32 // PCIe link generation
}

func DeviceConfigDesc() *prometheus.Desc {
	return prometheus.NewDesc(
		"habanalabs_device_config",
		"Gaudi device configuration",
		[]string{"device", "id", "fit", "spi", "driver"},
		nil, // no const labels
	)
}

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
	"fmt"
	"os"
	"strings"
)

const (

	// HLDriverPath indicates on habana device dir
	HLDriverPath = "/sys/class/accel"
	// HLModulePath indicates on habana module dir
	HLModulePath = "/sys/module/habanalabs"
)

// FWVersion returns the firmware version for a given device
func FWVersion(idx uint) (kernel string, uboot string, err error) {
	b, err := os.ReadFile(fmt.Sprintf("%s/accel%d/device/armcp_kernel_ver", HLDriverPath, idx))
	if err != nil {
		return "", "", fmt.Errorf("file reading error %s", err)
	}
	kernel = strings.TrimSpace(string(b))

	b, err = os.ReadFile(fmt.Sprintf("%s/accel%d/device/uboot_ver", HLDriverPath, idx))
	if err != nil {
		return "", "", fmt.Errorf("file reading error %s", err)
	}
	uboot = strings.TrimSpace(string(b))

	return kernel, uboot, nil
}

// SystemDriverVersion returns the driver version on the system
func SystemDriverVersion() (string, error) {
	driver, err := os.ReadFile(HLModulePath + "/version")
	if err != nil {
		return "", fmt.Errorf("file reading error %s", err)
	}
	return strings.TrimSpace(string(driver)), nil
}

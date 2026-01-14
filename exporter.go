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
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"time"

	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus"
)

type Exporter struct {
	log *slog.Logger
	err chan error
}

// Returns list of labels
func getMetricLabels(labels ...string) []string {
	common := []string{"Serial", "UUID", "device", "id", "fit", "spi", "driver", "pod_namespace", "pod_name"}
	common = append(common, labels...)
	return common
}

func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	ch <- DeviceConfigDesc()
	ch <- KubeConfigDesc()
	for _, samples := range SamplesDesc() {
		ch <- samples
	}
}

// kube.go
func kubeMetrics() (prometheus.Metric, error) {
	info := GetKubeInfo()
	return prometheus.NewConstMetric(
		KubeConfigDesc(),
		prometheus.GaugeValue,
		1,
		info.podName,
		info.namespace,
		info.podIP,
		info.node,
		info.hostname,
	)
}

func LocalExecContext(ctx context.Context, command string, out ...io.Writer) error {
	cmd := exec.CommandContext(ctx, "/bin/bash", "-c", command)
	cmd.WaitDelay = 10 * time.Second
	var errOut bytes.Buffer
	cmd.Stderr = &errOut

	if len(out) > 0 {
		cmd.Stdout = out[0]
	}

	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("local exec: %s %w", errOut.String(), err)
	}

	return nil
}

func deviceLabels(log *slog.Logger, device DeviceData, namespace, pod string) []string {
	driverVersion, err := SystemDriverVersion()
	if err != nil {
		log.With("error", err).Error("failed to get driver version")
	}

	kernel, uboot, err := FWVersion(uint(device.DeviceID))
	if err != nil {
		log.With("error", err).Error("failed to get fw version")
	}

	res := []string{
		device.Serial,
		device.Uuid,
		device.Serial,
		fmt.Sprintf("%d", device.DeviceID),
		uboot,
		kernel,
		driverVersion,
		namespace,
		pod,
	}

	return res
}

// promMetric returns a prometheus metric
func promMetric(log *slog.Logger, name string, value float64, device DeviceData, namespace string, pod string, kv ...map[string]string) prometheus.Metric {
	var desc string

	var keys, values []string
	if len(kv) > 0 {
		for k, v := range kv[0] {
			keys = append(keys, k)
			values = append(values, v)
		}
	}

	devLabels := deviceLabels(log, device, namespace, pod)
	devLabels = append(devLabels, values...)

	return prometheus.MustNewConstMetric(
		prometheus.NewDesc(
			name,
			desc,
			getMetricLabels(keys...), nil),
		prometheus.GaugeValue,
		value,
		devLabels...,
	)
}

type (
	DevicesData []DeviceData
	DeviceData  struct {
		PciId       string `json:"pci_id,omitempty"`
		DeviceID    int    `json:"device_id,omitempty"`
		Serial      string `json:"serial,omitempty"`
		ProductName string `json:"product_name,omitempty"`
		MinorNumber uint64 `json:"minor_number,omitempty"`

		// Metrics values
		PowerDefaultLimit             uint64       `json:"power_default_limit,omitempty"`
		PowerUsage                    uint64       `json:"power_usage,omitempty"`
		TemperatureOnBoard            uint64       `json:"temperature_on_board,omitempty"`
		TemperatureOnChip             uint64       `json:"temperature_on_chip,omitempty"`
		TemperatureThresholdGpu       uint64       `json:"temperature_threshold_gpu,omitempty"`
		TemperatureThresholdMemory    uint64       `json:"temperature_threshold_memory,omitempty"`
		TemperatureThresholdShutdown  uint64       `json:"temperature_threshold_shutdown,omitempty"`
		TemperatureThresholdSlowdown  uint64       `json:"temperature_threshold_slowdown,omitempty"`
		PendingRowsSingleBitEccErrors uint64       `json:"pending_rows_single_bit_ecc_errors,omitempty"`
		PendingRowsDoubleBitEccErrors uint64       `json:"pending_rows_double_bit_ecc_errors,omitempty"`
		PendingRowsState              uint64       `json:"pending_rows_state,omitempty"`
		PcieRx                        uint64       `json:"pcie_rx,omitempty"`
		PcieTx                        uint64       `json:"pcie_tx,omitempty"`
		PcieReplayCount               uint64       `json:"pcie_replay_count,omitempty"`
		PciLinkWidth                  uint64       `json:"pci_link_width,omitempty"`
		PciLinkSpeed                  uint64       `json:"pci_link_speed,omitempty"`
		MemoryTotalBytes              uint64       `json:"memory_total_bytes,omitempty"`
		MemoryUsedBytes               uint64       `json:"memory_used_bytes,omitempty"`
		MemoryFreeBytes               uint64       `json:"memory_free_bytes,omitempty"`
		EccCurrentFeatureMode         uint64       `json:"ecc_current_feature_mode,omitempty"`
		Energy                        uint64       `json:"energy,omitempty"`
		ClockCurrent                  uint64       `json:"clock_current,omitempty"`
		ClockMax                      uint64       `json:"clock_max,omitempty"`
		Utilization                   uint64       `json:"utilization,omitempty"`
		Uuid                          string       `json:"uuid,omitempty"`
		Ports                         []PortStatus `json:"ports"`
	}
)

type PortStatus struct {
	PortNumber int    `json:"port"`
	PortType   string `json:"port_type"`
	Status     int    `json:"status"`
}

func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
	uuid := uuid.NewString()
	ll := e.log.With("uuid", uuid)

	ll.Info("Starting collect()")
	defer ll.Info("Finished collect()")

	buf := bytes.Buffer{}
	err := LocalExecContext(context.Background(), "rhlml", &buf)
	if err != nil {
		ll.With("error", err).Error("failed to execute rhlml")
		return
	}

	var devices DevicesData

	err = json.Unmarshal(buf.Bytes(), &devices)
	if err != nil {
		ll.With("error", err).Error("failed to read device data")
		return
	}

	kubeDeviceMetadata := make(map[string]KubeInfo)
	// Try collecting metrics from kubelet only when running inside kubernetes environment
	if _, ok := os.LookupEnv("KUBERNETES_SERVICE_PORT"); ok {

		// habanalabs_kube_info
		kubeData, err := kubeMetrics()
		if err != nil {
			_, err := fmt.Fprintf(os.Stderr, "There was an error getting kube metrics: %s\n", err.Error())
			if err != nil {
				ll.With("error", err).Error("Failed to write error to stderr")
			}
		} else {
			ch <- kubeData
		}

		kubeDeviceMetadata, err = GetKubeDeviceMetadata(e.log)
		if err != nil {
			ll.With("error", err).Error("Could not find Kube assignment metadata")
			return
		}
	}

	for _, dev := range devices {

		kubeInfo := GetKubeMetadata(dev.Serial, kubeDeviceMetadata)

		namespace := kubeInfo.namespace
		pod := kubeInfo.podName

		// Expected metrics

		// TODO: According to python code: soc - gaudi, ic - goya, mme - goya, tpc - goya.
		// habanalabs_clock_ic_max_mhz
		// habanalabs_clock_mme_max_mhz
		// habanalabs_clock_tpc_max_mhz

		// habanalabs_clock_soc_max_mhz

		ch <- promMetric(e.log, "habanalabs_clock_soc_max_mhz", float64(dev.ClockMax), dev, namespace, pod)

		// accelerator_memory_clock_hertz
		ch <- promMetric(e.log, "accelerator_memory_clock_hertz", float64(18000000), dev, namespace, pod)

		// habanalabs_clock_soc_mhz and accelerator_sm_clock_hertz
		ch <- promMetric(e.log, "habanalabs_clock_soc_mhz", float64(dev.ClockCurrent), dev, namespace, pod)
		ch <- promMetric(e.log, "accelerator_sm_clock_hertz", float64(dev.ClockCurrent), dev, namespace, pod)

		// habanalabs_device_config
		ch <- promMetric(e.log, "habanalabs_device_config", float64(1), dev, namespace, pod)

		// habanalabs_ecc_feature_mode
		ch <- promMetric(e.log, "habanalabs_ecc_feature_mode", float64(dev.EccCurrentFeatureMode), dev, namespace, pod)

		// habanalabs_energy
		ch <- promMetric(e.log, "habanalabs_energy", float64(dev.Energy), dev, namespace, pod)

		// habanalabs_memory_free_bytes
		ch <- promMetric(e.log, "habanalabs_memory_free_bytes", float64(dev.MemoryFreeBytes), dev, namespace, pod)

		// habanalabs_memory_total_bytes and accelerator_memory_total_bytes
		ch <- promMetric(e.log, "habanalabs_memory_total_bytes", float64(dev.MemoryTotalBytes), dev, namespace, pod)
		ch <- promMetric(e.log, "accelerator_memory_total_bytes", float64(dev.MemoryTotalBytes), dev, namespace, pod)

		// habanalabs_memory_used_bytes and accelerator_memory_used_bytes
		ch <- promMetric(e.log, "habanalabs_memory_used_bytes", float64(dev.MemoryUsedBytes), dev, namespace, pod)
		ch <- promMetric(e.log, "accelerator_memory_used_bytes", float64(dev.MemoryUsedBytes), dev, namespace, pod)

		// habanalabs_pci_link_speed
		ch <- promMetric(e.log, "habanalabs_pci_link_speed", float64(dev.PciLinkSpeed), dev, namespace, pod)

		// habanalabs_pci_link_width
		ch <- promMetric(e.log, "habanalabs_pci_link_width", float64(dev.PciLinkWidth), dev, namespace, pod)

		// habanalabs_pcie_receive_throughput
		ch <- promMetric(e.log, "habanalabs_pcie_receive_throughput", float64(dev.PcieRx), dev, namespace, pod)

		// habanalabs_pcie_replay_count
		ch <- promMetric(e.log, "habanalabs_pcie_replay_count", float64(dev.PcieReplayCount), dev, namespace, pod)

		// habanalabs_pcie_rx
		ch <- promMetric(e.log, "habanalabs_pcie_rx", float64(dev.PcieRx), dev, namespace, pod)

		// habanalabs_pcie_transmit_throughput
		ch <- promMetric(e.log, "habanalabs_pcie_transmit_throughput", float64(dev.PcieTx), dev, namespace, pod)

		// habanalabs_pcie_tx
		ch <- promMetric(e.log, "habanalabs_pcie_tx", float64(dev.PcieTx), dev, namespace, pod)

		// habanalabs_pending_rows_state
		ch <- promMetric(e.log, "habanalabs_pending_rows_state", float64(dev.PendingRowsState), dev, namespace, pod)

		// habanalabs_pending_rows_with_double_bit_ecc_errors
		ch <- promMetric(e.log, "habanalabs_pending_rows_with_double_bit_ecc_errors", float64(dev.PendingRowsDoubleBitEccErrors), dev, namespace, pod)

		// habanalabs_pending_rows_with_single_bit_ecc_errors
		ch <- promMetric(e.log, "habanalabs_pending_rows_with_single_bit_ecc_errors", float64(dev.PendingRowsSingleBitEccErrors), dev, namespace, pod)

		// habanalabs_power_default_limit_mW
		ch <- promMetric(e.log, "habanalabs_power_default_limit_mW", float64(dev.PowerDefaultLimit), dev, namespace, pod)

		// habanalabs_power_mW and accelerator_power_usage_watts
		ch <- promMetric(e.log, "habanalabs_power_mW", float64(dev.PowerUsage), dev, namespace, pod)
		ch <- promMetric(e.log, "accelerator_power_usage_watts", float64(dev.PowerUsage), dev, namespace, pod)

		// habanalabs_temperature_onboard
		ch <- promMetric(e.log, "habanalabs_temperature_onboard", float64(dev.TemperatureOnBoard), dev, namespace, pod)

		// habanalabs_temperature_onchip and accelerator_temperature_celcius
		ch <- promMetric(e.log, "habanalabs_temperature_onchip", float64(dev.TemperatureOnChip), dev, namespace, pod)
		ch <- promMetric(e.log, "accelerator_temperature_celcius", float64(dev.TemperatureOnChip), dev, namespace, pod)

		// habanalabs_temperature_threshold_gpu
		ch <- promMetric(e.log, "habanalabs_temperature_threshold_gpu", float64(dev.TemperatureThresholdGpu), dev, namespace, pod)

		// habanalabs_temperature_threshold_memory
		ch <- promMetric(e.log, "habanalabs_temperature_threshold_memory", float64(dev.TemperatureThresholdMemory), dev, namespace, pod)

		// habanalabs_temperature_threshold_shutdown
		ch <- promMetric(e.log, "habanalabs_temperature_threshold_shutdown", float64(dev.TemperatureThresholdShutdown), dev, namespace, pod)

		// habanalabs_temperature_threshold_slowdown
		ch <- promMetric(e.log, "habanalabs_temperature_threshold_slowdown", float64(dev.TemperatureThresholdSlowdown), dev, namespace, pod)

		// habanalabs_utilization & accelerator_gpu_utilization
		ch <- promMetric(e.log, "habanalabs_utilization", float64(dev.Utilization), dev, namespace, pod)
		ch <- promMetric(e.log, "accelerator_gpu_utilization", float64(dev.Utilization), dev, namespace, pod)

		for _, port := range dev.Ports {
			ch <- promMetric(e.log, "habanalabs_nic_port_status", float64(port.Status), dev, namespace, pod, map[string]string{
				"port":      fmt.Sprintf("%d", port.PortNumber),
				"port_type": port.PortType,
			})
		}
	}
}

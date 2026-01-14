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
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"

	podresourcesapi "k8s.io/kubelet/pkg/apis/podresources/v1"
)

const (
	socketDir         = "/var/lib/kubelet/pod-resources"
	socketPath        = socketDir + "/kubelet.sock"
	connectionTimeout = 10 * time.Second
)

type KubeInfo struct {
	podName   string
	namespace string
	podIP     string
	node      string
	hostname  string
}

func connectToServer(socket string) (*grpc.ClientConn, error) {
	ctx, cancel := context.WithTimeout(context.Background(), connectionTimeout)
	defer cancel()

	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	conn, err := grpc.DialContext(ctx, fmt.Sprintf("unix://%s", socket), opts...)
	if err != nil {
		return nil, fmt.Errorf("failure connecting to %s: %v", socket, err)
	}
	return conn, nil
}

func KubeConfigDesc() *prometheus.Desc {
	return prometheus.NewDesc(
		"habanalabs_kube_info",
		"kubernetes info",
		[]string{"podName", "namespace", "podIP", "node", "hostname"},
		nil,
	)
}

// getKubeMetadata checks the map for serial number of device and Kubernetes metadata.
// if it can't find the Kubelet information, it defaults the labels to empty strings.
func GetKubeMetadata(serialNumber string, kubeDeviceMetadata map[string]KubeInfo) KubeInfo {
	if d, ok := kubeDeviceMetadata[serialNumber]; ok {
		return d
	}

	return KubeInfo{
		podName:   "",
		namespace: "",
	}
}

// GetDeviceMetadata returns a map from the device ID to the Kubernetes metadata
func GetKubeDeviceMetadata(log *slog.Logger) (map[string]KubeInfo, error) {
	log.Info("Connecting to pod resources...")

	conn, err := connectToServer(socketPath)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	client := podresourcesapi.NewPodResourcesListerClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), connectionTimeout)
	defer cancel()

	response, err := client.List(ctx, &podresourcesapi.ListPodResourcesRequest{})
	if err != nil {
		status, ok := status.FromError(err)
		if ok {
			return nil, fmt.Errorf("error listing pod-resources api [%s]: %s", status.Code(), status.Message())
		} else {
			return nil, fmt.Errorf("listing pod resources: %w", err)
		}
	}

	devices := make(map[string]KubeInfo)
	for _, pod := range response.GetPodResources() {
		for _, container := range pod.GetContainers() {
			for _, podDevices := range container.GetDevices() {
				info := KubeInfo{
					podName:   pod.GetName(),
					namespace: pod.GetNamespace(),
				}

				for _, deviceId := range podDevices.GetDeviceIds() {
					log.Info(
						"Added device metadata",
						"device_id", deviceId,
						"pod", pod.GetName(),
						"namespace", pod.GetNamespace(),
					)
					devices[deviceId] = info
				}
			}
		}
	}
	log.Info("Finished collecting pod resources")
	return devices, nil
}

// GetKubeInfo consumes envirovnment variables set by the kube
// downstream API. See manifest/metricExporter.yml in the HLO
func GetKubeInfo() KubeInfo {
	info := KubeInfo{
		podName:   os.Getenv("KUBE_POD_NAME"),
		namespace: os.Getenv("KUBE_NAMESPACE"),
		podIP:     os.Getenv("KUBE_POD_IP"),
		node:      os.Getenv("KUBE_NODE"),
		hostname:  os.Getenv("HOSTNAME"),
	}

	return info
}

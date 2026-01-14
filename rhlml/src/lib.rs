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

use hlml::{
    device::EccMode,
    device::{MemoryInfo, PortStatus as HlmlPorts},
    device_handle_by_index, init,
};
use log::{debug, error};
use serde::Serialize;

mod error;
use error::Error;

#[derive(Debug, Serialize)]
pub struct Metrics {
    // General static info
    device_id: u32,
    pci_id: String,
    serial: String,
    product_name: String,
    minor_number: u32,
    uuid: String,

    // Metrics values
    power_default_limit: u32,
    power_usage: u32,
    temperature_on_board: u32,
    temperature_on_chip: u32,
    temperature_threshold_gpu: u32,
    temperature_threshold_memory: u32,
    temperature_threshold_shutdown: u32,
    temperature_threshold_slowdown: u32,
    pending_rows_single_bit_ecc_errors: u32,
    pending_rows_double_bit_ecc_errors: u32,
    pending_rows_state: u32,
    pcie_rx: u32,
    pcie_tx: u32,
    pcie_replay_count: u32,
    pci_link_width: u32,
    pci_link_speed: u32,
    memory_total_bytes: u64,
    memory_used_bytes: u64,
    memory_free_bytes: u64,
    ecc_current_feature_mode: u32,
    energy: u64,
    clock_current: u32,
    clock_max: u32,
    utilization: u32,
    ports: Vec<PortStatus>,
}

#[derive(Debug, Serialize)]
pub struct PortStatus {
    pub port: u32,
    pub port_type: &'static str,
    pub status: u32,
}

impl From<HlmlPorts> for PortStatus {
    fn from(value: HlmlPorts) -> Self {
        Self {
            port: value.port,
            port_type: value.port_type.get_name(),
            status: value.status,
        }
    }
}

#[derive(Clone, Debug)]
pub struct Rhlml;

impl Rhlml {
    // Construcring the object making sure underlaying initialization done once.
    pub fn new() -> Result<Self, Error> {
        if init().is_err() {
            return Err(Error::FailedInit);
        }
        Ok(Rhlml)
    }

    pub fn device_count(&self) -> Result<u32, Error> {
        let count = match hlml::device_count() {
            Ok(c) => c,
            Err(e) => {
                error!("failed getting device count: {e:?}");
                std::process::exit(1);
            }
        };
        Ok(count)
    }

    pub fn collect_all(&self) -> Result<Vec<Metrics>, Error> {
        let count = self.device_count()?;
        let mut all_devices: Vec<Metrics> = vec![];
        for i in 0..count {
            all_devices.push(self.collect(i));
        }
        Ok(all_devices)
    }

    pub fn collect(&self, index: u32) -> Metrics {
        debug!("--> getting index {index}");
        let device = match device_handle_by_index(index) {
            Ok(d) => d,
            Err(e) => {
                error!("failed getting device by handle:{e:?}");
                std::process::exit(1);
            }
        };
        let pci_id = match device.pci_bus_id() {
            Ok(id) => id,
            Err(e) => {
                error!("failed getting pci id: {e:?}");
                "".to_string();
                std::process::exit(1);
            }
        };
        let serial = device.serial_number().unwrap_or("".to_string());
        let uuid = device.uuid().unwrap_or("na".to_string());
        let product_name = device.product_name().unwrap_or("na".to_string());
        let minor_number = device.minor_number().unwrap_or(0);
        let power_default_limit = device.power_management_default_limit().unwrap_or(0);
        let power_usage = device.power_usage().unwrap_or(0);
        let temperature_on_board = device.temperature_on_board().unwrap_or(0);
        let temperature_on_chip = device.temperature_on_chip().unwrap_or(0);
        let temperature_threshold_slowdown = device.temperature_threshold_slowdown().unwrap_or(0);
        let temperature_threshold_shutdown = device.temperature_threshold_shutdown().unwrap_or(0);
        let temperature_threshold_memory = device.temperature_threshold_memory().unwrap_or(0);
        let temperature_threshold_gpu = device.temperature_threshold_gpu().unwrap_or(0);
        let pending_rows_single_bit_ecc_errors = device.replaced_row_single_bit_ecc().unwrap_or(0);
        let pending_rows_double_bit_ecc_errors = device.replaced_row_double_bit_ecc().unwrap_or(0);
        let pending_rows_state = device.replaced_rows_pending_status().unwrap_or(0);
        let pcie_rx = device.pcie_rx().unwrap_or(0);
        let pcie_tx = device.pcie_tx().unwrap_or(0);
        let pcie_replay_count = device.pci_replay_counter().unwrap_or(0);
        let pci_link_width = device.pcie_link_width().unwrap_or(0);
        let pci_link_speed = device.link_speed().unwrap_or(0);
        let mem_info = device.memory_info().unwrap_or(MemoryInfo {
            total: 0,
            used: 0,
            free: 0,
        });
        let memory_total_bytes = mem_info.total;
        let memory_used_bytes = mem_info.used;
        let memory_free_bytes = mem_info.free;
        let ecc_current_feature_mode = device
            .ecc_mode()
            .unwrap_or(EccMode {
                current: 0,
                pending: 0,
            })
            .current;
        let energy = device.energy_consumption_counter().unwrap_or(0);
        let clock_current = device.clock_info().unwrap_or(0);
        let clock_max = device.clock_max().unwrap_or(0);
        let utilization = device.utilization_info().unwrap_or(0);

        let hlml_ports = device.ports_status().expect("masks");
        let ports = hlml_ports
            .iter()
            .map(|p| PortStatus {
                port: p.port,
                port_type: p.port_type.get_name(),
                status: p.status,
            })
            .collect::<Vec<PortStatus>>();

        Metrics {
            device_id: index,
            pci_id,
            serial,
            product_name,
            minor_number,
            uuid,
            power_default_limit,
            power_usage,
            temperature_on_board,
            temperature_on_chip,
            temperature_threshold_gpu,
            temperature_threshold_memory,
            temperature_threshold_shutdown,
            temperature_threshold_slowdown,
            pending_rows_single_bit_ecc_errors,
            pending_rows_double_bit_ecc_errors,
            pending_rows_state,
            pcie_rx,
            pcie_tx,
            pcie_replay_count,
            pci_link_width,
            pci_link_speed,
            memory_total_bytes,
            memory_used_bytes,
            memory_free_bytes,
            ecc_current_feature_mode,
            energy,
            clock_current,
            clock_max,
            utilization,
            ports,
        }
    }
}

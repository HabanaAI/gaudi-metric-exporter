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

#![allow(non_upper_case_globals)]
#![allow(non_camel_case_types)]
#![allow(non_snake_case)]

use log::debug;
use std::ffi::CStr;

const VERSION_MAX_LEN: u32 = 128;

use crate::{result_from_bindings, HlmlResult};
use hlml_sys::bindings::*;

pub struct Device {
    dev: hlml_device_t,
}

impl From<hlml_device_t> for Device {
    fn from(dev: hlml_device_t) -> Self {
        Device { dev }
    }
}

#[derive(Debug)]
pub struct PciInfo {
    pub bus: uint,
    pub bus_id: String,
    pub device: uint,
    pub domain: uint,
    pub pci_device_id: String,
    pub caps: PciCap,
}

#[derive(Debug)]
pub struct PciCap {
    pub link_speed: String,
    pub link_width: String,
}

#[derive(Debug)]
pub struct MemoryInfo {
    pub total: u64,
    pub used: u64,
    pub free: u64,
}

#[derive(Debug)]
pub struct EccMode {
    pub current: u32,
    pub pending: u32,
}

#[derive(Debug)]
pub enum PortType {
    External,
    Internal,
}

impl PortType {
    pub fn get_name(&self) -> &'static str {
        match self {
            PortType::External => "external",
            PortType::Internal => "internal",
        }
    }
}

#[derive(Debug)]
pub struct PortStatus {
    pub port: u32,
    pub port_type: PortType,
    pub status: u32,
}

impl Device {
    pub fn minor_number(&self) -> HlmlResult<u32> {
        debug!("--> minor number");
        let mut minor: u32 = 0;
        unsafe {
            result_from_bindings(hlml_device_get_minor_number(self.dev, &mut minor))?;
        }
        Ok(minor)
    }

    pub fn product_name(&self) -> HlmlResult<String> {
        debug!("--> product name");
        let mut name_buffer: Vec<u8> = vec![0; 8_usize];
        unsafe {
            result_from_bindings(hlml_device_get_name(
                self.dev,
                name_buffer.as_mut_ptr() as *mut i8,
                8,
            ))?;
        }
        let c_str = CStr::from_bytes_until_nul(&name_buffer).expect("bad name value");
        Ok(c_str.to_str().unwrap_or("").to_string())
    }

    pub fn uuid(&self) -> HlmlResult<String> {
        debug!("--> uuid");
        let mut uuid_buffer: Vec<u8> = vec![0; VERSION_MAX_LEN as usize];
        unsafe {
            result_from_bindings(hlml_device_get_uuid(
                self.dev,
                uuid_buffer.as_mut_ptr() as *mut i8,
                VERSION_MAX_LEN,
            ))?;
        }
        let c_str = CStr::from_bytes_until_nul(&uuid_buffer).expect("bad name value");
        Ok(c_str.to_str().unwrap_or("").to_string())
    }

    pub fn serial_number(&self) -> HlmlResult<String> {
        debug!("--> serial");
        let mut serial_buffer: Vec<u8> = vec![0; HL_FIELD_MAX_SIZE as usize];
        unsafe {
            result_from_bindings(hlml_device_get_serial(
                self.dev,
                serial_buffer.as_mut_ptr() as *mut i8,
                HL_FIELD_MAX_SIZE,
            ))?;
        }
        let c_str = CStr::from_bytes_until_nul(&serial_buffer).expect("bad serial value");
        Ok(c_str.to_str().unwrap_or("").to_string())
    }

    pub fn pci_domain(&self) -> HlmlResult<String> {
        debug!("--> pci domain");
        unsafe {
            let mut c_pci_info: hlml_pci_info = std::mem::zeroed();
            result_from_bindings(hlml_device_get_pci_info(self.dev, &mut c_pci_info))?;
            let domain = format!("{:#x}", c_pci_info.pci_device_id);
            Ok(domain)
        }
    }

    pub fn pci_bus(&self) -> HlmlResult<String> {
        debug!("--> pci bus");
        unsafe {
            let mut c_pci_info: hlml_pci_info = std::mem::zeroed();
            result_from_bindings(hlml_device_get_pci_info(self.dev, &mut c_pci_info))?;
            let bus = format!("{:#x}", c_pci_info.bus);
            Ok(bus)
        }
    }

    pub fn pci_bus_id(&self) -> HlmlResult<String> {
        debug!("--> pci bus id");
        unsafe {
            let mut c_pci_info: hlml_pci_info = std::mem::zeroed();
            result_from_bindings(hlml_device_get_pci_info(self.dev, &mut c_pci_info))?;
            let c_str = CStr::from_ptr(c_pci_info.bus_id.as_ptr());
            let bus_id = c_str.to_str().unwrap_or("").to_owned();
            Ok(bus_id)
        }
    }

    pub fn pci_id(&self) -> HlmlResult<String> {
        debug!("--> pci id");
        unsafe {
            let mut c_pci_info: hlml_pci_info = std::mem::zeroed();
            result_from_bindings(hlml_device_get_pci_info(self.dev, &mut c_pci_info))?;
            let pci_device_id = format!("{:#x}", c_pci_info.pci_device_id);
            Ok(pci_device_id)
        }
    }

    /// Returns the current PCI link speed for a given device
    pub fn link_speed(&self) -> HlmlResult<u32> {
        debug!("--> link speed");
        let speed = unsafe {
            let mut c_pci_info: hlml_pci_info = std::mem::zeroed();
            result_from_bindings(hlml_device_get_pci_info(self.dev, &mut c_pci_info))?;
            let c_str = CStr::from_ptr(c_pci_info.caps.link_speed.as_ptr());
            let link_speed = c_str.to_str().unwrap_or("").to_owned();
            link_speed
                .chars()
                .take_while(|c| c.is_numeric())
                .collect::<String>()
        };
        speed.parse::<u32>().map_err(|_| crate::HlmlError::NoData)
    }

    /// Returns the current PCI link width for a given device
    pub fn link_width(&self) -> HlmlResult<u32> {
        let width = unsafe {
            let mut c_pci_info: hlml_pci_info = std::mem::zeroed();
            result_from_bindings(hlml_device_get_pci_info(self.dev, &mut c_pci_info))?;
            let c_str = CStr::from_ptr(c_pci_info.caps.link_width.as_ptr());
            let mut link_width = c_str.to_str().unwrap_or("").to_owned();
            // Remove starting 'x' char
            _ = link_width.remove(0);
            link_width
        };
        width.parse::<u32>().map_err(|_| crate::HlmlError::NoData)
    }

    pub fn memory_info(&self) -> HlmlResult<MemoryInfo> {
        unsafe {
            let mut c_memory_info: hlml_memory_t = std::mem::zeroed();
            result_from_bindings(hlml_device_get_memory_info(self.dev, &mut c_memory_info))?;
            let total = c_memory_info.total;
            let used = c_memory_info.used;
            let free = total - used;
            Ok(MemoryInfo { total, used, free })
        }
    }

    pub fn utilization_info(&self) -> HlmlResult<u32> {
        unsafe {
            let mut c_util_info: hlml_utilization_t = std::mem::zeroed();
            result_from_bindings(hlml_device_get_utilization_rates(
                self.dev,
                &mut c_util_info,
            ))?;
            Ok(c_util_info.aip)
        }
    }

    /// Returns the Gaudi clock frequency
    pub fn clock_info(&self) -> HlmlResult<u32> {
        let mut freq: u32 = 0;
        unsafe {
            result_from_bindings(hlml_device_get_clock_info(
                self.dev,
                hlml_clock_type_HLML_CLOCK_SOC,
                &mut freq,
            ))?;
        }
        Ok(freq)
    }

    /// Returns the Gaudi maximum clock frequency
    pub fn clock_max(&self) -> HlmlResult<u32> {
        let mut freq: u32 = 0;
        unsafe {
            result_from_bindings(hlml_device_get_max_clock_info(
                self.dev,
                hlml_clock_type_HLML_CLOCK_SOC,
                &mut freq,
            ))?;
        }
        Ok(freq)
    }

    /// Returns the power usage in milliwatts for a given device
    pub fn power_usage(&self) -> HlmlResult<u32> {
        let mut power: u32 = 0;
        unsafe {
            result_from_bindings(hlml_device_get_power_usage(self.dev, &mut power))?;
        }
        Ok(power)
    }

    /// Returns the temperature in celsius for a device board
    pub fn temperature_on_board(&self) -> HlmlResult<u32> {
        let mut temp: u32 = 0;
        unsafe {
            result_from_bindings(hlml_device_get_temperature(
                self.dev,
                hlml_temperature_sensors_HLML_TEMPERATURE_ON_BOARD,
                &mut temp,
            ))?;
        }
        Ok(temp)
    }

    /// Returns the temperature in celsius for a the device chip
    pub fn temperature_on_chip(&self) -> HlmlResult<u32> {
        let mut temp: u32 = 0;
        unsafe {
            result_from_bindings(hlml_device_get_temperature(
                self.dev,
                hlml_temperature_sensors_HLML_TEMPERATURE_ON_AIP,
                &mut temp,
            ))?;
        }
        Ok(temp)
    }

    /// Retrieves the known temperature threshold for the AIP with the specified threshold type in degrees
    pub fn temperature_threshold_shutdown(&self) -> HlmlResult<u32> {
        let mut temp: u32 = 0;
        unsafe {
            result_from_bindings(hlml_device_get_temperature_threshold(
                self.dev,
                hlml_temperature_thresholds_HLML_TEMPERATURE_THRESHOLD_SHUTDOWN,
                &mut temp,
            ))?;
        }
        Ok(temp)
    }

    pub fn temperature_threshold_slowdown(&self) -> HlmlResult<u32> {
        let mut temp: u32 = 0;
        unsafe {
            result_from_bindings(hlml_device_get_temperature_threshold(
                self.dev,
                hlml_temperature_thresholds_HLML_TEMPERATURE_THRESHOLD_SLOWDOWN,
                &mut temp,
            ))?;
        }
        Ok(temp)
    }

    pub fn temperature_threshold_memory(&self) -> HlmlResult<u32> {
        let mut temp: u32 = 0;
        unsafe {
            result_from_bindings(hlml_device_get_temperature_threshold(
                self.dev,
                hlml_temperature_thresholds_HLML_TEMPERATURE_THRESHOLD_MEM_MAX,
                &mut temp,
            ))?;
        }
        Ok(temp)
    }

    pub fn temperature_threshold_gpu(&self) -> HlmlResult<u32> {
        let mut temp: u32 = 0;
        unsafe {
            result_from_bindings(hlml_device_get_temperature_threshold(
                self.dev,
                hlml_temperature_thresholds_HLML_TEMPERATURE_THRESHOLD_GPU_MAX,
                &mut temp,
            ))?;
        }
        Ok(temp)
    }

    // PowerManagementDefaultLimit Retrieves default power management limit on this device, in milliwatts.
    // Default power management limit is a power management limit that the device boots with.
    pub fn power_management_default_limit(&self) -> HlmlResult<u32> {
        let mut limit: u32 = 0;
        unsafe {
            result_from_bindings(hlml_device_get_power_management_default_limit(
                self.dev, &mut limit,
            ))?;
        }
        Ok(limit)
    }

    // ECCMode retrieves the current and pending ECC modes for the device
    //
    //	1 - ECCMode enabled
    //	0 - ECCMode disabled
    pub fn ecc_mode(&self) -> HlmlResult<EccMode> {
        let mut current: u32 = 0;
        let mut pending: u32 = 0;
        unsafe {
            result_from_bindings(hlml_device_get_ecc_mode(
                self.dev,
                &mut current,
                &mut pending,
            ))?;
        }
        Ok(EccMode { current, pending })
    }

    /// returns the revision of the HL library
    pub fn hl_revision(&self) -> HlmlResult<i32> {
        let mut revision: i32 = 0;
        unsafe {
            result_from_bindings(hlml_device_get_hl_revision(self.dev, &mut revision))?;
        }
        Ok(revision)
    }

    pub fn pcb_version(&self) -> HlmlResult<String> {
        unsafe {
            let mut c_pcb_info: hlml_pcb_info_t = std::mem::zeroed();
            result_from_bindings(hlml_device_get_pcb_info(self.dev, &mut c_pcb_info))?;
            let c_str = CStr::from_ptr(c_pcb_info.pcb_ver.as_ptr());
            let pcb_ver = c_str.to_str().unwrap_or("").to_owned();
            Ok(pcb_ver)
        }
    }

    pub fn pcb_assembly_version(&self) -> HlmlResult<String> {
        unsafe {
            let mut c_pcb_info: hlml_pcb_info_t = std::mem::zeroed();
            result_from_bindings(hlml_device_get_pcb_info(self.dev, &mut c_pcb_info))?;
            let c_str = CStr::from_ptr(c_pcb_info.pcb_assembly_ver.as_ptr());
            let pcb_ver = c_str.to_str().unwrap_or("").to_owned();
            Ok(pcb_ver)
        }
    }

    pub fn module_id(&self) -> HlmlResult<u32> {
        let mut mod_id: u32 = 0;
        unsafe {
            result_from_bindings(hlml_device_get_module_id(self.dev, &mut mod_id))?;
        }
        Ok(mod_id)
    }

    pub fn board_id(&self) -> HlmlResult<u32> {
        let mut board_id: u32 = 0;
        unsafe {
            result_from_bindings(hlml_device_get_board_id(self.dev, &mut board_id))?;
        }
        Ok(board_id)
    }

    /// PCIeTX returns PCIe transmit throughput
    pub fn pcie_tx(&self) -> HlmlResult<u32> {
        let mut val: u32 = 0;
        unsafe {
            result_from_bindings(hlml_device_get_pcie_throughput(
                self.dev,
                hlml_pcie_util_counter_HLML_PCIE_UTIL_TX_BYTES,
                &mut val,
            ))?;
        }
        Ok(val)
    }

    /// PCIeRX returns PCIe transmit throughput
    pub fn pcie_rx(&self) -> HlmlResult<u32> {
        let mut val: u32 = 0;
        unsafe {
            result_from_bindings(hlml_device_get_pcie_throughput(
                self.dev,
                hlml_pcie_util_counter_HLML_PCIE_UTIL_RX_BYTES,
                &mut val,
            ))?;
        }
        Ok(val)
    }

    /// returns PCIe replay count
    pub fn pci_replay_counter(&self) -> HlmlResult<u32> {
        let mut val: u32 = 0;
        unsafe {
            result_from_bindings(hlml_device_get_pcie_replay_counter(self.dev, &mut val))?;
        }
        Ok(val)
    }

    pub fn pcie_link_generation(&self) -> HlmlResult<u32> {
        let mut val: u32 = 0;
        unsafe {
            result_from_bindings(hlml_device_get_curr_pcie_link_generation(
                self.dev, &mut val,
            ))?;
        }
        Ok(val)
    }

    pub fn pcie_link_width(&self) -> HlmlResult<u32> {
        let mut val: u32 = 0;
        unsafe {
            result_from_bindings(hlml_device_get_curr_pcie_link_width(self.dev, &mut val))?;
        }
        Ok(val)
    }

    pub fn clock_throttle_reasons(&self) -> HlmlResult<u64> {
        let mut val: u64 = 0;
        unsafe {
            result_from_bindings(hlml_device_get_current_clocks_throttle_reasons(
                self.dev, &mut val,
            ))?;
        }
        Ok(val)
    }

    pub fn energy_consumption_counter(&self) -> HlmlResult<u64> {
        let mut val: u64 = 0;
        unsafe {
            result_from_bindings(hlml_device_get_total_energy_consumption(self.dev, &mut val))?;
        }
        Ok(val)
    }

    pub fn mac_address_info(&self) -> HlmlResult<u64> {
        todo!()
    }

    /// Returns the requested port status.
    /// 1 - up
    /// 0 - down
    pub fn nic_link_status(&self, port: u32) -> HlmlResult<u32> {
        let mut up: bool = false;
        unsafe {
            result_from_bindings(hlml_nic_get_link(self.dev, port, &mut up))?;
        }
        let status = match up {
            true => 1,
            false => 0,
        };
        Ok(status)
    }

    /// returns the number of rows with double-bit ecc errors
    pub fn replaced_row_double_bit_ecc(&self) -> HlmlResult<u32> {
        let mut val: u32 = 0;
        unsafe {
            let mut row: hlml_row_address = std::mem::zeroed();
            result_from_bindings(hlml_device_get_replaced_rows(
                self.dev,
                hlml_row_replacement_cause_HLML_ROW_REPLACEMENT_CAUSE_DOUBLE_BIT_ECC_ERROR,
                &mut val,
                &mut row,
            ))?;
        }

        Ok(val)
    }

    /// returns the number of rows with single-bit ecc errors
    pub fn replaced_row_single_bit_ecc(&self) -> HlmlResult<u32> {
        let mut val: u32 = 0;
        unsafe {
            let mut row: hlml_row_address = std::mem::zeroed();
            result_from_bindings(hlml_device_get_replaced_rows(
                self.dev,
                hlml_row_replacement_cause_HLML_ROW_REPLACEMENT_CAUSE_MULTIPLE_SINGLE_BIT_ECC_ERRORS,
                &mut val,
                &mut row,
            ))?;
        }

        Ok(val)
    }

    pub fn replaced_rows_pending_status(&self) -> HlmlResult<u32> {
        let mut pending: u32 = 0;
        unsafe {
            result_from_bindings(hlml_device_get_replaced_rows_pending_status(
                self.dev,
                &mut pending,
            ))?;
        }
        Ok(pending)
    }

    pub fn nic_get_statistics(&self) -> HlmlResult<u32> {
        unsafe {
            let mut stats: hlml_nic_stats_info_t = std::mem::zeroed();
            result_from_bindings(hlml_nic_get_statistics(self.dev, &mut stats))?;
            println!("{:?}", stats.port);
            println!("{:?}", stats.str_buf);
            Ok(0_u32)
        }
    }

    pub fn ports_status(&self) -> HlmlResult<Vec<PortStatus>> {
        let mut mask: u64 = 0;
        let mut ext_mask: u64 = 0;
        unsafe {
            result_from_bindings(hlml_get_mac_addr_info(self.dev, &mut mask, &mut ext_mask))?;
        }
        let mut ports = vec![];

        for port in 0..(std::mem::size_of_val(&mask) * 8) {
            if (mask & (1 << port)) == 0 {
                continue;
            }

            let is_ext = (ext_mask & (1 << port)) != 0;

            let link_status = self.nic_link_status(port as u32);
            let up = match link_status {
                Ok(1) => 1,
                Ok(0) => 0,
                _ => 0,
            };
            ports.push(PortStatus {
                port: port as u32,
                port_type: if is_ext {
                    PortType::External
                } else {
                    PortType::Internal
                },
                status: up as u32,
            });
        }

        Ok(ports)
    }
}

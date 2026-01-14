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

pub mod device;

use device::*;
use hlml_sys::bindings::*;
use thiserror::Error;

#[derive(Error, Debug)]
pub enum HlmlError {
    #[error("unknown error")]
    Unknown,
    #[error("hlml not initialized")]
    NotInitialized,
    #[error("invalid argument")]
    InvalidArgument,
    #[error("not supported")]
    NotSupported,
    #[error("already initialized")]
    AlreadyInitialized,
    #[error("not found")]
    NotFound,
    #[error("insufficient size")]
    InsufficientSize,
    #[error("driver not loaded")]
    DriverNotLoaded,
    #[error("aip is lost")]
    AipLost,
    #[error("memory error")]
    MemoryError,
    #[error("no data")]
    NoData,
}

pub type HlmlResult<T> = Result<T, HlmlError>;

fn result_from_bindings(i: hlml_return_t) -> HlmlResult<()> {
    match i {
        hlml_return_HLML_ERROR_UNKNOWN => Err(HlmlError::Unknown),
        hlml_return_HLML_ERROR_UNINITIALIZED => Err(HlmlError::NotInitialized),
        hlml_return_HLML_ERROR_INVALID_ARGUMENT => Err(HlmlError::InvalidArgument),
        hlml_return_HLML_ERROR_NOT_SUPPORTED => Err(HlmlError::NotSupported),
        hlml_return_HLML_ERROR_ALREADY_INITIALIZED => Err(HlmlError::AlreadyInitialized),
        hlml_return_HLML_ERROR_NOT_FOUND => Err(HlmlError::NotFound),
        hlml_return_HLML_ERROR_INSUFFICIENT_SIZE => Err(HlmlError::InsufficientSize),
        hlml_return_HLML_ERROR_DRIVER_NOT_LOADED => Err(HlmlError::DriverNotLoaded),
        hlml_return_HLML_ERROR_AIP_IS_LOST => Err(HlmlError::AipLost),
        hlml_return_HLML_ERROR_MEMORY => Err(HlmlError::MemoryError),
        hlml_return_HLML_ERROR_NO_DATA => Err(HlmlError::NoData),
        hlml_return_HLML_SUCCESS => Ok(()),
        _ => Err(HlmlError::Unknown),
    }
}

pub fn init() -> HlmlResult<()> {
    unsafe { result_from_bindings(hlml_init()) }
}

pub fn init_with_logs() -> HlmlResult<()> {
    unsafe { result_from_bindings(hlml_init_with_flags(0x6)) }
}

pub fn shutdown() -> HlmlResult<()> {
    unsafe { result_from_bindings(hlml_shutdown()) }
}

pub fn device_count() -> HlmlResult<u32> {
    let mut count: u32 = 0;
    unsafe {
        result_from_bindings(hlml_device_get_count(&mut count))?;
    }
    Ok(count)
}

pub fn device_handle_by_index(idx: u32) -> HlmlResult<Device> {
    unsafe {
        let mut c_device: hlml_device_t = std::mem::zeroed();
        result_from_bindings(hlml_device_get_handle_by_index(idx, &mut c_device))?;
        Ok(c_device.into())
    }
}

pub fn device_handle_by_uuid(uuid: &str) -> HlmlResult<Device> {
    unsafe {
        let mut c_device: hlml_device_t = std::mem::zeroed();
        result_from_bindings(hlml_device_get_handle_by_UUID(
            uuid.as_ptr() as *mut i8,
            &mut c_device,
        ))?;
        Ok(c_device.into())
    }
}

pub fn device_handle_by_serial(serial: &str) -> HlmlResult<Device> {
    let num_devices = device_count()?;
    for i in 0..num_devices {
        let device = device_handle_by_index(i)?;
        let dev_serial = device.serial_number()?;
        println!("got serial: {dev_serial:?}");
        println!("ask serial: {serial:?}");

        if dev_serial == serial {
            return Ok(device);
        }
    }
    Err(HlmlError::NotFound)
}

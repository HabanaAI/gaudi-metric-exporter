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

use log::error;
use rhlml::Rhlml;

fn main() {
    env_logger::init();

    let client = match Rhlml::new() {
        Ok(devs) => devs,
        Err(e) => {
            error!("Failed init: {e}");
            std::process::exit(1);
        }
    };

    let all_devices = match client.collect_all() {
        Ok(devs) => devs,
        Err(e) => {
            error!("Failed collecting devices metrics: {e}");
            std::process::exit(1);
        }
    };

    if serde_json::to_writer(std::io::stdout(), &all_devices).is_err() {
        error!("failed generating json output");
    }
}

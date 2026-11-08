# Marlinraker

Marlinraker lets you use [Mainsail](https://github.com/mainsail-crew/mainsail) and
[Fluidd](https://github.com/fluidd-core/fluidd) with 3D printers that run
[Marlin](https://marlinfw.org/) or Prusa firmware instead of Klipper.

It talks to the printer over USB serial and implements the
[Moonraker](https://github.com/Arksine/moonraker) API that these web interfaces expect,
emulating the Klipper objects (`toolhead`, `extruder`, `heater_bed`, `print_stats`, ...) on top
of regular G-code. It is a single static binary with no runtime dependencies, so it runs well on a
Raspberry Pi.

## Features

- Printer status, temperatures, position, speed and flow in Mainsail and Fluidd
- Printing from the host (files are streamed line by line), with pause, resume and cancel
- File manager with G-code metadata and thumbnails
- G-code console and Klipper-style G-code [macros](#macros)
- Print history and job totals
- Webcam configuration
- Host system information and service management via systemd
- OctoPrint upload API emulation, so slicers can upload and print directly

## Requirements

- Linux, e.g. Raspberry Pi OS. Release builds target `armv6`, `armv7` and `arm64`; other
  platforms can build from source.
- A printer running Marlin 2 or Prusa firmware, connected via USB
- Mainsail or Fluidd, served by a web server such as nginx

## Installation

1. Download the binary for your architecture from the
   [releases](https://github.com/marlinraker/marlinraker/releases) page, or build it from source:

   ```bash
   go install github.com/marlinraker/marlinraker/cmd/marlinraker@latest
   ```

2. Create a user that may access serial ports (do **not** run Marlinraker as root):

   ```bash
   sudo useradd --system --create-home --home-dir /home/marlinraker --groups dialout marlinraker
   ```

3. Put the binary in place and create the configuration:

   ```bash
   sudo install -m 755 marlinraker /usr/local/bin/marlinraker
   sudo -u marlinraker mkdir -p /home/marlinraker/marlinraker_files/config
   sudo -u marlinraker cp example_config/marlinraker.toml /home/marlinraker/marlinraker_files/config/
   sudo -u marlinraker cp example_config/printers/generic.toml /home/marlinraker/marlinraker_files/config/printer.toml
   ```

   Pick the printer file in [`example_config/printers`](example_config/printers) that is closest
   to your printer and adjust it.

4. Create `/etc/systemd/system/marlinraker.service`:

   ```ini
   [Unit]
   Description=Marlinraker
   After=network-online.target
   Wants=network-online.target

   [Service]
   User=marlinraker
   Environment=MARLINRAKER_DIR=/home/marlinraker/marlinraker_files
   ExecStart=/usr/local/bin/marlinraker
   Restart=always
   RestartSec=5

   [Install]
   WantedBy=multi-user.target
   ```

   Then start it:

   ```bash
   sudo systemctl daemon-reload
   sudo systemctl enable --now marlinraker
   ```

5. Serve Mainsail or Fluidd and point it at port 7125, e.g. with the nginx configuration that
   [Mainsail](https://docs.mainsail.xyz/setup/getting-started/manual-setup) or
   [Fluidd](https://docs.fluidd.xyz/installation/manual) documents for Moonraker. Marlinraker
   listens on the same port and paths as Moonraker.

Opening `http://<host>:7125/` shows a status page.

### Docker

```bash
docker build -t marlinraker .
docker run -d --name marlinraker \
  -p 7125:7125 \
  --device /dev/ttyUSB0 \
  -v marlinraker_files:/marlinraker/marlinraker_files \
  marlinraker
```

Put your configuration into `config/` inside the volume before starting the container.

## Data directory

Everything Marlinraker stores lives in the data directory, `./marlinraker_files` or the path in
the `MARLINRAKER_DIR` environment variable:

| Path              | Contents                                                          |
|-------------------|-------------------------------------------------------------------|
| `config/`         | Configuration, editable from Mainsail/Fluidd                      |
| `gcodes/`         | G-code files, thumbnails and metadata                             |
| `logs/`           | `marlinraker.log`, reopened on `SIGUSR1` for log rotation         |
| `db.json`         | Database (client settings, history, webcams)                      |

## Configuration

The main file is `config/marlinraker.toml`. It can include other files or whole directories with
`#include <path>` lines. Unknown options are reported as warnings in the log.

```toml
#include printer.toml

[web]
bind_address = ""   # listen on all interfaces
port = 7125
# Browser origins that may use the API, see below
cors_domains = ["*.lan", "*.local", "*://my.mainsail.xyz", "*://app.fluidd.xyz"]

[serial]
port = "auto"         # e.g. "/dev/ttyUSB0" or "auto" to scan
baud_rate = "auto"    # e.g. 115200 or "auto" to try common rates
max_connection_attempts = 5
connection_timeout = 5000   # ms

[misc]
octoprint_compat = true     # emulate the OctoPrint upload API
extended_logs = false       # debug logging, including all serial communication
allowed_services = ["marlinraker", "crowsnest", "KlipperScreen", "webcamd"]
```

### Security

Marlinraker, like Moonraker, has no user authentication: anyone who can reach port 7125 can control
the printer. Only expose it on networks you trust.

To stop arbitrary websites you visit from sending commands to your printer through your browser,
requests from a browser are only accepted when the page comes from the same hostname as
Marlinraker or matches `web.cors_domains`. Patterns may use `*` as a wildcard; patterns without
`://` match the hostname only. Requests without an `Origin` header (slicers, scripts) are not
affected. Set `cors_domains = ["*"]` to allow every origin.

`machine.reboot` and `machine.shutdown` run `systemctl reboot` and `systemctl poweroff`, which the
`marlinraker` user may need to be allowed to do (for example with a polkit rule).

### Printer

```toml
[printer]
bed_mesh = false
axis_minimum = [0, 0, 0]
axis_maximum = [220, 220, 240]

[printer.extruder]
min_temp = 0
max_temp = 250
min_extrude_temp = 180
filament_diameter = 1.75

[printer.heater_bed]
min_temp = 0
max_temp = 100

[printer.gcode]
send_m73 = true         # send print progress to the printer display
report_velocity = true  # poll the position more often to show the current speed
```

### Macros

Macros work like Klipper's `gcode_macro`. Defining a macro with the name of a built-in command
(`PAUSE`, `RESUME`, `CANCEL_PRINT`, ...) renames the built-in to `<NAME>_BASE`, or to
`rename_existing`.

```toml
[macros.cancel_print]
rename_existing = "cancel_print_base"
gcode = """
  CANCEL_PRINT_BASE
  _PARK_TOOLHEAD
"""

[macros._park_toolhead]
gcode = """
  G91
  G1 Z5 E-1 F900
  G90
  G1 X0 Y0 F2500
  TURN_OFF_HEATERS
"""
```

Macro G-code is a [Go template](https://pkg.go.dev/text/template) with the
[Sprig](https://masterminds.github.io/sprig/) functions. In the template:

- `.Params` holds the macro parameters by lower-case name and `.RawParams` the raw arguments,
- `.Printer` holds the printer objects, e.g. `.Printer.print_stats.state`; variables from
  `[macros.<name>.variables]` are in the `gcode_macro <NAME>` object, e.g.
  `(index .Printer "gcode_macro PAUSE").park_z`,
- `.ActionRespondInfo "message"` prints a message to the console and
  `.ActionRaiseError "message"` aborts the macro with an error.

```toml
[macros.sdcard_print_file]
rename_existing = "sdcard_print_file_base"
gcode = """
  {{ if eq .Params.filename "" }}
    {{ .ActionRaiseError "File name not provided" }}
  {{ else }}
    SDCARD_PRINT_FILE_BASE FILENAME="{{ .Params.filename }}"
  {{ end }}
"""
```

## Development

Marlinraker is written in Go. The commands below need Go 1.26 or newer and, for linting,
[golangci-lint](https://golangci-lint.run/) v2.

```bash
make build   # build ./marlinraker for this machine
make test    # run the tests with the race detector
make lint    # run golangci-lint
```

Hardware-specific code (`/proc`, `vcgencmd`, systemd) has mocks for non-Linux systems, so
Marlinraker can be developed on any OS. Without a printer it starts and serves the API, reporting
that no serial port was found.

Releases are built by [GoReleaser](https://goreleaser.com/) when a `v*` tag is pushed.

## License

[MIT](LICENSE)

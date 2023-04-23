package resources

import _ "embed"

//go:embed example_config/marlinraker.toml
var ExampleConfig string

//go:embed example_config/printers/generic.toml
var ExamplePrinterConfig string

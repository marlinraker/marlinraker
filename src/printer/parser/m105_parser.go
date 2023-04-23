package parser

import (
	"regexp"
	"strconv"
)

type Sensor struct {
	Temperature float64
}

type Heater struct {
	Sensor
	Target float64
	Power  float64
}

var (
	sensorRegex        = regexp.MustCompile(`([A-Z]+[0-9]*):([-0-9.]+)(?: /([-0-9.]+))?`)
	powerRegex         = regexp.MustCompile(`([A-Z]+)?@([0-9]+)?:([0-9]+)`)
	klipperHeaterNames = make(map[string]string)
)

func ParseM105(response string) (map[string]any, error) {

	temperatures := make(map[string]any)

	if matches := sensorRegex.FindAllStringSubmatch(response, -1); matches != nil {
		for _, match := range matches {

			isHeater := match[3] != "" && match[1] != "A"
			sensorName := klipperizeHeaterName(match[1], isHeater)

			temp, err := strconv.ParseFloat(match[2], 64)
			if err != nil {
				return nil, err
			}

			if isHeater {
				target, err := strconv.ParseFloat(match[3], 64)
				if err != nil {
					return nil, err
				}
				temperatures[sensorName] = Heater{
					Sensor: Sensor{temp},
					Target: target,
				}

			} else {
				temperatures[sensorName] = Sensor{
					Temperature: temp,
				}
			}
		}
	}

	if matches := powerRegex.FindAllStringSubmatch(response, -1); matches != nil {
		for _, match := range matches {
			heaterName := match[1]
			if heaterName == "" {
				heaterName = "T"
			}
			heaterName += match[2]
			heaterName = klipperizeHeaterName(heaterName, true)

			heater, exists := temperatures[heaterName]
			if !exists {
				continue
			}

			if heater, isHeater := heater.(Heater); isHeater {
				power, err := strconv.ParseFloat(match[3], 64)
				if err != nil {
					return nil, err
				}
				heater.Power = power / 127.0
			}
		}
	}

	return temperatures, nil
}

func klipperizeHeaterName(name string, isHeater bool) string {
	if klipperName, exists := klipperHeaterNames[name]; exists {
		return klipperName
	}

	var klipperName string
	if name[0] == 'T' {
		id := name[1:]
		if id == "0" {
			id = ""
		}
		klipperName = "extruder" + id
	} else {
		switch name {
		case "B":
			klipperName = "heater_bed"
		case "A":
			klipperName = "temperature_sensor ambient"
		case "P":
			klipperName = "temperature_sensor pinda"
		case "C":
			if isHeater {
				klipperName = "heater_generic chamber"
			} else {
				klipperName = "temperature_sensor chamber"
			}
		default:
			if isHeater {
				klipperName = "heater_generic " + name
			} else {
				klipperName = "temperature_sensor " + name
			}
		}
	}

	klipperHeaterNames[name] = klipperName
	return klipperName
}

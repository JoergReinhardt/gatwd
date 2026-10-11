package main

import _ "embed"

//go:embed patterns/gun_and_eater.life
var rawGunAndEater string

//go:embed patterns/gosper_gun.life
var rawGosperGun string

//go:embed patterns/pulsar.life
var rawPulsar string

//go:embed patterns/acorn.life
var rawAcorn string

//go:embed patterns/glider.life
var rawGlider string

const rawRPentomino = `#NAME R-pentomino
#DESCRIPTION The most famous active pentomino; runs for 1103 generations before stabilizing.
#SIZE 7x7
#RULE B3/S23
.......
..OO...
.OO....
..O....
.......
.......
.......
`

// GetPreset returns the Pattern associated with the given name.
func GetPreset(name string) (*Pattern, error) {
	switch name {
	case "gun-and-eater", "default":
		return ParseASCII(rawGunAndEater)
	case "gosper-gun", "gun":
		return ParseASCII(rawGosperGun)
	case "pulsar":
		return ParseASCII(rawPulsar)
	case "acorn":
		return ParseASCII(rawAcorn)
	case "r-pentomino", "rpentomino":
		return ParseASCII(rawRPentomino)
	case "glider":
		return ParseASCII(rawGlider)
	default:
		return ParseASCII(rawGunAndEater)
	}
}

// AvailablePresets lists all built-in pattern names.
var AvailablePresets = []string{
	"gun-and-eater (default - infinite glider emitter into sink)",
	"gosper-gun (classic Bill Gosper glider gun)",
	"pulsar (period-3 oscillator)",
	"acorn (methuselah, 5206 generations)",
	"r-pentomino (methuselah, 1103 generations)",
	"glider (elementary spaceship)",
}

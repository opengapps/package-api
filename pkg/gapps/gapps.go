//go:generate enumer -type=Android -json -transform=snake -trimprefix=Android
//go:generate enumer -type=Platform -json -transform=snake -trimprefix=Platform
//go:generate enumer -type=Variant -json -transform=snake -trimprefix=Variant

package gapps

import (
	"fmt"
)

// Platform is an enum for different chip architectures
type Platform uint

// Platform consts
const (
	PlatformArm Platform = iota
	PlatformArm64
	PlatformX86
	PlatformX86_64
)

// Android is an enum for different Android versions
type Android uint

// Android consts
const (
	Android44 Android = iota
	Android50
	Android51
	Android60
	Android70
	Android71
	Android80
	Android81
	Android90
	Android100
	Android110
)

// HumanString is required for human-readable Android version with . delimiter
func (a Android) HumanString() string {
	result := a.String()
	return result[:len(result)-1] + "." + result[len(result)-1:]
}

// Variant is an enum for different package variations
type Variant uint

// Variant consts
const (
	VariantTvstock Variant = iota
	VariantPico
	VariantNano
	VariantMicro
	VariantMini
	VariantFull
	VariantStock
	VariantSuper
	VariantAroma
	VariantTvmini
)

const parsingErrText = "parsing error: %w"

// ParsePackageParts helps to parse package info args into proper parts
func ParsePackageParts(args []string) (Platform, Android, Variant, error) {
	if len(args) != 3 {
		return 0, 0, 0, fmt.Errorf("bad number of arguments: want 4, got %d", len(args))
	}

	platform, err := PlatformString(args[0])
	if err != nil {
		return 0, 0, 0, fmt.Errorf(parsingErrText, err)
	}

	android, err := AndroidString(args[1])
	if err != nil {
		return 0, 0, 0, fmt.Errorf(parsingErrText, err)
	}

	variant, err := VariantString(args[2])
	if err != nil {
		return 0, 0, 0, fmt.Errorf(parsingErrText, err)
	}

	return platform, android, variant, nil
}

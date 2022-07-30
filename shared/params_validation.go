package shared

import (
	"log"
	"net"
	"sort"
)

func AssertParamIP(name, value string) error {
	if ip := net.ParseIP(value); ip == nil {
		log.Printf("Invalid parameter '%s' validated as ip address: %s\n", name, value)
		panic(ErrConfigurationValidationFailed)
	}
	return nil
}

func AssertParamPort(name string, value int) error {
	if value < 1 || value > 65536 {
		log.Printf("Invalid parameter '%s' validated as port [1-65536]: %d\n", name, value)
		panic(ErrConfigurationValidationFailed)
	}
	return nil
}

func AssertParamPortsDifferent(name string, values ...int) error {
	switch len(values) {
	case 0:
		fallthrough
	case 1:
		return nil

	case 2:
		if values[0] == values[1] {
			log.Printf("Ports '%s' must all be different: %v\n", name, values)
			panic(ErrConfigurationValidationFailed)
		}
	default:
		sort.Ints(values)
		for i := 1; i < len(values); i++ {
			if values[i-1] == values[i] {
				log.Printf("Ports '%s' must all be different: %v\n", name, values)
				panic(ErrConfigurationValidationFailed)
			}
		}
	}

	return nil
}

func AssertParamHostsDifferent(name string, values ...string) error {
	switch len(values) {
	case 0:
		fallthrough
	case 1:
		return nil

	case 2:
		if values[0] == values[1] {
			log.Printf("Addresses '%s' must all be different: %v\n", name, values)
			panic(ErrConfigurationValidationFailed)
		}
	default:
		sort.Strings(values)
		for i := 1; i < len(values); i++ {
			if values[i-1] == values[i] {
				log.Printf("Addresses '%s' must all be different: %v\n", name, values)
				panic(ErrConfigurationValidationFailed)
			}
		}
	}

	return nil
}

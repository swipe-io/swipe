package config

import "github.com/swipe-io/swipe/v3/option"

type Entity struct {
	Value *option.NamedType
}

// Config
// @swipe:"EventSourcing"
type Config struct {
	Entity Entity
}

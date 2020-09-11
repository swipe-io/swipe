package stcreator

import (
	"errors"
	"fmt"

	"gopkg.in/yaml.v3"
)

type yamlLoader struct {
	Type   string    `yaml:"type"`
	Params yaml.Node `yaml:"params"`
}

type Loaders []LoaderParams

type LoaderParams interface {
	Name() string
	Process() ([]StructMetadata, error)
}

func (l *Loaders) MarshalYAML() (interface{}, error) {
	return l, nil
}

func (l *Loaders) UnmarshalYAML(node *yaml.Node) error {
	var yamlLoaders []yamlLoader
	var dt LoaderParams
	if err := node.Decode(&yamlLoaders); err != nil {
		return errors.New(err.Error())
	}
	ll := make([]LoaderParams, len(yamlLoaders))
	for i, loader := range yamlLoaders {
		if f, ok := LoaderFactories[loader.Type]; ok {
			dt = f()
			if err := loader.Params.Decode(dt); err != nil {
				return err
			}
			ll[i] = dt
		} else {
			return fmt.Errorf("could not find loader type %s", loader.Type)
		}
	}
	*l = ll
	return nil
}

var LoaderFactories = map[string]func() LoaderParams{
	new(MongoLoader).Name(): func() LoaderParams {
		return new(MongoLoader)
	},
	new(PostgresLoader).Name(): func() LoaderParams {
		return new(PostgresLoader)
	},
}

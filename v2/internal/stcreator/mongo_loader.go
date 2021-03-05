package stcreator

type MongoLoader struct {
	Host string `yaml:"host"`
}

func (*MongoLoader) Name() string {
	return "mongo"
}

func (*MongoLoader) Process() (result []StructMetadata, err error) {
	return
}

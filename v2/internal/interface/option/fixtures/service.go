package fixtures

type Service interface {
	TestMethod(name string) error
}

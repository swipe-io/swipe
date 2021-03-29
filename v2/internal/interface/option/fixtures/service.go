package fixtures

type ServiceA interface {
	TestMethod(name string) error
}

type serviceA struct {
}

func (s *serviceA) TestMethod(name string) error {
	return nil
}

package fixtures

type ServiceA interface {
	// TestMethod dvsdvsdvsdvsdv
	// @name sdvsvsdv
	TestMethod(name string) error
	// TestMethod2 sdvsdsdv
	// @name sdvsvsdv
	TestMethod2(name string) error
}

type serviceA struct {
}

func (s *serviceA) TestMethod(name string) error {
	return nil
}

func (s *serviceA) TestMethod2(name string) error {
	return Test()
}

func Test() error {
	return nil
}

package fixtures

type Config struct {
	Name string
}

func SwipeConfig() {
	Build(
		ConfigEnv(&Config{}),
	)
}

func Swipe() {
	Build(
		Service(
			Interface((*ServiceA)(nil), "test"),
			Interface((*ServiceA)(nil), "test"),

			HTTPServer(),
			HTTPFast(),

			OpenapiTags([]interface{}{ServiceA.TestMethod2}, []string{"no"}),
			OpenapiInfo("title", "descr", "ok"),

			MethodOptions(ServiceA.TestMethod,
				RESTQueryVars("ok", "no"),
			),

			MethodOptions(ServiceA.TestMethod2,
				RESTQueryVars("ok1", "no2"),
				RESTMethod("POST"),
			),

			MethodDefaultOptions(
				RESTQueryVars("ok1", "no2"),
				RESTMethod("POST"),
			),
		),
	)
}

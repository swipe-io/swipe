package fixtures

func Swipe() {
	Build(
		Service(
			Interface((*ServiceA)(nil), ""),
		),
	)
}

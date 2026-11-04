package core

func b2f(b bool) float64 {
	if b {
		return 1
	}
	return 0
}

// guard is the setup line that stops a snippet on an error.
const guard = "if err != nil {\n\t\tpanic(err)\n\t}\n\t"

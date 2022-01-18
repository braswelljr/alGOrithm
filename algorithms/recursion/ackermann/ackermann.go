package ackermann

// Ackermann function
// https://en.wikipedia.org/wiki/Ackermann_function

func Ackermann(m, n int) int {

  if m == 0 {
    return n + 1
  }

  if n == 0 {
    return Ackermann(m-1, 1)
  }

  return Ackermann(m-1, Ackermann(m, n-1))
}

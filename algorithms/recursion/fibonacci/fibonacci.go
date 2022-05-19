package fibonacci

// Fibonacci returns the nth number in the Fibonacci sequence.
// The first two numbers in the sequence are 0 and 1.
// The nth number is the sum of the n-1 and n-2 numbers in the sequence.
func Fibonacci(n int) int {
  if n < 0 {
    return Fibonacci(n*-1) * -1 // for negative numbers
  } else if n == 0 || n == 1 {
    return n // base case
  }

  return Fibonacci(n-1) + Fibonacci(n-2)
}

//go:build !prod

package testhelpers

// RunConcurrently runs a given function concurrently.
func RunConcurrently(funcToRun func(), threadsNumber int) {
	done := make(chan bool, threadsNumber)
	for range threadsNumber {
		go func() {
			defer func() {
				done <- true
			}()
			funcToRun()
		}()
	}
	for range threadsNumber {
		<-done
	}
}

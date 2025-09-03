//go:build !prod

package testhelpers

// RunConcurrently runs a given function concurrently.
func RunConcurrently(funcToRun func(), threadsNumber int) {
	done := make(chan bool, threadsNumber)
	for i := 0; i < threadsNumber; i++ {
		go func() {
			defer func() {
				done <- true
			}()
			funcToRun()
		}()
	}
	for i := 0; i < threadsNumber; i++ {
		<-done
	}
}

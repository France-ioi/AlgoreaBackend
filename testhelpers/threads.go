// +build !prod

package testhelpers

// RunConcurrently runs a given function concurrently.
func RunConcurrently(f func(), threadsNumber int) {
	done := make(chan bool, threadsNumber)
	for i := 0; i < threadsNumber; i++ {
		go func() {
			defer func() {
				done <- true
			}()
			f()
		}()
	}
	for i := 0; i < threadsNumber; i++ {
		<-done
	}
}

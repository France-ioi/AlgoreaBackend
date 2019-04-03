package testhelpers

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

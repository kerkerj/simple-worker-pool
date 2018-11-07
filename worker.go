package worker

// Job interface is created for Pool,
// any strcut that implements Perform() can be processed by Pool
type Job interface {
	Perform()
}

// Pool
type Pool struct {
	JobsInChan    chan Job
	ResultOutChan chan Job
	workerCount   int
}

// NewWorkerPool - initialize a simple worker pool to execute Job.Perform().
func NewWorkerPool(bufferLength int, workerCount int) *Pool {
	return &Pool{
		JobsInChan:    make(chan Job, bufferLength),
		ResultOutChan: make(chan Job, bufferLength),
		workerCount:   workerCount,
	}
}

// Run - open goroutines to work by w.workerCount
func (w *Pool) Run() {
	for i := 0; i < w.workerCount; i++ {
		go w.work()
	}
}

// work - get a job from w.JobsInChan, and do Job.Perform, send back to w.ResultOutChan
// after w.JobsInChan closed, send a signal to w.closeChan to notify that this worker is out, no job will be processed.
func (w *Pool) work() {
	for job := range w.JobsInChan {
		job.Perform()
		w.ResultOutChan <- job
	}
}

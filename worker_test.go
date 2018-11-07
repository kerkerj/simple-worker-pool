package worker

import (
	"errors"
	"sync"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

type testJob struct {
	ID     int
	Result int

	Err error
}

func (t *testJob) Perform() {
	t.Result = t.ID * t.ID

	// test error
	if t.ID == 5 {
		t.Err = errors.New("err")
	}
}

func TestWorkerPool(t *testing.T) {
	Convey("TestWorkerPool", t, func() {
		Convey("With Empty Arguments", func() {
			// Arrange/Act
			workerPool := NewWorkerPool(0, 0)

			// Assert
			So(workerPool.JobsInChan, ShouldNotBeNil)
			So(workerPool.ResultOutChan, ShouldNotBeNil)
			So(workerPool.workerCount, ShouldEqual, 0)
		})

		Convey("With Jobs", func() {
			// test data
			// e.g. input: 2, expected result: 2*2 = 4
			testCases := map[int]int{}
			testCases[1] = 1
			testCases[2] = 4
			testCases[3] = 9
			testCases[4] = 16

			// Arrange
			workerPool := NewWorkerPool(4, 3)
			workerPool.Run()

			// Act
			for i := 1; i <= 4; i++ {
				workerPool.JobsInChan <- &testJob{
					ID: i,
				}
			}
			close(workerPool.JobsInChan)

			// Assert
			for i := 1; i <= 4; i++ {
				item := <-workerPool.ResultOutChan
				result, _ := item.(*testJob)

				So(result.Result, ShouldEqual, testCases[result.ID])
				So(result.Err, ShouldBeNil)
			}
		})

		Convey("With Error Jobs", func() {
			// test data
			// e.g. input: 2, expected result: 2*2 = 4
			testCases := map[int]int{}
			testCases[5] = 25

			// Arrange
			workerPool := NewWorkerPool(1, 3)
			workerPool.Run()

			// Act
			workerPool.JobsInChan <- &testJob{
				ID: 5,
			}
			close(workerPool.JobsInChan)

			// Assert
			item := <-workerPool.ResultOutChan
			result, _ := item.(*testJob)

			So(result.Result, ShouldEqual, testCases[result.ID])
			So(result.Err.Error(), ShouldEqual, "err")
		})

	})
}

type testJob2 struct {
	Time    time.Duration
	Success bool
}

func (t *testJob2) Perform() {
	time.Sleep(t.Time)
	t.Success = true
}

func TestLotsOfJobs(t *testing.T) {
	Convey("TestLotsOfJobs with low bufferLength", t, func() {
		// Arrange
		workerPool := NewWorkerPool(1, 10)
		workerPool.Run()

		Convey("Should not be blocked", func() {
			wg := sync.WaitGroup{}
			// Act
			// insert job
			wg.Add(1)
			go func() {
				for i := 0; i < 100; i++ {
					workerPool.JobsInChan <- &testJob2{
						Time: 1*time.Second + time.Duration(i)*time.Millisecond,
					}
				}
				wg.Done()
			}()

			hasError := false

			// Assert
			wg.Add(1)
			go func() {
				for i := 0; i < 100; i++ {
					data := <-workerPool.ResultOutChan
					r := data.(*testJob2)

					if !r.Success {
						hasError = true
					}
				}
				wg.Done()
			}()
			wg.Wait()

			So(hasError, ShouldBeFalse)
		})
	})
}

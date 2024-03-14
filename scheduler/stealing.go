package scheduler

import (
	"math"
	"math/rand"
	"parallel_model_evaluator/metrics"
	"parallel_model_evaluator/worksteal"
	"sync"
	"sync/atomic"

	"github.com/grassmudhorses/vader-go/lexicon"
	"github.com/grassmudhorses/vader-go/sentitext"
)

// Parallel worker which has a deque of tasks and steals from other workers when idle
func stealer(threadMapping map[int]*worksteal.Deque, ownIdx int, threshold float32, wg *sync.WaitGroup, metricScores *metrics.Metrics, globalCounter *int32) {
	dq := threadMapping[ownIdx]
	updatedCounter := false // Flag to indicate whether own dequeue is empty
	var victimIdx int
	for {

		var record []string
		curTask := dq.PopBottom() // Try taking task from own deque
		if curTask != nil {
			record = *curTask.Record
		} else {
			// If failed, then deque empty
			// Break if global counter = number of threads, i.e., all queues are empty

			if int(atomic.LoadInt32(globalCounter)) == len(threadMapping) {
				break
			}
			if !updatedCounter {
				// Deque just became empty. Increment global counter and update own flag
				atomic.AddInt32(globalCounter, 1)
				updatedCounter = true
			}
			// Get random victim thread to steal from
			victimIdx = rand.Intn(len(threadMapping))
			if victimIdx == ownIdx {
				continue
			}
			victimThread := threadMapping[victimIdx]
			stolenTask := victimThread.PopTop()
			// If stole task then process it, otherwise try to steal again
			if stolenTask != nil {
				record = *stolenTask.Record
			} else {
				continue
			}
		}

		// Process task
		text, label := record[1], record[0]
		parsedtext := sentitext.Parse(text, lexicon.DefaultLexicon)
		negativeScore := float32(sentitext.PolarityScore(parsedtext).Negative)

		metrics.CalculateMetrics(negativeScore, label, threshold, metricScores)

	}
	wg.Done()
}

// Parallel worker where each thread has its own deque but does not steal from other workers
func nonstealWorker(threadMapping map[int]*worksteal.Deque, ownIdx int, threshold float32, wg *sync.WaitGroup, metricScores *metrics.Metrics) {
	dq := threadMapping[ownIdx]
	for {
		// Get own task and process
		curTask := dq.PopBottom()
		if curTask != nil {
			record := *curTask.Record
			text, label := record[1], record[0]

			parsedtext := sentitext.Parse(text, lexicon.DefaultLexicon)
			negativeScore := float32(sentitext.PolarityScore(parsedtext).Negative)

			metrics.CalculateMetrics(negativeScore, label, threshold, metricScores)
		} else {
			// If failed, work done, exit
			break
		}
	}
	wg.Done()
}

// Runs the parallel deque implementation with/without workstealing based on input
func RunParallelSteal(config Config, records [][]string, thresholds []float32, optimalMetrics *metrics.OptimalMetrics, steal bool) {
	threadNum := config.ThreadNum
	var wg sync.WaitGroup
	numElements := len(records)
	numElementsPerThread := int(math.Ceil(float64(numElements) / float64(threadNum)))
	thresholdScores := make(map[float32]metrics.Metrics)
	threadMapping := make(map[int]*worksteal.Deque) // map thread number to thread deques
	var globalCounter int32                         // Global counter to represent when all queues are empty

	for _, threshold := range thresholds {
		metricScores := make([]metrics.Metrics, threadNum*16) // Avoid cache invalidation
		globalCounter = 0                                     // Reset counter
		// Need to create threadMapping before spawning any threads to ensure no concurrent read/writes
		for i := 0; i < threadNum; i++ {
			chunkStart := numElementsPerThread * i
			if chunkStart > numElements {
				chunkStart = numElements
			}
			chunkEnd := chunkStart + numElementsPerThread
			if chunkEnd > numElements {
				chunkEnd = numElements
			}
			// Create new deque with given bounds
			threadDq := worksteal.NewDeque(&records, chunkStart, chunkEnd, i)
			threadMapping[i] = threadDq
		}

		for i := 0; i < threadNum; i++ {
			chunkStart := numElementsPerThread * i
			if chunkStart > numElements {
				chunkStart = numElements
			}
			chunkEnd := chunkStart + numElementsPerThread
			if chunkEnd > numElements {
				chunkEnd = numElements
			}
			// Spawn threads in correct mode
			wg.Add(1)
			if !steal {
				go nonstealWorker(threadMapping, i, threshold, &wg, &metricScores[i*16])
			} else {
				go stealer(threadMapping, i, threshold, &wg, &metricScores[i*16], &globalCounter)
			}

		}
		wg.Wait()

		// Reconcile results from all threads and update metrics
		var correctP, correctR, totalP, totalR = 0, 0, 0, 0
		for j := 0; j < threadNum; j++ {
			correctP += metricScores[j*16].CorrectP
			correctR += metricScores[j*16].CorrectR
			totalP += metricScores[j*16].TotalP
			totalR += metricScores[j*16].TotalR
		}
		var precision float32
		if totalP == 0 {
			precision = 0
		} else {
			precision = float32(correctP) / float32(totalP)
		}
		combinedMetrics := &metrics.Metrics{Precision: precision, Recall: float32(correctR) / float32(totalR)}
		thresholdScores[threshold] = *combinedMetrics
		metrics.UpdateOptimalMetrics(combinedMetrics, optimalMetrics, threshold)
	}

	// Generate output
	metrics.PrintMetrics(&thresholdScores, optimalMetrics)
	metrics.PRCurve(&thresholdScores, thresholds)
}

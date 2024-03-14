package scheduler

import (
	"math"
	"parallel_model_evaluator/barrier"
	"parallel_model_evaluator/metrics"
)

// Parallel worker that works on a single chunk and then waits at the barrier
func parallelWorker(records *[][]string, chunkStart int, chunkEnd int, threshold float32, b *barrier.Barrier, metricScores *metrics.Metrics) {
	metrics.ChunkMetrics(records, chunkStart, chunkEnd, metricScores, threshold)
	b.Await()
}

// Runs the normal parallel version which splits and assigns chunks of the file to each worker
func RunParallelNormal(config Config, records [][]string, thresholds []float32, optimalMetrics *metrics.OptimalMetrics) {
	threadNum := config.ThreadNum
	numElements := len(records)
	numElementsPerThread := int(math.Ceil(float64(numElements) / float64(threadNum)))
	thresholdScores := make(map[float32]metrics.Metrics)
	b := barrier.NewBarrier()
	for _, threshold := range thresholds {
		// Re-initialize barrier counters
		b.Ctr = 0
		b.TotalCount = 1                                      // Set as 1 as main goroutine also needs to wait at the barrier
		metricScores := make([]metrics.Metrics, threadNum*16) // Take into account cache invalidation
		for i := 0; i < threadNum; i++ {
			chunkStart := numElementsPerThread * i
			if chunkStart > numElements {
				chunkStart = numElements
			}
			chunkEnd := chunkStart + numElementsPerThread
			if chunkEnd > numElements {
				chunkEnd = numElements
			}
			b.TotalCount += 1
			go parallelWorker(&records, chunkStart, chunkEnd, threshold, b, &metricScores[i*16])
		}
		b.Await()

		// Reconcile the results from all chunks and update the metrics
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

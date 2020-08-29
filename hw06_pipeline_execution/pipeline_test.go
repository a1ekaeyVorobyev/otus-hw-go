package hw06_pipeline_execution //nolint:golint,stylecheck

import (
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

const (
	sleepPerStage = time.Millisecond * 100
	fault         = sleepPerStage / 2
)

func TestPipeline(t *testing.T) {
	// Stage generator
	g := func(name string, f func(v interface{}) interface{}) Stage {
		return func(in In) Out {
			out := make(Bi)
			go func() {
				defer close(out)
				for v := range in {
					time.Sleep(sleepPerStage)
					out <- f(v)
				}
			}()
			return out
		}
	}

	stages := []Stage{
		g("Dummy", func(v interface{}) interface{} { return v }),
		g("Multiplier (* 2)", func(v interface{}) interface{} { return v.(int) * 2 }),
		g("Adder (+ 100)", func(v interface{}) interface{} { return v.(int) + 100 }),
		g("Stringifier", func(v interface{}) interface{} { return strconv.Itoa(v.(int)) }),
	}

	t.Run("simple case", func(t *testing.T) {
		in := make(Bi)
		data := []int{1, 2, 3, 4, 5}

		go func() {
			for _, v := range data {
				in <- v
			}
			close(in)
		}()

		result := make([]string, 0, 10)
		start := time.Now()
		for s := range ExecutePipeline(in, nil, stages...) {
			result = append(result, s.(string))
		}
		elapsed := time.Since(start)

		require.Equal(t, result, []string{"102", "104", "106", "108", "110"})
		require.Less(t,
			int64(elapsed),
			// ~0.8s for processing 5 values in 4 stages (100ms every) concurrently
			int64(sleepPerStage)*int64(len(stages)+len(data)-1)+int64(fault))
	})

	t.Run("done case", func(t *testing.T) {
		in := make(Bi)
		done := make(Bi)
		data := []int{1, 2, 3, 4, 5}

		// Abort after 200ms
		abortDur := sleepPerStage * 2
		go func() {
			<-time.After(abortDur)
			close(done)
		}()

		go func() {
			for _, v := range data {
				in <- v
			}
			close(in)
		}()

		result := make([]string, 0, 10)
		start := time.Now()
		for s := range ExecutePipeline(in, done, stages...) {
			result = append(result, s.(string))
		}
		elapsed := time.Since(start)

		require.Len(t, result, 0)
		require.Less(t, int64(elapsed), int64(abortDur)+int64(fault))
	})
}

func TestPipelineMy(t *testing.T) {
	// Stage generator
	g := func(in In) Out {
		out := make(Bi)
		go func() {
			defer func() {
				if !IsClosed(out) {
					close(out)
				}
			}()
			for v := range in {
				out <- Factorial(v.(uint64))
			}
		}()
		return out
	}

	g1 := func(in In) Out {
		out := make(Bi)
		go func() {
			defer func() {
				if !IsClosed(out) {
					close(out)
				}
			}()
			for v := range in {
				out <- (v.(uint64) + 2)
			}
		}()
		return out
	}

	t.Run("Check Stage", func(t *testing.T) {
		in := make(Bi)
		data := []uint64{0, 1, 2, 3, 4}
		flag := true
		go func() {
			for _, v := range data {
				in <- v
			}
			flag = false
			close(in)
		}()
		result := make([]uint64, 0, len(data))
		for flag {
			for s := range g(in) {
				result = append(result, s.(uint64))
			}
		}
		require.Equal(t, result, []uint64{1, 1, 2, 6, 24})
	})

	t.Run("Check Gage", func(t *testing.T) {
		in := make(Bi)
		data := []uint64{0, 1, 2, 3, 4}
		flag := true
		go func() {
			for _, v := range data {
				in <- v
			}
			flag = false
			close(in)
		}()
		result := make([]uint64, 0, len(data))
		stage := []Stage{g, g1}
		for flag {
			for s := range ExecutePipeline(in, nil, stage...) {
				result = append(result, s.(uint64))
			}
		}
		require.Equal(t, result, []uint64{3, 3, 4, 8, 26})
	})

	t.Run("Check Done", func(t *testing.T) {
		in := make(Bi)
		done := make(Bi)
		data := []uint64{0, 1, 2, 3, 4}
		flag := true
		go func() {
			for i, v := range data {
				if i == 3 {
					done <- v
					break
				}
				in <- v
			}
			flag = false
			close(in)
		}()
		result := make([]uint64, 0, 3)
		stage := []Stage{g, g1}
		for flag {
			for s := range ExecutePipeline(in, done, stage...) {
				result = append(result, s.(uint64))
			}
		}
		require.Equal(t, result, []uint64{3, 3, 4})
	})
}

func IsClosed(ch Bi) bool {
	select {
	case <-ch:
		return true
	default:
	}

	return false
}

func Factorial(n uint64) (result uint64) {
	if n > 0 {
		result = n * Factorial(n-1)
		return result
	}
	return 1
}

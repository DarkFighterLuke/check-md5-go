package main

import (
	"runtime"
	"sync"
)

// MergeSort performs the merge sort algorithm taking advantage of multiple processors.
func MergeSort(src *[]string) {
	// We subtract 1 goroutine which is the one we are already running in.
	extraGoroutines := runtime.NumCPU() - 1
	semChan := make(chan struct{}, extraGoroutines)
	defer close(semChan)
	mergesort(src, semChan)
}

func mergesort(src *[]string, semChan chan struct{}) {
	if len(*src) <= 1 {
		return
	}

	mid := len(*src) / 2

	left := make([]string, mid)
	right := make([]string, len(*src)-mid)
	copy(left, (*src)[:mid])
	copy(right, (*src)[mid:])

	wg := sync.WaitGroup{}

	select {
	case semChan <- struct{}{}:
		wg.Add(1)
		go func() {
			mergesort(&left, semChan)
			<-semChan
			wg.Done()
		}()
	default:
		// Can't create a new goroutine, let's do the job ourselves.
		mergesort(&left, semChan)
	}

	mergesort(&right, semChan)

	wg.Wait()

	merge(src, &left, &right)
}

func merge(result, left, right *[]string) {
	var l, r, i int

	for l < len(*left) || r < len(*right) {
		if l < len(*left) && r < len(*right) {
			if (*left)[l] <= (*right)[r] {
				(*result)[i] = (*left)[l]
				l++
			} else {
				(*result)[i] = (*right)[r]
				r++
			}
		} else if l < len(*left) {
			(*result)[i] = (*left)[l]
			l++
		} else if r < len(*right) {
			(*result)[i] = (*right)[r]
			r++
		}
		i++
	}
}

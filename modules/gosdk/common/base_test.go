package common

import (
	"fmt"
	"time"
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestBase(t *testing.T) {
	p := NewBase()
	assert.Equal(t, 3, p.RetryCount)
	assert.Equal(t, 0, p.InstanceID)
	assert.Equal(t, 0, p.IterationCount)

	p.SetIterationInterval("3s")
	assert.Equal(t, 3, p.interval.mark)
	assert.Equal(t, FIXED, p.interval.intervalType)
	// Test the wait time will be roughly 3 seconds
	start := time.Now()
	p.Wait()
	waitTime := time.Now().Sub(start)
	fmt.Println("Fixed acutal wait time:", waitTime)
	value := waitTime >= time.Duration(3) * time.Second
	assert.Equal(t, true, value)

	p.SetIterationInterval("10r")
	assert.Equal(t, 10, p.interval.mark)
	assert.Equal(t, RANDOM, p.interval.intervalType)

	start = time.Now()
	p.Wait()
	end := time.Now()
	fmt.Println("Random acutal wait time:", end.Sub(start))
	waitTime = end.Add(time.Second).Sub(start)
	value = waitTime <= time.Duration(10) * time.Second
	assert.Equal(t, true, value)

	assert.Equal(t, p.Next(), 1)
	assert.Equal(t, p.CurrentIter(), 1)

	assert.Equal(t,p.CurrentRetry(),0)
	assert.Equal(t,p.NextRetry(),1)
	assert.Equal(t,p.ResetCurrentRetry(),0)
}

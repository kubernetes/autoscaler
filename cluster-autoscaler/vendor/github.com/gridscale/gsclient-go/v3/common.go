package gsclient

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type emptyStruct struct {
}

// retryableFunc defines a function that can be retried.
type retryableFunc func() (bool, error)

// isValidUUID validates the uuid.
func isValidUUID(u string) bool {
	_, err := uuid.Parse(u)
	return err == nil
}

// retryWithContext reruns a function until the context is done.
func retryWithContext(ctx context.Context, targetFunc retryableFunc, delay time.Duration) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			continueRetrying, err := targetFunc()
			if !continueRetrying {
				return err
			}
			time.Sleep(delay) //delay between retries
		}
	}
}

// retryNTimes reruns a function within a number of retries.
func retryNTimes(targetFunc retryableFunc, numOfRetries int, delay time.Duration) error {
	retryNo := 0
	var err error
	var continueRetrying bool
	for retryNo <= numOfRetries {
		continueRetrying, err = targetFunc()
		if !continueRetrying {
			return err
		}
		//delay between retries.
		retryNo++
		time.Sleep(delay * time.Duration(retryNo))
	}
	if err != nil {
		reqErr, ok := err.(RequestError)
		if ok {
			if reqErr.Description == "" {
				reqErr.Description = "no error message received from server"
			}
			reqErr.Description = fmt.Sprintf("Maximum number of re-tries has been exhausted with error: %s", reqErr.Description)
			return reqErr
		}
		return fmt.Errorf("maximum number of tries has been exhausted with error: %v", err)
	}
	return errors.New("maximum number of tries has been exhausted")
}

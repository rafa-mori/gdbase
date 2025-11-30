package cron

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"reflect"
	"time"

	"github.com/google/uuid"

	t "github.com/kubex-ecosystem/gdbase/internal/types"
	"github.com/kubex-ecosystem/logz"
	gl "github.com/kubex-ecosystem/logz"
	l "github.com/kubex-ecosystem/logz"
)

type ICronJobValidation interface {
	ValidateCronJobProperties(ctx context.Context, cron *CronJob, restrict bool) (*CronJob, error)
	ValidateCronJobRestrict(field any) error
}
type CronModelValidation struct {
	*t.Mutexes
	Logger        *logz.LoggerZ
	defaultValues map[string]any
}

func NewCronModelValidation(ctx context.Context, logger *logz.LoggerZ, debug bool) ICronJobValidation {
	if logger == nil {
		logger = l.GetLogger("CronModelValidation")
	}
	if debug {
		gl.SetDebug(true)
	}
	return &CronModelValidation{
		Mutexes: t.NewMutexesType(),
		Logger:  logger,
		defaultValues: map[string]any{
			"RetryInterval":    func() int { return rand.Intn(10) + 1 }(),
			"CreatedAt":        time.Now(),
			"UserID":           uuid.Nil,
			"Retries":          0,
			"ExecTimeout":      30,
			"MaxRetries":       3,
			"MaxExecutionTime": 150,
			"IsRecurring":      false,
			"IsActive":         true,
			"CronType":         "cron",
			"CronExpression":   "2 * * * *",
			"Command":          "echo 'Hello, World!'",
			"Method":           "GET",
		},
	}
}
func (cv *CronModelValidation) IsValidField(field any) bool {
	if field == nil {
		logz.Log("warn", fmt.Sprintf("Field is nil: %v", field))
		return false
	}
	if reflect.TypeOf(field).Kind() == reflect.Ptr {
		logz.Log("warn", fmt.Sprintf("Field is a pointer: %v", field))
		return false
	}
	if vl := reflect.ValueOf(field); !vl.IsValid() || !vl.IsNil() || vl.IsZero() {
		logz.Log("warn", fmt.Sprintf("Field is not valid: %v", field))
		return false
	}
	return true
}
func (cv *CronModelValidation) GetValueOrDefault(ctx context.Context, name string, field any) any {
	// If the field is not valid, search for a default value
	if defaultValue, ok := cv.defaultValues[name]; ok {
		logz.Log("notice", fmt.Sprintf("Field %s is not valid, using default value: %v", name, defaultValue))
		return defaultValue
	}
	// If no default value is found, return the original field
	return field // If the field is valid, return the field itself
}
func (cv *CronModelValidation) ValidateCronJobRestrict(field any) error {
	if !cv.IsValidField(field) {
		logz.Log("warn", fmt.Sprintf("Field is not valid: %v", field))
		return errors.New("field is not valid")
	}
	switch v := field.(type) {
	case string:
		if v == "" {
			logz.Log("warn", fmt.Sprintf("Field is empty: %v", field))
			return errors.New("field cannot be empty")
		}
	case int:
		if v <= 0 {
			logz.Log("warn", fmt.Sprintf("Field is not positive: %v", field))
			return errors.New("field must be positive")
		}
	case uuid.UUID:
		if v == uuid.Nil {
			logz.Log("warn", fmt.Sprintf("Field is not a valid UUID: %v", field))
			return errors.New("field must be a valid UUID")
		}
	}
	return nil
}
func (cv *CronModelValidation) ValidateCronJobProperties(ctx context.Context, cron *CronJob, restrict bool) (*CronJob, error) {
	// Check if the cron job is nil
	if cron == nil {
		logz.Log("error", "Cron job cannot be nil")
		return nil, errors.New("cron job cannot be nil")
	}

	// Get the values of the cron job struct
	val := reflect.ValueOf(cron).Elem()

	// Iterate over the fields of the struct
	for i := 0; i < val.NumField(); i++ {

		// Get the field name
		name := val.Type().Field(i)

		// Get the field value
		value := val.Field(i)

		// If it is a restricted validation, check the field value
		if restrict {
			err := cv.ValidateCronJobRestrict(value.Interface())
			if err != nil {
				logz.Log("warn", fmt.Sprintf("Field is not valid: %v", value.Interface()))
				return nil, err
			}
		}

		// Set the field value back to the struct
		value.Set(reflect.ValueOf(cv.GetValueOrDefault(ctx, name.Name, value.Interface())))
	}

	// Lock the mutex to restore the values back to the struct
	cv.MuLock()
	defer cv.MuUnlock()

	logz.Log("info", fmt.Sprintf("Cron job properties validated: %v", cron))

	// return the cron job with the validated and defaulted values
	return cron, nil
}

package websocket

import (
	"github.com/google/uuid"
	"golang.org/x/net/context"
	"time"
)

// OTP is a one-time password for websocket authentication
type OTP struct {
	Key     string
	Created time.Time
}

// RetentionMap is a map of OTPs with their keys as the map keys
type RetentionMap map[string]OTP

// NewRetentionMap creates a new retention map
func NewRetentionMap(ctx context.Context, retentionPeriod time.Duration) RetentionMap {
	rm := make(RetentionMap)

	//start the retention process with a goroutine
	go rm.Retention(ctx, retentionPeriod)

	return rm
}

// NewOTP creates a new OTP and adds it to the retention map
func (rm RetentionMap) NewOTP() OTP {
	o := OTP{
		Key:     uuid.NewString(),
		Created: time.Now(),
	}

	rm[o.Key] = o
	return o
}

// VerifyOTP verifies if the OTP is a valid password and deletes it if so & returns true (or false if not valid)
func (rm RetentionMap) VerifyOTP(otp string) bool {
	if _, ok := rm[otp]; !ok {
		return false //otp is not valid
	}
	//if it does exist, it deletes the one-time password and returns true
	delete(rm, otp)
	return true
}

// Retention checks for expired OTPs and removes them
func (rm RetentionMap) Retention(ctx context.Context, retentionPeriod time.Duration) {
	// time for re-checking one time passwords
	ticker := time.NewTicker(400 * time.Millisecond)

	for {
		select {
		case <-ticker.C:
			for _, otp := range rm {
				// if the otp is older than the retention period, delete it
				if otp.Created.Add(retentionPeriod).Before(time.Now()) {
					delete(rm, otp.Key)
				}
			}
		//when the context is done, stop the retention process
		case <-ctx.Done():
			return
		}
	}
}

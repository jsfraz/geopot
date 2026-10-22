package database

import (
	"context"
	"jsfraz/geopot/models"
	"jsfraz/geopot/utils"
	"log"
	"sync"
	"time"
)

const (
	// How often the server's own public IP info is refreshed.
	selfInfoRefreshInterval = 6 * time.Hour
	// Delay before retrying a failed self info refresh.
	selfInfoRetryDelay = time.Minute
	// Upper bound of a single upstream geolocation call.
	selfInfoTimeout = 15 * time.Second
	// Minimum gap between two lazily triggered refreshes.
	selfInfoTriggerGap = time.Minute
)

var (
	selfInfoMutex         sync.Mutex
	selfInfoTriggerMutex  sync.Mutex
	lastSelfInfoTriggerAt time.Time
)

// Insert connection record to TimescaleDB database.
//
//	@param connection
//	@return error
func InsertConnection(connection *models.Connection) error {
	// Ensure timestamp is in UTC
	connection.Timestamp = connection.Timestamp.UTC()
	// Use Create to add the record to the hypertable
	return utils.GetSingleton().Postgres.Create(connection).Error
}

// Fetches the server's own public IP info and stores it in Valkey.
//
//	@param ctx
//	@return error
func RefreshSelfIpInfo(ctx context.Context) error {
	selfInfoMutex.Lock()
	defer selfInfoMutex.Unlock()

	json, err := utils.GetIpInfo(ctx, nil)
	if err != nil {
		return err
	}
	// Unmarshal to struct
	connection, err := models.ConnectionFromJson(*json)
	if err != nil {
		return err
	}
	// Push to Valkey
	if err := PushSelfRecord(*connection); err != nil {
		return err
	}
	log.Printf("Server public IP info: %s", connection.IPAddress)
	return nil
}

// Refreshes the server's own IP info in the background, at most once a minute.
// Used when the "self" record is missing (e.g. after a Valkey flush).
func TriggerSelfIpRefresh() {
	selfInfoTriggerMutex.Lock()
	if time.Since(lastSelfInfoTriggerAt) < selfInfoTriggerGap {
		selfInfoTriggerMutex.Unlock()
		return
	}
	lastSelfInfoTriggerAt = time.Now()
	selfInfoTriggerMutex.Unlock()

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), selfInfoTimeout)
		defer cancel()
		if err := RefreshSelfIpInfo(ctx); err != nil {
			log.Printf("Failed to refresh server IP info: %v", err)
		}
	}()
}

// Keeps the server's own IP info up to date until ctx is cancelled. A failed
// upstream call is retried after a short delay instead of aborting the process.
//
//	@param ctx
func StartSelfInfoRefresher(ctx context.Context) {
	go func() {
		delay := time.Duration(0)
		for {
			select {
			case <-ctx.Done():
				return
			case <-time.After(delay):
			}

			refreshCtx, cancel := context.WithTimeout(context.Background(), selfInfoTimeout)
			err := RefreshSelfIpInfo(refreshCtx)
			cancel()

			if err != nil {
				log.Printf("Failed to refresh server IP info: %v (retrying in %s)", err, selfInfoRetryDelay)
				delay = selfInfoRetryDelay
				continue
			}
			delay = selfInfoRefreshInterval
		}
	}()
}

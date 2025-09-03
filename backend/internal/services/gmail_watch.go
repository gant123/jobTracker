package services

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	gmail "google.golang.org/api/gmail/v1"
)

type GmailWatchService struct {
	projectID string
	topicName string
	db        *sql.DB
}

func (s *GmailWatchService) SetupWatch(ctx context.Context, srv *gmail.Service, userID int) error {
	// Create watch on Gmail inbox
	watchReq := &gmail.WatchRequest{
		TopicName:         fmt.Sprintf("projects/%s/topics/%s", s.projectID, s.topicName),
		LabelIds:          []string{"INBOX"},
		LabelFilterAction: "include",
	}

	watchResp, err := srv.Users.Watch("me", watchReq).Do()
	if err != nil {
		return fmt.Errorf("failed to setup watch: %w", err)
	}

	// Store watch details
	_, err = s.db.Exec(`
        INSERT INTO gmail_sync_status (user_id, watch_expiration, last_history_id)
        VALUES ($1, $2, $3)
        ON CONFLICT (user_id) 
        DO UPDATE SET 
            watch_expiration = EXCLUDED.watch_expiration,
            last_history_id = EXCLUDED.last_history_id,
            updated_at = NOW()
    `, userID, time.Unix(0, watchResp.Expiration*1000000), watchResp.HistoryId)

	return err
}

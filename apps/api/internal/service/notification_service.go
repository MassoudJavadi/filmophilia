package service

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/MassoudJavadi/filmophilia/api/internal/db"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

var (
	ErrNotificationNotFound = errors.New("notification not found")
	ErrUnauthorized         = errors.New("unauthorized")
)

type NotificationService interface {
	CreateNotification(ctx context.Context, userID int32, notifType, title, content string, metadata map[string]interface{}) error
	GetUserNotifications(ctx context.Context, userID, limit, offset int32) ([]db.Notification, int32, int32, error)
	GetUnreadNotifications(ctx context.Context, userID, limit, offset int32) ([]db.Notification, int32, error)
	GetNotificationByID(ctx context.Context, notificationID int32) (db.Notification, error)
	MarkNotificationAsRead(ctx context.Context, userID, notificationID int32) error
	MarkAllAsRead(ctx context.Context, userID int32) error
	DeleteNotification(ctx context.Context, userID, notificationID int32) error
	DeleteAllNotifications(ctx context.Context, userID int32) error
	CleanupOldNotifications(ctx context.Context, olderThan time.Time) error
}

type notificationService struct {
	queries *db.Queries
}

func NewNotificationService(queries *db.Queries) NotificationService {
	return &notificationService{queries: queries}
}

func (s *notificationService) CreateNotification(ctx context.Context, userID int32, notifType, title, content string, metadata map[string]interface{}) error {
	// Marshal metadata to JSON
	var metadataBytes []byte
	if metadata != nil {
		var err error
		metadataBytes, err = json.Marshal(metadata)
		if err != nil {
			return err
		}
	}

	// Convert notification type string to enum
	var typeEnum db.NotificationType
	switch notifType {
	case "NEW_FOLLOWER":
		typeEnum = db.NotificationTypeNEWFOLLOWER
	case "NEW_LIKE":
		typeEnum = db.NotificationTypeNEWLIKE
	case "NEW_COMMENT":
		typeEnum = db.NotificationTypeNEWCOMMENT
	case "ACCOUNT_ACTIVATED":
		typeEnum = db.NotificationTypeACCOUNTACTIVATED
	case "ACCOUNT_SUSPENDED":
		typeEnum = db.NotificationTypeACCOUNTSUSPENDED
	case "ACCOUNT_BANNED":
		typeEnum = db.NotificationTypeACCOUNTBANNED
	case "SYSTEM_ALERT":
		typeEnum = db.NotificationTypeSYSTEMALERT
	default:
		typeEnum = db.NotificationTypeSYSTEMALERT
	}

	contentText := pgtype.Text{}
	if content != "" {
		contentText = pgtype.Text{String: content, Valid: true}
	}

	_, err := s.queries.CreateNotification(ctx, db.CreateNotificationParams{
		UserID:   userID,
		Type:     typeEnum,
		Title:    title,
		Content:  contentText,
		Metadata: metadataBytes,
	})

	return err
}

func (s *notificationService) GetUserNotifications(ctx context.Context, userID, limit, offset int32) ([]db.Notification, int32, int32, error) {
	notifications, err := s.queries.GetUserNotifications(ctx, db.GetUserNotificationsParams{
		UserID: userID,
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, 0, 0, err
	}

	total, err := s.queries.CountUserNotifications(ctx, userID)
	if err != nil {
		return nil, 0, 0, err
	}

	unreadCount, err := s.queries.CountUnreadNotifications(ctx, userID)
	if err != nil {
		return nil, 0, 0, err
	}

	return notifications, total, unreadCount, nil
}

func (s *notificationService) GetUnreadNotifications(ctx context.Context, userID, limit, offset int32) ([]db.Notification, int32, error) {
	notifications, err := s.queries.GetUnreadNotifications(ctx, db.GetUnreadNotificationsParams{
		UserID: userID,
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, 0, err
	}

	count, err := s.queries.CountUnreadNotifications(ctx, userID)
	if err != nil {
		return nil, 0, err
	}

	return notifications, count, nil
}

func (s *notificationService) GetNotificationByID(ctx context.Context, notificationID int32) (db.Notification, error) {
	notification, err := s.queries.GetNotificationByID(ctx, notificationID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return db.Notification{}, ErrNotificationNotFound
		}
		return db.Notification{}, err
	}
	return notification, nil
}

func (s *notificationService) MarkNotificationAsRead(ctx context.Context, userID, notificationID int32) error {
	// Verify notification belongs to user before marking as read
	notification, err := s.queries.GetNotificationByID(ctx, notificationID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNotificationNotFound
		}
		return err
	}

	if notification.UserID != userID {
		return ErrUnauthorized
	}

	return s.queries.MarkNotificationAsRead(ctx, db.MarkNotificationAsReadParams{
		ID:     notificationID,
		UserID: userID,
	})
}

func (s *notificationService) MarkAllAsRead(ctx context.Context, userID int32) error {
	return s.queries.MarkAllNotificationsAsRead(ctx, userID)
}

func (s *notificationService) DeleteNotification(ctx context.Context, userID, notificationID int32) error {
	// Verify notification belongs to user before deleting
	notification, err := s.queries.GetNotificationByID(ctx, notificationID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNotificationNotFound
		}
		return err
	}

	if notification.UserID != userID {
		return ErrUnauthorized
	}

	return s.queries.DeleteNotification(ctx, db.DeleteNotificationParams{
		ID:     notificationID,
		UserID: userID,
	})
}

func (s *notificationService) DeleteAllNotifications(ctx context.Context, userID int32) error {
	return s.queries.DeleteAllNotifications(ctx, userID)
}

func (s *notificationService) CleanupOldNotifications(ctx context.Context, olderThan time.Time) error {
	return s.queries.DeleteOldNotifications(ctx, pgtype.Timestamptz{
		Time:  olderThan,
		Valid: true,
	})
}

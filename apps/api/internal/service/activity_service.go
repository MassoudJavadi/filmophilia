package service

import (
	"context"
	"encoding/json"

	"github.com/MassoudJavadi/filmophilia/api/internal/db"
)

type ActivityService interface {
	CreateActivity(ctx context.Context, userID int32, action, entityType string, entityID int32, metadata map[string]interface{}) error
	GetUserActivities(ctx context.Context, userID, limit, offset int32) ([]db.GetUserActivitiesRow, int32, error)
	GetFollowingActivities(ctx context.Context, userID, limit, offset int32) ([]db.GetFollowingActivitiesRow, int32, error)
	GetActivitiesByEntity(ctx context.Context, entityType, entityID, limit, offset int32) ([]db.GetActivitiesByEntityRow, error)
	DeleteUserActivities(ctx context.Context, userID int32) error
	DeleteActivityByEntity(ctx context.Context, entityType string, entityID int32) error
}

type activityService struct {
	queries *db.Queries
}

func NewActivityService(queries *db.Queries) ActivityService {
	return &activityService{queries: queries}
}

func (s *activityService) CreateActivity(ctx context.Context, userID int32, action, entityType string, entityID int32, metadata map[string]interface{}) error {
	// Marshal metadata to JSON
	var metadataBytes []byte
	if metadata != nil {
		var err error
		metadataBytes, err = json.Marshal(metadata)
		if err != nil {
			return err
		}
	}

	// Convert entity type string to enum
	var entityTypeEnum db.EntityType
	switch entityType {
	case "MOVIE":
		entityTypeEnum = db.EntityTypeMOVIE
	case "REVIEW":
		entityTypeEnum = db.EntityTypeREVIEW
	case "COMMENT":
		entityTypeEnum = db.EntityTypeCOMMENT
	case "USER":
		entityTypeEnum = db.EntityTypeUSER
	default:
		entityTypeEnum = db.EntityTypeMOVIE
	}

	_, err := s.queries.CreateActivity(ctx, db.CreateActivityParams{
		UserID:     userID,
		Action:     action,
		EntityType: entityTypeEnum,
		EntityID:   entityID,
		Metadata:   metadataBytes,
	})

	return err
}

func (s *activityService) GetUserActivities(ctx context.Context, userID, limit, offset int32) ([]db.GetUserActivitiesRow, int32, error) {
	activities, err := s.queries.GetUserActivities(ctx, db.GetUserActivitiesParams{
		UserID: userID,
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, 0, err
	}

	count, err := s.queries.CountUserActivities(ctx, userID)
	if err != nil {
		return nil, 0, err
	}

	return activities, count, nil
}

func (s *activityService) GetFollowingActivities(ctx context.Context, userID, limit, offset int32) ([]db.GetFollowingActivitiesRow, int32, error) {
	activities, err := s.queries.GetFollowingActivities(ctx, db.GetFollowingActivitiesParams{
		FollowerID: userID,
		Limit:      limit,
		Offset:     offset,
	})
	if err != nil {
		return nil, 0, err
	}

	count, err := s.queries.CountFollowingActivities(ctx, userID)
	if err != nil {
		return nil, 0, err
	}

	return activities, count, nil
}

func (s *activityService) GetActivitiesByEntity(ctx context.Context, entityType, entityID, limit, offset int32) ([]db.GetActivitiesByEntityRow, error) {
	// Convert entityType int to enum
	var entityTypeEnum db.EntityType
	switch entityType {
	case 1:
		entityTypeEnum = db.EntityTypeMOVIE
	case 2:
		entityTypeEnum = db.EntityTypeREVIEW
	case 3:
		entityTypeEnum = db.EntityTypeCOMMENT
	case 4:
		entityTypeEnum = db.EntityTypeUSER
	default:
		entityTypeEnum = db.EntityTypeMOVIE
	}

	return s.queries.GetActivitiesByEntity(ctx, db.GetActivitiesByEntityParams{
		EntityType: entityTypeEnum,
		EntityID:   entityID,
		Limit:      limit,
		Offset:     offset,
	})
}

func (s *activityService) DeleteUserActivities(ctx context.Context, userID int32) error {
	return s.queries.DeleteUserActivities(ctx, userID)
}

func (s *activityService) DeleteActivityByEntity(ctx context.Context, entityType string, entityID int32) error {
	var entityTypeEnum db.EntityType
	switch entityType {
	case "MOVIE":
		entityTypeEnum = db.EntityTypeMOVIE
	case "REVIEW":
		entityTypeEnum = db.EntityTypeREVIEW
	case "COMMENT":
		entityTypeEnum = db.EntityTypeCOMMENT
	case "USER":
		entityTypeEnum = db.EntityTypeUSER
	default:
		entityTypeEnum = db.EntityTypeMOVIE
	}

	return s.queries.DeleteActivityByEntity(ctx, db.DeleteActivityByEntityParams{
		EntityType: entityTypeEnum,
		EntityID:   entityID,
	})
}

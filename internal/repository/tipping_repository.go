package repository

import (
	"context"
	"time"

	"vybes/internal/domain"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type TippingRepository struct {
	db *mongo.Database
}

func NewTippingRepository(db *mongo.Database) *TippingRepository {
	return &TippingRepository{db: db}
}

// Allowance methods
func (r *TippingRepository) CreateAllowance(ctx context.Context, allowance *domain.TippingAllowance) error {
	allowance.CreatedAt = time.Now()
	allowance.UpdatedAt = time.Now()
	
	_, err := r.db.Collection("tipping_allowances").InsertOne(ctx, allowance)
	return err
}

func (r *TippingRepository) GetAllowanceByUserID(ctx context.Context, userID primitive.ObjectID) (*domain.TippingAllowance, error) {
	var allowance domain.TippingAllowance
	err := r.db.Collection("tipping_allowances").FindOne(ctx, bson.M{"userId": userID}).Decode(&allowance)
	if err != nil {
		return nil, err
	}
	return &allowance, nil
}

func (r *TippingRepository) UpdateAllowance(ctx context.Context, allowance *domain.TippingAllowance) error {
	allowance.UpdatedAt = time.Now()
	
	_, err := r.db.Collection("tipping_allowances").UpdateOne(
		ctx,
		bson.M{"_id": allowance.ID},
		bson.M{"$set": allowance},
	)
	return err
}

func (r *TippingRepository) ResetWeeklyAllowances(ctx context.Context) error {
	now := time.Now()
	weekStart := getWeekStart(now)
	
	_, err := r.db.Collection("tipping_allowances").UpdateMany(
		ctx,
		bson.M{},
		bson.M{
			"$set": bson.M{
				"usedAmount": 0,
				"weekStart":  weekStart,
				"lastReset":  now,
				"updatedAt":  now,
			},
		},
	)
	return err
}

// Tip methods
func (r *TippingRepository) CreateTip(ctx context.Context, tip *domain.Tip) error {
	tip.CreatedAt = time.Now()
	
	_, err := r.db.Collection("tips").InsertOne(ctx, tip)
	return err
}

func (r *TippingRepository) GetTipByID(ctx context.Context, tipID primitive.ObjectID) (*domain.Tip, error) {
	var tip domain.Tip
	err := r.db.Collection("tips").FindOne(ctx, bson.M{"_id": tipID}).Decode(&tip)
	if err != nil {
		return nil, err
	}
	return &tip, nil
}

func (r *TippingRepository) UpdateTipStatus(ctx context.Context, tipID primitive.ObjectID, status domain.TipStatus) error {
	update := bson.M{
		"status": status,
	}
	
	if status == domain.TipStatusCompleted {
		now := time.Now()
		update["completedAt"] = now
	}
	
	_, err := r.db.Collection("tips").UpdateOne(
		ctx,
		bson.M{"_id": tipID},
		bson.M{"$set": update},
	)
	return err
}

func (r *TippingRepository) GetTipsByUser(ctx context.Context, userID primitive.ObjectID, limit int64) ([]*domain.Tip, error) {
	opts := options.Find().SetLimit(limit).SetSort(bson.M{"createdAt": -1})
	
	cursor, err := r.db.Collection("tips").Find(
		ctx,
		bson.M{
			"$or": []bson.M{
				{"fromUserId": userID},
				{"toUserId": userID},
			},
		},
		opts,
	)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	
	var tips []*domain.Tip
	if err = cursor.All(ctx, &tips); err != nil {
		return nil, err
	}
	
	return tips, nil
}

func (r *TippingRepository) GetTipsByContent(ctx context.Context, contentID primitive.ObjectID) ([]*domain.Tip, error) {
	cursor, err := r.db.Collection("tips").Find(
		ctx,
		bson.M{"contentId": contentID},
		options.Find().SetSort(bson.M{"createdAt": -1}),
	)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	
	var tips []*domain.Tip
	if err = cursor.All(ctx, &tips); err != nil {
		return nil, err
	}
	
	return tips, nil
}

// Stats methods
func (r *TippingRepository) GetTipStats(ctx context.Context, userID primitive.ObjectID) (*domain.TipStats, error) {
	var stats domain.TipStats
	err := r.db.Collection("tip_stats").FindOne(ctx, bson.M{"userId": userID}).Decode(&stats)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			// Create new stats if not exists
			stats = domain.TipStats{
				UserID:      userID,
				LastUpdated: time.Now(),
			}
			_, err = r.db.Collection("tip_stats").InsertOne(ctx, stats)
			if err != nil {
				return nil, err
			}
			return &stats, nil
		}
		return nil, err
	}
	return &stats, nil
}

func (r *TippingRepository) UpdateTipStats(ctx context.Context, userID primitive.ObjectID, received, sent int64) error {
	now := time.Now()
	
	_, err := r.db.Collection("tip_stats").UpdateOne(
		ctx,
		bson.M{"userId": userID},
		bson.M{
			"$inc": bson.M{
				"totalReceived": received,
				"totalSent":     sent,
			},
			"$set": bson.M{
				"lastUpdated": now,
			},
		},
		options.Update().SetUpsert(true),
	)
	return err
}

func (r *TippingRepository) ResetWeeklyStats(ctx context.Context) error {
	now := time.Now()
	
	_, err := r.db.Collection("tip_stats").UpdateMany(
		ctx,
		bson.M{},
		bson.M{
			"$set": bson.M{
				"weeklyReceived": 0,
				"weeklySent":     0,
				"lastUpdated":    now,
			},
		},
	)
	return err
}

// Helper function to get the start of the current week (Monday)
func getWeekStart(t time.Time) time.Time {
	weekday := t.Weekday()
	if weekday == time.Sunday {
		weekday = 7
	} else {
		weekday--
	}
	
	// Go back to Monday
	return t.AddDate(0, 0, -int(weekday)).Truncate(24 * time.Hour)
}
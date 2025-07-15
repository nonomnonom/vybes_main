package repository

import (
	"context"
	"vybes/internal/domain"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// BookmarkRepository defines the interface for bookmark data operations.
// Bookmarks allow users to save content for later viewing, providing
// a way to organize and access favorite posts and content.
type BookmarkRepository interface {
	// CreateBookmark saves a content item to a user's bookmarks
	CreateBookmark(ctx context.Context, bookmark *domain.Bookmark) error
	// DeleteBookmark removes a content item from a user's bookmarks
	DeleteBookmark(ctx context.Context, userID, postID primitive.ObjectID) error
	// GetUserBookmarks retrieves all bookmarked content for a specific user
	GetUserBookmarks(ctx context.Context, userID primitive.ObjectID, page, limit int) ([]domain.Bookmark, error)
	// IsBookmarked checks if a user has bookmarked a specific content item
	IsBookmarked(ctx context.Context, userID, postID primitive.ObjectID) (bool, error)
	// GetBookmarkCount returns the number of bookmarks for a content item
	GetBookmarkCount(ctx context.Context, postID primitive.ObjectID) (int64, error)
}

// mongoBookmarkRepository implements BookmarkRepository using MongoDB as the backend
type mongoBookmarkRepository struct {
	collection *mongo.Collection
}

// NewMongoBookmarkRepository creates a new bookmark repository instance with MongoDB backend.
// The repository handles all bookmark-related database operations including
// creating, deleting, and querying user bookmarks.
//
// Parameters:
//   - db: MongoDB database instance
//
// Returns:
//   - BookmarkRepository: A configured bookmark repository ready for use
func NewMongoBookmarkRepository(db *mongo.Database) BookmarkRepository {
	return &mongoBookmarkRepository{
		collection: db.Collection("bookmarks"),
	}
}

// CreateBookmark adds a content item to a user's bookmarks collection.
// Uses upsert to avoid duplicate bookmarks for the same content.
//
// Parameters:
//   - ctx: Context for the operation
//   - bookmark: The bookmark object to create
//
// Returns:
//   - error: Any error that occurred during the operation
func (r *mongoBookmarkRepository) CreateBookmark(ctx context.Context, bookmark *domain.Bookmark) error {
	// Use upsert to avoid duplicate bookmarks
	filter := bson.M{
		"userid":    bookmark.UserID,
		"postid": bookmark.PostID,
	}
	update := bson.M{"$setOnInsert": bookmark}
	opts := options.Update().SetUpsert(true)
	
	_, err := r.collection.UpdateOne(ctx, filter, update, opts)
	return err
}

// DeleteBookmark removes a content item from a user's bookmarks.
func (r *mongoBookmarkRepository) DeleteBookmark(ctx context.Context, userID, postID primitive.ObjectID) error {
	filter := bson.M{
		"userid":    userID,
		"postid": postID,
	}
	_, err := r.collection.DeleteOne(ctx, filter)
	return err
}

// GetUserBookmarks retrieves all bookmarked content for a specific user.
func (r *mongoBookmarkRepository) GetUserBookmarks(ctx context.Context, userID primitive.ObjectID, page, limit int) ([]domain.Bookmark, error) {
	var bookmarks []domain.Bookmark
	opts := options.Find().SetSkip(int64((page - 1) * limit)).SetLimit(int64(limit))
	cursor, err := r.collection.Find(ctx, bson.M{"userid": userID}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	err = cursor.All(ctx, &bookmarks)
	return bookmarks, err
}

// IsBookmarked checks if a user has bookmarked a specific content item.
func (r *mongoBookmarkRepository) IsBookmarked(ctx context.Context, userID, postID primitive.ObjectID) (bool, error) {
	filter := bson.M{
		"userid":    userID,
		"postid": postID,
	}
	count, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// GetBookmarkCount returns the number of bookmarks for a content item.
func (r *mongoBookmarkRepository) GetBookmarkCount(ctx context.Context, postID primitive.ObjectID) (int64, error) {
	return r.collection.CountDocuments(ctx, bson.M{"postid": postID})
}

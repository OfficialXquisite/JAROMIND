package services_impl

import (
    "context"
    "errors"
    "time"
	"fmt"
    "github.com/AbaraEmmanuel/jaromind-backend/database"
    "github.com/AbaraEmmanuel/jaromind-backend/models"
    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/bson/primitive"
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
)

// ReviewServiceImpl implements the ReviewService interface
type ReviewServiceImpl struct {
    reviewCollection *mongo.Collection
    courseCollection *mongo.Collection
}

// NewReviewServiceImpl creates a new review service implementation
func NewReviewServiceImpl() *ReviewServiceImpl {
    db := database.GetDB()
    return &ReviewServiceImpl{
        reviewCollection: db.Collection("reviews"),
        courseCollection: db.Collection("courses"),
    }
}

func (s *ReviewServiceImpl) CreateReview(ctx context.Context, review *models.Review) (*models.Review, error) {
    fmt.Println("\n=== REVIEW SERVICE - CreateReview ===")
    fmt.Printf("Review received: CourseID=%s, UserID=%v, Rating=%d\n", 
        review.CourseID, review.UserID, review.Rating)
    
    // Check if user already reviewed this course
    fmt.Println("Checking for existing review...")
    existingReview, err := s.GetReviewByUserAndCourse(ctx, review.UserID.Hex(), review.CourseID)
    if err != nil {
        fmt.Printf("Error checking existing review: %v\n", err)
        return nil, err
    }
    
    if existingReview != nil {
        fmt.Printf("Found existing review: %v - Updating instead\n", existingReview.ID)
        // Update existing review instead
        return s.UpdateReview(ctx, existingReview.ID.Hex(), review)
    }
    
    fmt.Println("No existing review found - creating new one")
    
    // Set timestamps
    review.CreatedAt = time.Now()
    review.UpdatedAt = time.Now()
    review.Date = time.Now()
    
    fmt.Printf("Inserting review into database...\n")
    
    // Insert review
    result, err := s.reviewCollection.InsertOne(ctx, review)
    if err != nil {
        fmt.Printf("Error inserting review: %v\n", err)
        return nil, err
    }
    
    review.ID = result.InsertedID.(primitive.ObjectID)
    fmt.Printf("Review inserted with ID: %v\n", review.ID)
    
    // Update course rating
    fmt.Println("Updating course rating...")
    if err := s.updateCourseRating(ctx, review.CourseID); err != nil {
        fmt.Printf("Warning: Failed to update course rating: %v\n", err)
        // Don't fail the review creation
    }
    
    fmt.Println("=== REVIEW SERVICE - Success ===")
    return review, nil
}

// GetReviewsByCourseID retrieves all reviews for a specific course
func (s *ReviewServiceImpl) GetReviewsByCourseID(ctx context.Context, courseID string) ([]models.Review, error) {
    // Try to convert as ObjectID first
    if objID, err := primitive.ObjectIDFromHex(courseID); err == nil {
        // It's an ObjectID
        filter := bson.M{"course_id": objID}
        opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}})
        
        cursor, err := s.reviewCollection.Find(ctx, filter, opts)
        if err != nil {
            return nil, err
        }
        defer cursor.Close(ctx)
        
        var reviews []models.Review
        if err := cursor.All(ctx, &reviews); err != nil {
            return nil, err
        }
        
        if reviews == nil {
            reviews = []models.Review{}
        }
        
        return reviews, nil
    }
    
    // If not ObjectID, try as string/UUID
    filter := bson.M{"course_id": courseID}
    opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}})
    
    cursor, err := s.reviewCollection.Find(ctx, filter, opts)
    if err != nil {
        return nil, err
    }
    defer cursor.Close(ctx)
    
    var reviews []models.Review
    if err := cursor.All(ctx, &reviews); err != nil {
        return nil, err
    }
    
    if reviews == nil {
        reviews = []models.Review{}
    }
    
    return reviews, nil
}

// GetReviewByID retrieves a single review by ID
func (s *ReviewServiceImpl) GetReviewByID(ctx context.Context, reviewID string) (*models.Review, error) {
    objectID, err := primitive.ObjectIDFromHex(reviewID)
    if err != nil {
        return nil, errors.New("invalid review ID")
    }

    var review models.Review
    filter := bson.M{"_id": objectID}

    err = s.reviewCollection.FindOne(ctx, filter).Decode(&review)
    if err != nil {
        if err == mongo.ErrNoDocuments {
            return nil, errors.New("review not found")
        }
        return nil, err
    }

    return &review, nil
}

// UpdateReview updates an existing review
func (s *ReviewServiceImpl) UpdateReview(ctx context.Context, reviewID string, review *models.Review) (*models.Review, error) {
    objectID, err := primitive.ObjectIDFromHex(reviewID)
    if err != nil {
        return nil, errors.New("invalid review ID")
    }

    review.UpdatedAt = time.Now()

    update := bson.M{
        "$set": bson.M{
            "rating":     review.Rating,
            "comment":    review.Comment,
            "updated_at": review.UpdatedAt,
        },
    }

    opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
    result := s.reviewCollection.FindOneAndUpdate(ctx, bson.M{"_id": objectID}, update, opts)

    var updatedReview models.Review
    if err := result.Decode(&updatedReview); err != nil {
        if err == mongo.ErrNoDocuments {
            return nil, errors.New("review not found")
        }
        return nil, err
    }

    // Update course rating
    if err := s.updateCourseRating(ctx, updatedReview.CourseID); err != nil {
        // Log error but don't fail the update
    }

    return &updatedReview, nil
}

// DeleteReview deletes a review
func (s *ReviewServiceImpl) DeleteReview(ctx context.Context, reviewID string) error {
    objectID, err := primitive.ObjectIDFromHex(reviewID)
    if err != nil {
        return errors.New("invalid review ID")
    }

    // Get review first to update course rating after deletion
    review, err := s.GetReviewByID(ctx, reviewID)
    if err != nil {
        return err
    }

    result, err := s.reviewCollection.DeleteOne(ctx, bson.M{"_id": objectID})
    if err != nil {
        return err
    }

    if result.DeletedCount == 0 {
        return errors.New("review not found")
    }

    // Update course rating
    if err := s.updateCourseRating(ctx, review.CourseID); err != nil {
        // Log error but don't fail the deletion
    }

    return nil
}

// GetReviewByUserAndCourse checks if user already reviewed a course
func (s *ReviewServiceImpl) GetReviewByUserAndCourse(ctx context.Context, userID, courseID string) (*models.Review, error) {
    userObjectID, err := primitive.ObjectIDFromHex(userID)
    if err != nil {
        return nil, errors.New("invalid user ID")
    }

    // Try to convert courseID to ObjectID first
    var filter bson.M
    if courseObjID, err := primitive.ObjectIDFromHex(courseID); err == nil {
        // CourseID is an ObjectID
        filter = bson.M{
            "user_id":   userObjectID,
            "course_id": courseObjID,
        }
    } else {
        // CourseID is a string/UUID
        filter = bson.M{
            "user_id":   userObjectID,
            "course_id": courseID,
        }
    }

    var review models.Review
    err = s.reviewCollection.FindOne(ctx, filter).Decode(&review)
    if err != nil {
        if err == mongo.ErrNoDocuments {
            return nil, nil // No existing review found (not an error)
        }
        return nil, err
    }

    return &review, nil
}

// CalculateCourseRating calculates average rating for a course
func (s *ReviewServiceImpl) CalculateCourseRating(ctx context.Context, courseID string) (float64, int, error) {
    // Try to convert as ObjectID first
    var filter bson.M
    if objID, err := primitive.ObjectIDFromHex(courseID); err == nil {
        // It's an ObjectID
        filter = bson.M{"course_id": objID}
    } else {
        // It's a string/UUID
        filter = bson.M{"course_id": courseID}
    }

    // Aggregate pipeline to calculate average rating
    pipeline := mongo.Pipeline{
        {{Key: "$match", Value: filter}},
        {{Key: "$group", Value: bson.M{
            "_id":         nil,
            "avgRating":   bson.M{"$avg": "$rating"},
            "totalReviews": bson.M{"$sum": 1},
        }}},
    }

    cursor, err := s.reviewCollection.Aggregate(ctx, pipeline)
    if err != nil {
        return 0, 0, err
    }
    defer cursor.Close(ctx)

    var result []struct {
        AvgRating    float64 `bson:"avgRating"`
        TotalReviews int     `bson:"totalReviews"`
    }

    if err := cursor.All(ctx, &result); err != nil {
        return 0, 0, err
    }

    if len(result) == 0 {
        return 0, 0, nil
    }

    return result[0].AvgRating, result[0].TotalReviews, nil
}

// updateCourseRating updates the course document with new rating
func (s *ReviewServiceImpl) updateCourseRating(ctx context.Context, courseID string) error {
    avgRating, totalReviews, err := s.CalculateCourseRating(ctx, courseID)
    if err != nil {
        return err
    }

    // Try to update by ID (string/UUID) first
    update := bson.M{
        "$set": bson.M{
            "rating":       avgRating,
            "review_count": totalReviews,
            "updated_at":   time.Now(),
        },
    }

    // Try to update with courseID as string first (for UUIDs)
    _, err = s.courseCollection.UpdateOne(ctx, bson.M{"id": courseID}, update)
    if err == nil {
        return nil
    }

    // If that failed, try to convert to ObjectID
    if objID, err2 := primitive.ObjectIDFromHex(courseID); err2 == nil {
        _, err = s.courseCollection.UpdateOne(ctx, bson.M{"_id": objID}, update)
        return err
    }

    return errors.New("could not update course rating - invalid course ID format")
}
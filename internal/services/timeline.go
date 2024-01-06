package services

import (
	"context"
	"fmt"
	"log"

	"github.com/erdemkosk/golang-twitter-timeline-service/internal/models"
	"github.com/erdemkosk/golang-twitter-timeline-service/internal/repositories"
	"github.com/erdemkosk/golang-twitter-timeline-service/pkg/database"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type TimelineService struct {
	timelineRepository repositories.MongoDBRepository
	userRepository     repositories.MongoDBRepository
	kafkaService       *KafkaService
}

func CreateTimelineService() *TimelineService {
	timelineRepository := repositories.NewMongoDBRepository(database.Client.Database("twitter"), "timeline", models.Timeline{})
	userRepository := repositories.NewMongoDBRepository(database.Client.Database("twitter"), "user", models.User{})
	kafkaService := CreateKafkaService()

	timelineService := &TimelineService{
		timelineRepository: *timelineRepository,
		userRepository:     *userRepository,
		kafkaService:       kafkaService,
	}

	go func() {
		for {
			select {
			case tweet := <-kafkaService.tweetChannel:
				err := timelineService.InsertTweetToTimeline(tweet)
				if err != nil {
					log.Printf("Error inserting tweet into timeline: %v\n", err)
				}
			}
		}
	}()

	return timelineService
}

func (this TimelineService) InsertTweetToTimeline(tweet models.Tweet) error {
	filterAll := bson.M{}

	resultAll, err := this.userRepository.Find(context.TODO(), filterAll)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(resultAll)

	objectID, err := primitive.ObjectIDFromHex(tweet.UserId)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(objectID)

	filter := bson.M{"_id": objectID}

	result, err := this.userRepository.FindOne(context.TODO(), filter)

	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(result)

	user, ok := result.(*models.User)
	if !ok {
		log.Fatal("Type Error")
	}

	fmt.Println(user)

	followers := user.Followers

	fmt.Println(followers)

	for _, followerID := range followers {
		filter := bson.M{"_id": followerID}
		update := bson.M{"$addToSet": bson.M{"tweets": tweet.ID}}

		err := this.timelineRepository.UpdateOneWithUpsert(context.TODO(), filter, update)
		if err != nil {
			return err
		}
	}

	return nil
}

func (this TimelineService) GetTimelineByUserId(ctx context.Context, userID string) ([]models.Tweet, error) {
	userObjectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, err
	}

	filter := bson.M{"userId": userObjectID}
	userTimelines, err := this.timelineRepository.Find(ctx, filter)
	if err != nil {
		return nil, err
	}

	var tweets []models.Tweet
	for _, userTimeline := range userTimelines {
		timeline, ok := userTimeline.(models.Timeline)
		if !ok {
			return nil, fmt.Errorf("Type assertion failed")
		}

		// Timeline'daki tweet'leri al
		tweetIDs := timeline.Tweets

		// Tweet'leri populat et
		populatedTweets, err := this.PopulateTweets(ctx, tweetIDs)
		if err != nil {
			return nil, err
		}

		// Append populatedTweets to tweets slice
		tweets = append(tweets, populatedTweets...)
	}

	return tweets, nil
}

func (this TimelineService) PopulateTweets(ctx context.Context, tweetIDs []primitive.ObjectID) ([]models.Tweet, error) {
	var tweets []models.Tweet
	for _, tweetID := range tweetIDs {
		filter := bson.M{"_id": tweetID}
		result, err := this.userRepository.FindOne(ctx, filter)
		if err != nil {
			return nil, err
		}

		tweet, ok := result.(models.Tweet)
		if !ok {
			return nil, fmt.Errorf("Type assertion failed")
		}

		tweets = append(tweets, tweet)
	}

	return tweets, nil
}

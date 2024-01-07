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
	tweetRepository    repositories.MongoDBRepository
	kafkaService       *KafkaService
}

func CreateTimelineService() *TimelineService {
	timelineRepository := repositories.NewMongoDBRepository(database.Client.Database("twitter"), "timeline", models.Timeline{})
	userRepository := repositories.NewMongoDBRepository(database.Client.Database("twitter"), "user", models.User{})
	tweetRepository := repositories.NewMongoDBRepository(database.Client.Database("twitter"), "tweet", models.Tweet{})
	kafkaService := CreateKafkaService()

	timelineService := &TimelineService{
		timelineRepository: *timelineRepository,
		userRepository:     *userRepository,
		tweetRepository:    *tweetRepository,
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
	objectID, err := primitive.ObjectIDFromHex(tweet.UserId)
	if err != nil {
		fmt.Println(err)
	}

	filter := bson.M{"_id": objectID}

	result, err := this.userRepository.FindOne(context.TODO(), filter)

	if err != nil {
		fmt.Println(err)
	}

	user, ok := result.(*models.User)
	if !ok {
		log.Fatal("Type Error")
	}

	followers := user.Followers

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

	filter := bson.M{"_id": userObjectID}
	userTimelines, err := this.timelineRepository.Find(ctx, filter)
	if err != nil {
		return nil, err
	}

	var tweets []models.Tweet
	for _, userTimeline := range userTimelines {
		var timeline models.Timeline
		timelineBytes, err := bson.Marshal(userTimeline)
		if err != nil {
			return nil, err
		}
		if err := bson.Unmarshal(timelineBytes, &timeline); err != nil {
			return nil, err
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

		result, err := this.tweetRepository.FindOne(ctx, filter)

		if err != nil {
			return nil, err
		}

		tweet, ok := result.(*models.Tweet)
		if !ok {
			log.Fatal("Type Error")
		}

		tweets = append(tweets, *tweet)
	}

	return tweets, nil
}

package mapping

import (
    "time"
    "github.com/google/uuid"
    "github.com/Greeshmanth1909/blogAggregator/internal/database"
)

type FeedResponse struct{
    ID uuid.UUID
    CreatedAt time.Time
    UpdatedAt time.Time
    Name string
    Url string
    UserID uuid.UUID
    LastFetchedAt *time.Time
}

func DatabaseFeedToResponseFeed(feed database.Feed) FeedResponse {
    feedRes := FeedResponse{feed.ID,
                feed.CreatedAt,
                feed.UpdatedAt,
                feed.Name,
                feed.Url,
                feed.UserID,
                &feed.LastFetchedAt.Time}
    return feedRes
}

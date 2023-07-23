package gmailext

import (
	"context"
	"fmt"

	"google.golang.org/api/gmail/v1"
)

const (
	limitPerRequest = 500
)

func Iter(
	ctx context.Context,
	service *gmail.Service,
	userID string,
	query string,
	process func(context.Context, *gmail.Message) (bool, error),
) error {
	messageIDs := []string{}
	if err := service.Users.Messages.List(userID).
		Q(query).
		MaxResults(limitPerRequest).
		Context(ctx).
		Pages(ctx, func(res *gmail.ListMessagesResponse) error {
			for _, messageSummary := range res.Messages {
				messageIDs = append(messageIDs, messageSummary.Id)
			}
			return nil
		}); err != nil {
		return fmt.Errorf("list gmail messages: %w", err)
	}

	for _, messageID := range messageIDs {
		message, err := Get(ctx, service, userID, messageID)
		if err != nil {
			return fmt.Errorf("get one gmail message: %w", err)
		}

		next, err := process(ctx, message)
		if err != nil {
			return err
		}
		if !next {
			break
		}
	}

	return nil
}

func Get(
	ctx context.Context,
	service *gmail.Service,
	userID string,
	messageID string,
) (*gmail.Message, error) {
	return service.Users.Messages.Get(userID, messageID).
		Format("full").
		Context(ctx).
		Do()
}

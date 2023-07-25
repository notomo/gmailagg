package gmailext

import (
	"context"
	"errors"
	"fmt"

	"golang.org/x/sync/errgroup"
	"google.golang.org/api/gmail/v1"
)

const (
	limitPerRequest = 500
)

var (
	errBreak = fmt.Errorf("break")
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

	eg, ctx := errgroup.WithContext(ctx)
	eg.SetLimit(4)
	for _, messageID := range messageIDs {
		messageID := messageID
		eg.Go(func() error {
			message, err := Get(ctx, service, userID, messageID)
			if err != nil {
				return fmt.Errorf("get one gmail message: %w", err)
			}

			next, err := process(ctx, message)
			if err != nil {
				return err
			}
			if !next {
				return errBreak
			}
			return nil
		})
	}
	if err := eg.Wait(); err != nil && !errors.Is(err, errBreak) {
		return err
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

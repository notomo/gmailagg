package gmailext

import (
	"context"
	"errors"
	"fmt"
	"iter"

	"golang.org/x/sync/errgroup"
	"google.golang.org/api/gmail/v1"
)

const (
	limitPerRequest = 500
)

type Got[T any] struct {
	Value T
	Err   error
}

var (
	errBreak = fmt.Errorf("break")
)

const userID = "me"

func getMessageIDs(
	ctx context.Context,
	service *gmail.Service,
	query string,
) iter.Seq[Got[string]] {
	return func(yield func(Got[string]) bool) {
		stopped := false
		if err := service.Users.Messages.List(userID).
			Q(query).
			MaxResults(limitPerRequest).
			Context(ctx).
			Pages(ctx, func(res *gmail.ListMessagesResponse) error {
				if stopped {
					return nil
				}
				for _, messageSummary := range res.Messages {
					got := Got[string]{Value: messageSummary.Id}
					if !yield(got) {
						stopped = true
						return nil
					}
				}
				return nil
			}); err != nil {
			got := Got[string]{Err: fmt.Errorf("list gmail messages: %w", err)}
			yield(got)
		}
	}
}

func Iter(
	ctx context.Context,
	service *gmail.Service,
	query string,
	process func(context.Context, *gmail.Message) (bool, error),
) error {
	eg, ctx := errgroup.WithContext(ctx)
	eg.SetLimit(4)
	for g := range getMessageIDs(ctx, service, query) {
		if g.Err != nil {
			return g.Err
		}

		eg.Go(func() error {
			msg, err := Get(ctx, service, userID, g.Value)
			if err != nil {
				return fmt.Errorf("get one gmail message: %w", err)
			}

			next, err := process(ctx, msg)
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

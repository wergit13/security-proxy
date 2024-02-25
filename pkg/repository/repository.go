package repository

import (
	"context"
	"sc-proxy/pkg/models"

	sq "github.com/Masterminds/squirrel"
)

var psql = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

type RequestResponce interface {
	Save(ctx context.Context, req []byte, resq []byte) (int, error)
	GetAllRequests(ctx context.Context) ([]models.Request, error)
	GetAllPairs(ctx context.Context) ([]models.RequestResponce, error)
	GetPairById(ctx context.Context, id int) (models.RequestResponce, error)
	GetRequestById(ctx context.Context, id int) (models.Request, error)
}

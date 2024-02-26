package repository

import (
	"context"
	"encoding/json"
	"strings"

	"sc-proxy/pkg/models"

	sq "github.com/Masterminds/squirrel"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db}
}

func (r *Repository) Save(ctx context.Context, req *[]byte, resp *[]byte) (int, error) {
	var id int

	query, args, err := psql.Insert(Table).
		Columns("request, responce").
		Values(req, resp).
		ToSql()

	if err != nil {
		return 0, err
	}

	query += " RETURNING id"
	row := r.db.QueryRow(ctx, query, args...)
	err = row.Scan(&id)
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (r *Repository) GetAllRequests(ctx context.Context) ([]models.Request, error) {
	query, args, err := psql.
		Select("id, request").
		From(Table).
		ToSql()

	if err != nil {
		return nil, err
	}

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var requests []models.Request
	for rows.Next() {
		var request models.Request
		request, err = scanRequest(rows)
		requests = append(requests, request)
	}
	if err != nil {
		return nil, err
	}
	if rows.Err() != nil {
		return nil, err
	}

	return requests, nil
}

func (r *Repository) GetAllPairs(ctx context.Context) ([]models.RequestResponce, error) {
	query, args, err := psql.
		Select("id, request, responce").
		From(Table).
		ToSql()

	if err != nil {
		return nil, err
	}

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var requests []models.RequestResponce
	for rows.Next() {
		var pair models.RequestResponce
		pair, err = scanPair(rows)
		requests = append(requests, pair)
	}
	if err != nil {
		return nil, err
	}
	if rows.Err() != nil {
		return nil, err
	}

	return requests, nil
}

func (r *Repository) GetPairById(ctx context.Context, id int) (models.RequestResponce, error) {
	query, args, err := psql.
		Select("id, request, responce").
		From(Table).
		Where(sq.Eq{"id": id}).
		ToSql()

	if err != nil {
		return models.RequestResponce{}, err
	}

	row := r.db.QueryRow(ctx, query, args...)

	return scanPair(row)
}

func (r *Repository) GetRequestById(ctx context.Context, id int) (models.Request, error) {
	query, args, err := psql.
		Select("id, request").
		From(Table).
		Where(sq.Eq{"id": id}).
		ToSql()

	if err != nil {
		return models.Request{}, err
	}

	row := r.db.QueryRow(ctx, query, args...)
	if err != nil {
		return models.Request{}, err
	}

	request, err := scanRequest(row)

	return request, err
}

func scanRequest(row pgx.Row) (models.Request, error) {
	var request models.Request
	var data string
	var id int
	if err := row.Scan(&id, &data); err != nil {
		return models.Request{}, err
	}
	data = strings.ReplaceAll(data, "\\", "")
	err := json.Unmarshal([]byte(data), &request)
	request.Id = id
	return request, err
}

func scanPair(row pgx.Row) (models.RequestResponce, error) {
	var request models.Request
	var responce *models.Responce = nil
	var dataReq string
	var dataResp *string
	var id int
	err := row.Scan(&id, &dataReq, &dataResp)
	if err != nil {
		return models.RequestResponce{}, err
	}

	dataReq = strings.ReplaceAll(dataReq, "\\", "")
	err = json.Unmarshal([]byte(dataReq), &request)
	if err != nil {
		return models.RequestResponce{}, err
	}

	if dataResp != nil {
		resnNoEscape := strings.ReplaceAll(*dataResp, "\\", "")
		err = json.Unmarshal([]byte(resnNoEscape), responce)
	}

	request.Id = id
	return models.RequestResponce{Request: request, Responce: responce}, err
}

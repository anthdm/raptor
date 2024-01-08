package storage

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/anthdm/raptor/pkg/types"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

type SQLStore struct {
	db *sql.DB
}

func NewSQLStore(user, password, dbname, host, port, sslmode string) (*SQLStore, error) {
	uri := fmt.Sprintf(
		"user=%s password=%s dbname=%s host=%s port=%s sslmode=%s",
		user,
		password,
		dbname,
		host,
		port,
		sslmode,
	)
	db, err := sql.Open("postgres", uri)
	if err != nil {
		return nil, err
	}
	_, err = db.Exec(createAllTablesQuery)
	if err != nil {
		return nil, err
	}

	return &SQLStore{
		db: db,
	}, nil
}

func (s *SQLStore) CreateEndpoint(endpoint *types.Endpoint) error {
	stmt := `
INSERT INTO endpoint (id, name, runtime, environment, created_at)
VALUES ($1, $2, $3, $4, $5)
RETURNING id`
	b, err := json.Marshal(endpoint.Environment)
	if err != nil {
		return err
	}
	_, err = s.db.Exec(stmt,
		endpoint.ID,
		endpoint.Name,
		endpoint.Runtime,
		b,
		endpoint.CreatedAT)
	return err
}

func (s *SQLStore) GetEndpoint(id uuid.UUID) (*types.Endpoint, error) {
	row := s.db.QueryRow("SELECT * FROM endpoint WHERE id = $1", id)
	var endpoint types.Endpoint
	err := scanEndpoint(row, &endpoint)
	return &endpoint, err
}

func (s *SQLStore) GetEndpoints() ([]types.Endpoint, error) {
	rows, err := s.db.Query("SELECT * FROM endpoint")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var endpoints []types.Endpoint
	for rows.Next() {
		var endpoint types.Endpoint
		if err := scanEndpoint(rows, &endpoint); err != nil {
			return nil, err
		}
		endpoints = append(endpoints, endpoint)
	}
	return endpoints, nil
}

func (s *SQLStore) UpdateEndpoint(id uuid.UUID, params UpdateEndpointParams) error {
	query, args := buildUpdateEndpointQuery(id, params)
	_, err := s.db.Exec(query, args...)
	return err
}

func (s *SQLStore) GetDeployment(id uuid.UUID) (*types.Deployment, error) {
	stmt := "SELECT id, endpoint_id, hash, blob, created_at FROM deployment WHERE id = $1"
	row := s.db.QueryRow(stmt, id)

	var deploy types.Deployment
	err := scanDeploy(row, &deploy)
	return &deploy, err
}

func (s *SQLStore) CreateDeployment(deploy *types.Deployment) error {
	stmt := `
INSERT INTO deployment (id, endpoint_id, hash, blob, created_at)
VALUES ($1, $2, $3, $4, $5)
RETURNING id`
	_, err := s.db.Exec(stmt,
		deploy.ID,
		deploy.EndpointID,
		deploy.Hash,
		deploy.Blob,
		deploy.CreatedAT)
	return err
}

func (s *SQLStore) CreateRuntimeMetric(metric *types.RuntimeMetric) error {
	return nil
}

func (s *SQLStore) GetRuntimeMetrics(id uuid.UUID) ([]types.RuntimeMetric, error) {
	return nil, nil
}

type Scanner interface {
	Scan(dest ...interface{}) error
}

func buildUpdateEndpointQuery(id uuid.UUID, params UpdateEndpointParams) (string, []any) {
	var (
		updates []string
		args    []any
		counter = 1
	)

	if params.ActiveDeployID.String() != "00000000-0000-0000-0000-000000000000" {
		updates = append(updates, fmt.Sprintf("active_deployment_id = $%d", counter))
		args = append(args, params.ActiveDeployID)
		counter++
	}
	if params.Environment != nil {
		b, err := json.Marshal(params.Environment)
		if err != nil {
			panic(err)
		}
		updates = append(updates, fmt.Sprintf("environment = $%d", counter))
		args = append(args, b)
		counter++
	}
	args = append(args, id)

	setClause := strings.Join(updates, ", ")
	query := fmt.Sprintf("UPDATE endpoint SET %s WHERE id = $%d", setClause, counter)

	return query, args
}

func scanDeploy(s Scanner, d *types.Deployment) error {
	return s.Scan(
		&d.ID,
		&d.EndpointID,
		&d.Hash,
		&d.Blob,
		&d.CreatedAT,
	)
}

func scanEndpoint(s Scanner, e *types.Endpoint) error {
	var envData []byte
	err := s.Scan(
		&e.ID,
		&e.Name,
		&e.Runtime,
		&envData,
		&e.CreatedAT,
		&e.ActiveDeploymentID,
	)
	if err != nil {
		return err
	}
	return json.Unmarshal(envData, &e.Environment)
}

var createAllTablesQuery = `
CREATE TABLE if not exists endpoint (
	id UUID primary key, 
	name text not null,
	runtime text not null,
	environment jsonb,
	created_at timestamp not null default now()
);

CREATE TABLE if not exists deployment (
	id UUID primary key, 
	endpoint_id UUID not null references endpoint,
	hash text not null,
	blob bytea not null,
	created_at timestamp not null default now()
);

ALTER table endpoint
ADD COLUMN if not exists active_deployment_id UUID references deployment
`

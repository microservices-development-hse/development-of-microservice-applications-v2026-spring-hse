package postgres

import (
	"context"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/microservices-development-hse/backend/internal/models"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestGetLatestSnapshot(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock: %s", err)
	}

	defer func() {
		if closeErr := db.Close(); closeErr != nil {
			t.Errorf("failed to close sqlmock: %v", closeErr)
		}
	}()

	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: db,
	}), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open gorm: %s", err)
	}

	repo := NewAnalyticsRepository(gormDB)

	t.Run("Success", func(t *testing.T) {
		projectID := 1
		reportType := "status_distribution"
		expectedData := `[{"name":"Done","value":10}]`

		rows := sqlmock.NewRows([]string{"id", "project_id", "type", "creation_time", "data"}).
			AddRow(1, projectID, reportType, time.Now(), []byte(expectedData))

		query := `SELECT * FROM analytics_snapshots WHERE project_id = $1 AND type = $2 ORDER BY creation_time DESC,analytics_snapshots."id" LIMIT $3`

		mock.ExpectQuery(regexp.QuoteMeta(query)).
			WithArgs(projectID, reportType, 1).
			WillReturnRows(rows)

		res, err := repo.GetLatestSnapshot(context.Background(), projectID, reportType)

		assert.NoError(t, err)

		if assert.NotNil(t, res) {
			assert.Equal(t, reportType, res.Type)
			assert.JSONEq(t, expectedData, string(res.Data))
		}
	})
}
func TestSaveSnapshot(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock: %s", err)
	}

	defer func() {
		if closeErr := db.Close(); closeErr != nil {
			t.Errorf("failed to close sqlmock: %v", closeErr)
		}
	}()

	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: db,
	}), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open gorm: %s", err)
	}

	repo := NewAnalyticsRepository(gormDB)

	t.Run("Success", func(t *testing.T) {
		snapshot := &models.AnalyticsSnapshot{
			ProjectID: 1,
			Type:      "complexity",
			Data:      []byte(`{"key": "val"}`),
		}

		mock.ExpectBegin()

		mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO analytics_snapshots ("project_id","type","data") VALUES ($1,$2,$3) RETURNING "creation_time","id"`)).
			WithArgs(snapshot.ProjectID, snapshot.Type, snapshot.Data).
			WillReturnRows(sqlmock.NewRows([]string{"creation_time", "id"}).
				AddRow(time.Now(), 1))

		mock.ExpectCommit()

		err := repo.SaveSnapshot(context.Background(), snapshot)

		assert.NoError(t, err)
		assert.Equal(t, uint(1), snapshot.ID)
	})
}

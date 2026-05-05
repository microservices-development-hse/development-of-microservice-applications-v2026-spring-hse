package postgres

import (
	"testing"

	"github.com/microservices-development-hse/backend/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestRepositories_Constructors(t *testing.T) {
	env := SetupRepoTestEnv(t)

	t.Run("NewRepositories_Check", func(t *testing.T) {
		repos := NewRepositories(env.DB)

		assert.NotNil(t, repos.Project)
		assert.NotNil(t, repos.Issue)
		assert.NotNil(t, repos.Analytics)
	})
}

func TestInitializeRepositories_Error(t *testing.T) {
	cfg := &config.Config{}
	repos, err := InitializeRepositories(cfg)

	assert.Error(t, err)
	assert.Nil(t, repos)
}

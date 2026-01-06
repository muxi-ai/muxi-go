//go:build integration
// +build integration

package muxi

import (
	"context"
	"os"
	"testing"
	"time"
)

// Required environment variables:
//   MUXI_E2E_SERVER_URL   - e.g. http://localhost:7890
//   MUXI_E2E_KEY_ID       - server HMAC key ID
//   MUXI_E2E_SECRET_KEY   - server HMAC secret
//   MUXI_E2E_FORMATION_ID - formation ID
//   MUXI_E2E_CLIENT_KEY   - formation client key
//   MUXI_E2E_ADMIN_KEY    - formation admin key (optional, used if set)

func requireEnv(t *testing.T, name string) string {
	v := os.Getenv(name)
	if v == "" {
		t.Skipf("missing %s (skip integration)", name)
	}
	return v
}

func TestServerSmoke(t *testing.T) {
	url := requireEnv(t, "MUXI_E2E_SERVER_URL")
	keyID := requireEnv(t, "MUXI_E2E_KEY_ID")
	secret := requireEnv(t, "MUXI_E2E_SECRET_KEY")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	server := NewServerClient(&ServerConfig{
		URL:        url,
		KeyID:      keyID,
		SecretKey:  secret,
		MaxRetries: 1,
	})

	if _, err := server.Ping(ctx); err != nil {
		t.Fatalf("ping: %v", err)
	}

	if h, err := server.Health(ctx); err != nil {
		t.Fatalf("health: %v", err)
	} else if h.Data.Formations < 0 {
		t.Fatalf("health formations invalid: %d", h.Data.Formations)
	}

	if st, err := server.Status(ctx); err != nil {
		t.Fatalf("status: %v", err)
	} else if st.Formations.Total < 0 {
		t.Fatalf("status formations invalid: %d", st.Formations.Total)
	}

	if lf, err := server.ListFormations(ctx); err != nil {
		t.Fatalf("list formations: %v", err)
	} else if lf.Total != 0 && len(lf.Formations) == 0 {
		t.Fatalf("list formations mismatch: total=%d len=%d", lf.Total, len(lf.Formations))
	}

	if _, err := server.GetServerLogs(ctx, 10); err != nil {
		t.Fatalf("server logs: %v", err)
	}
}

func TestFormationSmoke(t *testing.T) {
	base := requireEnv(t, "MUXI_E2E_SERVER_URL")
	formationID := requireEnv(t, "MUXI_E2E_FORMATION_ID")
	clientKey := requireEnv(t, "MUXI_E2E_CLIENT_KEY")
	adminKey := os.Getenv("MUXI_E2E_ADMIN_KEY")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	formation := NewFormationClient(&FormationConfig{
		FormationID: formationID,
		ServerURL:   base,
		ClientKey:   clientKey,
		AdminKey:    adminKey,
		MaxRetries:  1,
	})

	if h, err := formation.Health(ctx); err != nil {
		t.Fatalf("health: %v", err)
	} else if h.Status == "" {
		t.Fatalf("health status empty")
	}

	if st, err := formation.GetStatus(ctx); err != nil {
		t.Fatalf("status: %v", err)
	} else if st.Formation.ID == "" {
		t.Fatalf("formation id empty")
	}

	if cfg, err := formation.GetConfig(ctx); err != nil {
		t.Fatalf("config: %v", err)
	} else if cfg.SchemaVersion == "" {
		t.Fatalf("config schema_version empty")
	}

	if agents, err := formation.GetAgents(ctx); err != nil {
		t.Fatalf("agents: %v", err)
	} else if agents.Count < 0 {
		t.Fatalf("agents count invalid: %d", agents.Count)
	}
}

package godo

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

var privateNetworkUUID = "880b7f98-f062-404d-b33c-458d545696f6"

var db = Database{
	ID:          "da4e0206-d019-41d7-b51f-deadbeefbb8f",
	Name:        "dbtest",
	EngineSlug:  "pg",
	VersionSlug: "11",
	Connection: &DatabaseConnection{
		URI:      "postgres://doadmin:zt91mum075ofzyww@dbtest-do-user-3342561-0.db.ondigitalocean.com:25060/defaultdb?sslmode=require",
		Database: "defaultdb",
		Host:     "dbtest-do-user-3342561-0.db.ondigitalocean.com",
		Port:     25060,
		User:     "doadmin",
		Password: "zt91mum075ofzyww",
		SSL:      true,
	},
	PrivateConnection: &DatabaseConnection{
		URI:      "postgres://doadmin:zt91mum075ofzyww@private-dbtest-do-user-3342561-0.db.ondigitalocean.com:25060/defaultdb?sslmode=require",
		Database: "defaultdb",
		Host:     "dbtest-do-user-3342561-0.db.ondigitalocean.com",
		Port:     25060,
		User:     "doadmin",
		Password: "zt91mum075ofzyww",
		SSL:      true,
	},
	Users: []DatabaseUser{
		DatabaseUser{
			Name:     "doadmin",
			Role:     "primary",
			Password: "zt91mum075ofzyww",
		},
	},
	DBNames: []string{
		"defaultdb",
	},
	NumNodes:   3,
	RegionSlug: "sfo2",
	Status:     "online",
	CreatedAt:  time.Date(2019, 2, 26, 6, 12, 39, 0, time.UTC),
	MaintenanceWindow: &DatabaseMaintenanceWindow{
		Day:         "monday",
		Hour:        "13:51:14",
		Pending:     false,
		Description: nil,
	},
	SizeSlug:           "db-s-2vcpu-4gb",
	PrivateNetworkUUID: "da4e0206-d019-41d7-b51f-deadbeefbb8f",
}

var dbJSON = `
{
	"id": "da4e0206-d019-41d7-b51f-deadbeefbb8f",
	"name": "dbtest",
	"engine": "pg",
	"version": "11",
	"connection": {
		"uri": "postgres://doadmin:zt91mum075ofzyww@dbtest-do-user-3342561-0.db.ondigitalocean.com:25060/defaultdb?sslmode=require",
		"database": "defaultdb",
		"host": "dbtest-do-user-3342561-0.db.ondigitalocean.com",
		"port": 25060,
		"user": "doadmin",
		"password": "zt91mum075ofzyww",
		"ssl": true
	},
	"private_connection": {
		"uri": "postgres://doadmin:zt91mum075ofzyww@private-dbtest-do-user-3342561-0.db.ondigitalocean.com:25060/defaultdb?sslmode=require",
		"database": "defaultdb",
		"host": "dbtest-do-user-3342561-0.db.ondigitalocean.com",
		"port": 25060,
		"user": "doadmin",
		"password": "zt91mum075ofzyww",
		"ssl": true
	},
	"users": [
		{
			"name": "doadmin",
			"role": "primary",
			"password": "zt91mum075ofzyww"
		}
	],
	"db_names": [
		"defaultdb"
	],
	"num_nodes": 3,
	"region": "sfo2",
	"status": "online",
	"created_at": "2019-02-26T06:12:39Z",
	"maintenance_window": {
		"day": "monday",
		"hour": "13:51:14",
		"pending": false,
		"description": null
	},
	"size": "db-s-2vcpu-4gb",
	"private_network_uuid": "da4e0206-d019-41d7-b51f-deadbeefbb8f"
}
`

var dbsJSON = fmt.Sprintf(`
{
  "databases": [
	%s
  ]
}
`, dbJSON)

func TestDatabases_List(t *testing.T) {
	setup()
	defer teardown()

	dbSvc := client.Databases

	want := []Database{db}

	mux.HandleFunc("/v2/databases", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, dbsJSON)
	})

	got, _, err := dbSvc.List(ctx, nil)
	require.NoError(t, err)
	require.Equal(t, want, got)
}

func TestDatabases_Get(t *testing.T) {
	setup()
	defer teardown()

	dbID := "da4e0206-d019-41d7-b51f-deadbeefbb8f"

	body := fmt.Sprintf(`
{
  "database": %s
}
`, dbJSON)

	path := fmt.Sprintf("/v2/databases/%s", dbID)

	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, body)
	})

	got, _, err := client.Databases.Get(ctx, dbID)
	require.NoError(t, err)
	require.Equal(t, &db, got)
}

func TestDatabases_Create(t *testing.T) {
	setup()
	defer teardown()

	want := &Database{
		ID:          "8d91899c-0739-4a1a-acc5-deadbeefbb8f",
		Name:        "backend-test",
		EngineSlug:  "pg",
		VersionSlug: "10",
		Connection: &DatabaseConnection{
			URI:      "postgres://doadmin:zt91mum075ofzyww@dbtest-do-user-3342561-0.db.ondigitalocean.com:25060/defaultdb?sslmode=require",
			Database: "defaultdb",
			Host:     "dbtest-do-user-3342561-0.db.ondigitalocean.com",
			Port:     25060,
			User:     "doadmin",
			Password: "zt91mum075ofzyww",
			SSL:      true,
		},
		PrivateConnection: &DatabaseConnection{
			URI:      "postgres://doadmin:zt91mum075ofzyww@private-dbtest-do-user-3342561-0.db.ondigitalocean.com:25060/defaultdb?sslmode=require",
			Database: "defaultdb",
			Host:     "dbtest-do-user-3342561-0.db.ondigitalocean.com",
			Port:     25060,
			User:     "doadmin",
			Password: "zt91mum075ofzyww",
			SSL:      true,
		},
		Users:             nil,
		DBNames:           nil,
		NumNodes:          2,
		RegionSlug:        "nyc3",
		Status:            "creating",
		CreatedAt:         time.Date(2019, 2, 26, 6, 12, 39, 0, time.UTC),
		MaintenanceWindow: nil,
		SizeSlug:          "db-s-2vcpu-4gb",
	}

	createRequest := &DatabaseCreateRequest{
		Name:       "backend-test",
		EngineSlug: "pg",
		Version:    "10",
		Region:     "nyc3",
		SizeSlug:   "db-s-2vcpu-4gb",
		NumNodes:   2,
	}

	body := `
{
	"database": {
		"id": "8d91899c-0739-4a1a-acc5-deadbeefbb8f",
		"name": "backend-test",
		"engine": "pg",
		"version": "10",
		"connection": {
			"uri": "postgres://doadmin:zt91mum075ofzyww@dbtest-do-user-3342561-0.db.ondigitalocean.com:25060/defaultdb?sslmode=require",
			"database": "defaultdb",
			"host": "dbtest-do-user-3342561-0.db.ondigitalocean.com",
			"port": 25060,
			"user": "doadmin",
			"password": "zt91mum075ofzyww",
			"ssl": true
		},
		"private_connection": {
			"uri": "postgres://doadmin:zt91mum075ofzyww@private-dbtest-do-user-3342561-0.db.ondigitalocean.com:25060/defaultdb?sslmode=require",
			"database": "defaultdb",
			"host": "dbtest-do-user-3342561-0.db.ondigitalocean.com",
			"port": 25060,
			"user": "doadmin",
			"password": "zt91mum075ofzyww",
			"ssl": true
		},
		"users": null,
		"db_names": null,
		"num_nodes": 2,
		"region": "nyc3",
		"status": "creating",
		"created_at": "2019-02-26T06:12:39Z",
		"maintenance_window": null,
		"size": "db-s-2vcpu-4gb"
	}
}`

	mux.HandleFunc("/v2/databases", func(w http.ResponseWriter, r *http.Request) {
		v := new(DatabaseCreateRequest)
		err := json.NewDecoder(r.Body).Decode(v)
		if err != nil {
			t.Fatal(err)
		}

		testMethod(t, r, http.MethodPost)
		require.Equal(t, v, createRequest)
		fmt.Fprint(w, body)
	})

	got, _, err := client.Databases.Create(ctx, createRequest)
	require.NoError(t, err)
	require.Equal(t, want, got)
}

func TestDatabases_Delete(t *testing.T) {
	setup()
	defer teardown()

	dbID := "deadbeef-dead-4aa5-beef-deadbeef347d"

	path := fmt.Sprintf("/v2/databases/%s", dbID)

	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
	})

	_, err := client.Databases.Delete(ctx, "deadbeef-dead-4aa5-beef-deadbeef347d")
	require.NoError(t, err)
}

func TestDatabases_Resize(t *testing.T) {
	setup()
	defer teardown()

	resizeRequest := &DatabaseResizeRequest{
		SizeSlug: "db-s-16vcpu-64gb",
		NumNodes: 3,
	}

	dbID := "deadbeef-dead-4aa5-beef-deadbeef347d"

	path := fmt.Sprintf("/v2/databases/%s/resize", dbID)

	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPut)
	})

	_, err := client.Databases.Resize(ctx, "deadbeef-dead-4aa5-beef-deadbeef347d", resizeRequest)
	require.NoError(t, err)
}

func TestDatabases_Migrate(t *testing.T) {
	setup()
	defer teardown()

	migrateRequest := &DatabaseMigrateRequest{
		Region: "lon1",
	}

	dbID := "deadbeef-dead-4aa5-beef-deadbeef347d"

	path := fmt.Sprintf("/v2/databases/%s/migrate", dbID)

	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPut)
	})

	_, err := client.Databases.Migrate(ctx, "deadbeef-dead-4aa5-beef-deadbeef347d", migrateRequest)
	require.NoError(t, err)
}

func TestDatabases_Migrate_WithPrivateNet(t *testing.T) {
	setup()
	defer teardown()

	migrateRequest := &DatabaseMigrateRequest{
		Region:             "lon1",
		PrivateNetworkUUID: privateNetworkUUID,
	}

	dbID := "deadbeef-dead-4aa5-beef-deadbeef347d"

	path := fmt.Sprintf("/v2/databases/%s/migrate", dbID)

	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPut)
	})

	_, err := client.Databases.Migrate(ctx, "deadbeef-dead-4aa5-beef-deadbeef347d", migrateRequest)
	require.NoError(t, err)
}

func TestDatabases_UpdateMaintenance(t *testing.T) {
	setup()
	defer teardown()

	maintenanceRequest := &DatabaseUpdateMaintenanceRequest{
		Day:  "thursday",
		Hour: "16:00",
	}

	dbID := "deadbeef-dead-4aa5-beef-deadbeef347d"

	path := fmt.Sprintf("/v2/databases/%s/maintenance", dbID)

	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPut)
	})

	_, err := client.Databases.UpdateMaintenance(ctx, "deadbeef-dead-4aa5-beef-deadbeef347d", maintenanceRequest)
	require.NoError(t, err)
}

func TestDatabases_ListBackups(t *testing.T) {
	setup()
	defer teardown()

	want := []DatabaseBackup{
		DatabaseBackup{
			CreatedAt:     time.Date(2019, 1, 11, 18, 42, 27, 0, time.UTC),
			SizeGigabytes: 0.03357696,
		},
		DatabaseBackup{
			CreatedAt:     time.Date(2019, 1, 12, 18, 42, 29, 0, time.UTC),
			SizeGigabytes: 0.03364864,
		},
	}

	body := `
{
  "backups": [
    {
      "created_at": "2019-01-11T18:42:27Z",
      "size_gigabytes": 0.03357696
    },
    {
      "created_at": "2019-01-12T18:42:29Z",
      "size_gigabytes": 0.03364864
    }
  ]
}
`

	dbID := "deadbeef-dead-4aa5-beef-deadbeef347d"

	path := fmt.Sprintf("/v2/databases/%s/backups", dbID)

	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, body)
	})

	got, _, err := client.Databases.ListBackups(ctx, "deadbeef-dead-4aa5-beef-deadbeef347d", nil)
	require.NoError(t, err)
	require.Equal(t, want, got)
}

func TestDatabases_GetUser(t *testing.T) {
	setup()
	defer teardown()

	want := &DatabaseUser{
		Name:     "name",
		Role:     "foo",
		Password: "pass",
	}

	body := `
{
  "user": {
    "name": "name",
    "role": "foo",
    "password": "pass"
  }
}
`

	dbID := "deadbeef-dead-4aa5-beef-deadbeef347d"

	path := fmt.Sprintf("/v2/databases/%s/users/name", dbID)

	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, body)
	})

	got, _, err := client.Databases.GetUser(ctx, dbID, "name")
	require.NoError(t, err)
	require.Equal(t, want, got)
}

func TestDatabases_ListUsers(t *testing.T) {
	setup()
	defer teardown()

	dbID := "deadbeef-dead-4aa5-beef-deadbeef347d"

	want := []DatabaseUser{
		{
			Name:     "name",
			Role:     "foo",
			Password: "pass",
		},
		{
			Name:     "bar",
			Role:     "foo",
			Password: "pass",
		},
	}

	body := `
{
  "users": [{
    "name": "name",
    "role": "foo",
    "password": "pass"
  },
  {
    "name": "bar",
    "role": "foo",
    "password": "pass"
  }]
}
`
	path := fmt.Sprintf("/v2/databases/%s/users", dbID)

	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, body)
	})

	got, _, err := client.Databases.ListUsers(ctx, dbID, nil)
	require.NoError(t, err)
	require.Equal(t, want, got)
}

func TestDatabases_CreateUser(t *testing.T) {
	setup()
	defer teardown()

	dbID := "deadbeef-dead-4aa5-beef-deadbeef347d"

	want := &DatabaseUser{
		Name:     "name",
		Role:     "foo",
		Password: "pass",
	}

	body := `
{
  "user": {
    "name": "name",
    "role": "foo",
    "password": "pass"
  }
}
`
	path := fmt.Sprintf("/v2/databases/%s/users", dbID)

	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		fmt.Fprint(w, body)
	})

	got, _, err := client.Databases.CreateUser(ctx, dbID, &DatabaseCreateUserRequest{Name: "user"})
	require.NoError(t, err)
	require.Equal(t, want, got)
}

func TestDatabases_DeleteUser(t *testing.T) {
	setup()
	defer teardown()

	dbID := "deadbeef-dead-4aa5-beef-deadbeef347d"

	path := fmt.Sprintf("/v2/databases/%s/users/user", dbID)

	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
	})

	_, err := client.Databases.DeleteUser(ctx, dbID, "user")
	require.NoError(t, err)
}

func TestDatabases_ListDBs(t *testing.T) {
	setup()
	defer teardown()

	dbID := "deadbeef-dead-4aa5-beef-deadbeef347d"

	want := []DatabaseDB{
		{Name: "foo"},
		{Name: "bar"},
	}

	body := `
{
  "dbs": [{
    "name": "foo"
  },
  {
    "name": "bar"
  }]
}
`
	path := fmt.Sprintf("/v2/databases/%s/dbs", dbID)

	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, body)
	})

	got, _, err := client.Databases.ListDBs(ctx, dbID, nil)
	require.NoError(t, err)
	require.Equal(t, want, got)
}

func TestDatabases_CreateDB(t *testing.T) {
	setup()
	defer teardown()

	dbID := "deadbeef-dead-4aa5-beef-deadbeef347d"

	want := &DatabaseDB{
		Name: "foo",
	}

	body := `
{
  "db": {
    "name": "foo"
  }
}
`
	path := fmt.Sprintf("/v2/databases/%s/dbs", dbID)

	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		fmt.Fprint(w, body)
	})

	got, _, err := client.Databases.CreateDB(ctx, dbID, &DatabaseCreateDBRequest{Name: "foo"})
	require.NoError(t, err)
	require.Equal(t, want, got)
}

func TestDatabases_GetDB(t *testing.T) {
	setup()
	defer teardown()

	dbID := "deadbeef-dead-4aa5-beef-deadbeef347d"

	want := &DatabaseDB{
		Name: "foo",
	}

	body := `
{
  "db": {
    "name": "foo"
  }
}
`
	path := fmt.Sprintf("/v2/databases/%s/dbs/foo", dbID)

	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, body)
	})

	got, _, err := client.Databases.GetDB(ctx, dbID, "foo")
	require.NoError(t, err)
	require.Equal(t, want, got)
}

func TestDatabases_DeleteDB(t *testing.T) {
	setup()
	defer teardown()

	dbID := "deadbeef-dead-4aa5-beef-deadbeef347d"

	body := `
{
  "db": {
    "name": "foo"
  }
}
`
	path := fmt.Sprintf("/v2/databases/%s/dbs/foo", dbID)

	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
		fmt.Fprint(w, body)
	})

	_, err := client.Databases.DeleteDB(ctx, dbID, "foo")
	require.NoError(t, err)
}

func TestDatabases_ListPools(t *testing.T) {
	setup()
	defer teardown()

	dbID := "deadbeef-dead-4aa5-beef-deadbeef347d"

	want := []DatabasePool{
		{
			Name:     "pool",
			User:     "user",
			Size:     10,
			Mode:     "transaction",
			Database: "db",
			Connection: &DatabaseConnection{
				URI:      "postgresql://user:pass@host.com/db",
				Host:     "host.com",
				Port:     1234,
				User:     "user",
				Password: "pass",
				SSL:      true,
				Database: "db",
			},
			PrivateConnection: &DatabaseConnection{
				URI:      "postgresql://user:pass@private-host.com/db",
				Host:     "private-host.com",
				Port:     1234,
				User:     "user",
				Password: "pass",
				SSL:      true,
				Database: "db",
			},
		},
	}

	body := `
{
  "pools": [{
    "name": "pool",
    "user": "user",
    "size": 10,
    "mode": "transaction",
    "db": "db",
    "connection": {
      "uri": "postgresql://user:pass@host.com/db",
      "host": "host.com",
      "port": 1234,
      "user": "user",
      "password": "pass",
      "database": "db",
      "ssl": true
    },
    "private_connection": {
      "uri": "postgresql://user:pass@private-host.com/db",
      "host": "private-host.com",
      "port": 1234,
      "user": "user",
      "password": "pass",
      "database": "db",
      "ssl": true
    }
  }]
}
`
	path := fmt.Sprintf("/v2/databases/%s/pools", dbID)

	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, body)
	})

	got, _, err := client.Databases.ListPools(ctx, dbID, nil)
	require.NoError(t, err)
	require.Equal(t, want, got)
}

func TestDatabases_CreatePool(t *testing.T) {
	setup()
	defer teardown()

	dbID := "deadbeef-dead-4aa5-beef-deadbeef347d"

	want := &DatabasePool{
		Name:     "pool",
		User:     "user",
		Size:     10,
		Mode:     "transaction",
		Database: "db",
		Connection: &DatabaseConnection{
			URI:      "postgresql://user:pass@host.com/db",
			Host:     "host.com",
			Port:     1234,
			User:     "user",
			Password: "pass",
			SSL:      true,
			Database: "db",
		},
		PrivateConnection: &DatabaseConnection{
			URI:      "postgresql://user:pass@private-host.com/db",
			Host:     "private-host.com",
			Port:     1234,
			User:     "user",
			Password: "pass",
			SSL:      true,
			Database: "db",
		},
	}

	body := `
{
  "pool": {
    "name": "pool",
    "user": "user",
    "size": 10,
    "mode": "transaction",
    "db": "db",
    "connection": {
      "uri": "postgresql://user:pass@host.com/db",
      "host": "host.com",
      "port": 1234,
      "user": "user",
      "password": "pass",
      "database": "db",
      "ssl": true
    },
    "private_connection": {
      "uri": "postgresql://user:pass@private-host.com/db",
      "host": "private-host.com",
      "port": 1234,
      "user": "user",
      "password": "pass",
      "database": "db",
      "ssl": true
    }
  }
}
`
	path := fmt.Sprintf("/v2/databases/%s/pools", dbID)

	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		fmt.Fprint(w, body)
	})

	got, _, err := client.Databases.CreatePool(ctx, dbID, &DatabaseCreatePoolRequest{
		Name:     "pool",
		Database: "db",
		Size:     10,
		User:     "foo",
		Mode:     "transaction",
	})
	require.NoError(t, err)
	require.Equal(t, want, got)
}

func TestDatabases_GetPool(t *testing.T) {
	setup()
	defer teardown()

	dbID := "deadbeef-dead-4aa5-beef-deadbeef347d"

	want := &DatabasePool{
		Name:     "pool",
		User:     "user",
		Size:     10,
		Mode:     "transaction",
		Database: "db",
		Connection: &DatabaseConnection{
			URI:      "postgresql://user:pass@host.com/db",
			Host:     "host.com",
			Port:     1234,
			User:     "user",
			Password: "pass",
			SSL:      true,
			Database: "db",
		},
		PrivateConnection: &DatabaseConnection{
			URI:      "postgresql://user:pass@private-host.com/db",
			Host:     "private-host.com",
			Port:     1234,
			User:     "user",
			Password: "pass",
			SSL:      true,
			Database: "db",
		},
	}

	body := `
{
  "pool": {
    "name": "pool",
    "user": "user",
    "size": 10,
    "mode": "transaction",
    "db": "db",
    "connection": {
      "uri": "postgresql://user:pass@host.com/db",
      "host": "host.com",
      "port": 1234,
      "user": "user",
      "password": "pass",
      "database": "db",
      "ssl": true
    },
    "private_connection": {
      "uri": "postgresql://user:pass@private-host.com/db",
      "host": "private-host.com",
      "port": 1234,
      "user": "user",
      "password": "pass",
      "database": "db",
      "ssl": true
    }
  }
}
`
	path := fmt.Sprintf("/v2/databases/%s/pools/pool", dbID)

	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, body)
	})

	got, _, err := client.Databases.GetPool(ctx, dbID, "pool")
	require.NoError(t, err)
	require.Equal(t, want, got)
}

func TestDatabases_DeletePool(t *testing.T) {
	setup()
	defer teardown()

	dbID := "deadbeef-dead-4aa5-beef-deadbeef347d"

	path := fmt.Sprintf("/v2/databases/%s/pools/pool", dbID)

	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
	})

	_, err := client.Databases.DeletePool(ctx, dbID, "pool")
	require.NoError(t, err)
}

func TestDatabases_GetReplica(t *testing.T) {
	setup()
	defer teardown()

	dbID := "deadbeef-dead-4aa5-beef-deadbeef347d"

	createdAt := time.Date(2019, 01, 01, 0, 0, 0, 0, time.UTC)

	want := &DatabaseReplica{
		Name:      "pool",
		Region:    "nyc1",
		Status:    "online",
		CreatedAt: createdAt,
		Connection: &DatabaseConnection{
			URI:      "postgresql://user:pass@host.com/db",
			Host:     "host.com",
			Port:     1234,
			User:     "user",
			Password: "pass",
			SSL:      true,
			Database: "db",
		},
		PrivateConnection: &DatabaseConnection{
			URI:      "postgresql://user:pass@private-host.com/db",
			Host:     "private-host.com",
			Port:     1234,
			User:     "user",
			Password: "pass",
			SSL:      true,
			Database: "db",
		},
		PrivateNetworkUUID: "deadbeef-dead-4aa5-beef-deadbeef347d",
	}

	body := `
{
  "replica": {
    "name": "pool",
    "region": "nyc1",
    "status": "online",
    "created_at": "` + createdAt.Format(time.RFC3339) + `",
    "connection": {
      "uri": "postgresql://user:pass@host.com/db",
      "host": "host.com",
      "port": 1234,
      "user": "user",
      "password": "pass",
      "database": "db",
      "ssl": true
    },
    "private_connection": {
      "uri": "postgresql://user:pass@private-host.com/db",
      "host": "private-host.com",
      "port": 1234,
      "user": "user",
      "password": "pass",
      "database": "db",
      "ssl": true
    },
    "private_network_uuid": "deadbeef-dead-4aa5-beef-deadbeef347d"
  }
}
`
	path := fmt.Sprintf("/v2/databases/%s/replicas/replica", dbID)

	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, body)
	})

	got, _, err := client.Databases.GetReplica(ctx, dbID, "replica")
	require.NoError(t, err)
	require.Equal(t, want, got)
}

func TestDatabases_ListReplicas(t *testing.T) {
	setup()
	defer teardown()

	dbID := "deadbeef-dead-4aa5-beef-deadbeef347d"

	createdAt := time.Date(2019, 01, 01, 0, 0, 0, 0, time.UTC)

	want := []DatabaseReplica{
		{
			Name:      "pool",
			Region:    "nyc1",
			Status:    "online",
			CreatedAt: createdAt,
			Connection: &DatabaseConnection{
				URI:      "postgresql://user:pass@host.com/db",
				Host:     "host.com",
				Port:     1234,
				User:     "user",
				Password: "pass",
				SSL:      true,
				Database: "db",
			},
			PrivateConnection: &DatabaseConnection{
				URI:      "postgresql://user:pass@private-host.com/db",
				Host:     "private-host.com",
				Port:     1234,
				User:     "user",
				Password: "pass",
				SSL:      true,
				Database: "db",
			},
			PrivateNetworkUUID: "deadbeef-dead-4aa5-beef-deadbeef347d",
		},
	}

	body := `
{
  "replicas": [{
    "name": "pool",
    "region": "nyc1",
    "status": "online",
    "created_at": "` + createdAt.Format(time.RFC3339) + `",
    "connection": {
      "uri": "postgresql://user:pass@host.com/db",
      "host": "host.com",
      "port": 1234,
      "user": "user",
      "password": "pass",
      "database": "db",
      "ssl": true
    },
    "private_connection": {
      "uri": "postgresql://user:pass@private-host.com/db",
      "host": "private-host.com",
      "port": 1234,
      "user": "user",
      "password": "pass",
      "database": "db",
      "ssl": true
    },
    "private_network_uuid": "deadbeef-dead-4aa5-beef-deadbeef347d"
  }]
}
`
	path := fmt.Sprintf("/v2/databases/%s/replicas", dbID)

	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, body)
	})

	got, _, err := client.Databases.ListReplicas(ctx, dbID, nil)
	require.NoError(t, err)
	require.Equal(t, want, got)
}

func TestDatabases_CreateReplica(t *testing.T) {
	setup()
	defer teardown()

	dbID := "deadbeef-dead-4aa5-beef-deadbeef347d"

	createdAt := time.Date(2019, 01, 01, 0, 0, 0, 0, time.UTC)

	want := &DatabaseReplica{
		Name:      "pool",
		Region:    "nyc1",
		Status:    "online",
		CreatedAt: createdAt,
		Connection: &DatabaseConnection{
			URI:      "postgresql://user:pass@host.com/db",
			Host:     "host.com",
			Port:     1234,
			User:     "user",
			Password: "pass",
			SSL:      true,
			Database: "db",
		},
		PrivateConnection: &DatabaseConnection{
			URI:      "postgresql://user:pass@private-host.com/db",
			Host:     "private-host.com",
			Port:     1234,
			User:     "user",
			Password: "pass",
			SSL:      true,
			Database: "db",
		},
		PrivateNetworkUUID: "deadbeef-dead-4aa5-beef-deadbeef347d",
	}

	body := `
{
  "replica": {
    "name": "pool",
    "region": "nyc1",
    "status": "online",
    "created_at": "` + createdAt.Format(time.RFC3339) + `",
    "connection": {
      "uri": "postgresql://user:pass@host.com/db",
      "host": "host.com",
      "port": 1234,
      "user": "user",
      "password": "pass",
      "database": "db",
      "ssl": true
    },
    "private_connection": {
      "uri": "postgresql://user:pass@private-host.com/db",
      "host": "private-host.com",
      "port": 1234,
      "user": "user",
      "password": "pass",
      "database": "db",
      "ssl": true
    },
    "private_network_uuid": "deadbeef-dead-4aa5-beef-deadbeef347d"
  }
}
`
	path := fmt.Sprintf("/v2/databases/%s/replicas", dbID)

	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		fmt.Fprint(w, body)
	})

	got, _, err := client.Databases.CreateReplica(ctx, dbID, &DatabaseCreateReplicaRequest{
		Name:               "replica",
		Region:             "nyc1",
		Size:               "db-s-2vcpu-4gb",
		PrivateNetworkUUID: privateNetworkUUID,
	})
	require.NoError(t, err)
	require.Equal(t, want, got)
}

func TestDatabases_DeleteReplica(t *testing.T) {
	setup()
	defer teardown()

	dbID := "deadbeef-dead-4aa5-beef-deadbeef347d"

	path := fmt.Sprintf("/v2/databases/%s/replicas/replica", dbID)

	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
	})

	_, err := client.Databases.DeleteReplica(ctx, dbID, "replica")
	require.NoError(t, err)
}

func TestDatabases_SetEvictionPolicy(t *testing.T) {
	setup()
	defer teardown()

	dbID := "deadbeef-dead-4aa5-beef-deadbeef347d"

	path := fmt.Sprintf("/v2/databases/%s/eviction_policy", dbID)

	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPut)
	})

	_, err := client.Databases.SetEvictionPolicy(ctx, dbID, "allkeys_lru")
	require.NoError(t, err)
}

func TestDatabases_GetEvictionPolicy(t *testing.T) {
	setup()
	defer teardown()

	dbID := "deadbeef-dead-4aa5-beef-deadbeef347d"

	want := "allkeys_lru"

	body := `{ "eviction_policy": "allkeys_lru" }`

	path := fmt.Sprintf("/v2/databases/%s/eviction_policy", dbID)

	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, body)
	})

	got, _, err := client.Databases.GetEvictionPolicy(ctx, dbID)
	require.NoError(t, err)
	require.Equal(t, want, got)
}

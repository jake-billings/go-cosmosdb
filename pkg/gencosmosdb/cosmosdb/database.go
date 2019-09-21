package cosmosdb

import (
	"encoding/base64"
	"net/http"
)

// Database represents a database
type Database struct {
	ID          string `json:"id,omitempty"`
	ResourceID  string `json:"_rid,omitempty"`
	Timestamp   int    `json:"_ts,omitempty"`
	Self        string `json:"_self,omitempty"`
	ETag        string `json:"_etag,omitempty"`
	Collections string `json:"_colls,omitempty"`
	Users       string `json:"_users,omitempty"`
}

// Databases represents databases
type Databases struct {
	Count      int        `json:"_count,omitempty"`
	ResourceID string     `json:"_rid,omitempty"`
	Databases  []Database `json:"Databases,omitempty"`
}

type databaseClient struct {
	hc              *http.Client
	databaseAccount string
	masterKey       []byte
}

// DatabaseClient is a database client
type DatabaseClient interface {
	Create(*Database) (*Database, error)
	List() DatabaseIterator
	All(DatabaseIterator) (*Databases, error)
	Get(string) (*Database, error)
	Delete(*Database) error
}

type databaseListIterator struct {
	*databaseClient
	continuation string
	done         bool
}

// DatabaseIterator is a database iterator
type DatabaseIterator interface {
	Next() (*Databases, error)
}

// NewDatabaseClient returns a new database client
func NewDatabaseClient(hc *http.Client, databaseAccount, masterKey string) (DatabaseClient, error) {
	var err error

	c := &databaseClient{
		hc:              hc,
		databaseAccount: databaseAccount,
	}

	c.masterKey, err = base64.StdEncoding.DecodeString(masterKey)
	if err != nil {
		return nil, err
	}

	return c, nil
}

func (c *databaseClient) Create(newdb *Database) (db *Database, err error) {
	err = c.do(http.MethodPost, "dbs", "dbs", "", http.StatusCreated, &newdb, &db, nil)
	return
}

func (c *databaseClient) List() DatabaseIterator {
	return &databaseListIterator{databaseClient: c}
}

func (c *databaseClient) All(i DatabaseIterator) (*Databases, error) {
	alldbs := &Databases{}

	for {
		dbs, err := i.Next()
		if err != nil {
			return nil, err
		}
		if dbs == nil {
			break
		}

		alldbs.Count += dbs.Count
		alldbs.ResourceID = dbs.ResourceID
		alldbs.Databases = append(alldbs.Databases, dbs.Databases...)
	}

	return alldbs, nil
}

func (c *databaseClient) Get(dbid string) (db *Database, err error) {
	err = c.do(http.MethodGet, "dbs/"+dbid, "dbs", "dbs/"+dbid, http.StatusOK, nil, &db, nil)
	return
}

func (c *databaseClient) Delete(db *Database) error {
	if db.ETag == "" {
		return ErrETagRequired
	}
	headers := http.Header{}
	headers.Set("If-Match", db.ETag)
	return c.do(http.MethodDelete, "dbs/"+db.ID, "dbs", "dbs/"+db.ID, http.StatusNoContent, nil, nil, headers)
}

func (i *databaseListIterator) Next() (dbs *Databases, err error) {
	if i.done {
		return
	}

	headers := http.Header{}
	if i.continuation != "" {
		headers.Set("X-Ms-Continuation", i.continuation)
	}

	err = i.do(http.MethodGet, "dbs", "dbs", "", http.StatusOK, nil, &dbs, headers)
	if err != nil {
		return
	}

	i.continuation = headers.Get("X-Ms-Continuation")
	i.done = i.continuation == ""

	return
}

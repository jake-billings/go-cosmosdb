package example

import (
	"net/http"
	"os"
	"testing"

	"github.com/jim-minter/go-cosmosdb/example/cosmosdb"
	"github.com/jim-minter/go-cosmosdb/example/types"
)

const (
	dbid     = "testdb"
	collid   = "people"
	personid = "jim"
)

func TestE2E(t *testing.T) {
	dbc, err := cosmosdb.NewDatabaseClient(http.DefaultClient, os.Getenv("DATABASE_ACCOUNT"), os.Getenv("MASTER_KEY"))
	if err != nil {
		t.Fatal(err)
	}

	db, err := dbc.Create(&cosmosdb.Database{ID: dbid})
	if err != nil {
		t.Error(err)
	}
	t.Logf("%#v\n", db)

	dbi := dbc.List()
	for {
		dbs, err := dbi.Next()
		if err != nil {
			t.Error(err)
		}
		if dbs == nil {
			break
		}
		t.Logf("%#v\n", dbs)
	}

	db, err = dbc.Get(dbid)
	if err != nil {
		t.Error(err)
	}
	t.Logf("%#v\n", db)

	collc := cosmosdb.NewCollectionClient(dbc, dbid)

	coll, err := collc.Create(&cosmosdb.Collection{
		ID: collid,
		PartitionKey: &cosmosdb.PartitionKey{
			Paths: []string{
				"/id",
			},
		},
	})
	if err != nil {
		t.Error(err)
	}
	t.Logf("%#v\n", coll)

	colli := collc.List()
	for {
		colls, err := colli.Next()
		if err != nil {
			t.Error(err)
		}
		if colls == nil {
			break
		}
		t.Logf("%#v\n", colls)
	}

	coll, err = collc.Get(collid)
	if err != nil {
		t.Error(err)
	}
	t.Logf("%#v\n", coll)

	pkrs, err := collc.PartitionKeyRanges(collid)
	if err != nil {
		t.Error(err)
	}
	t.Logf("%#v\n", pkrs)

	dc := cosmosdb.NewPersonClient(collc, collid, true)

	doc, err := dc.Create(personid, &types.Person{
		ID:      personid,
		Surname: "Minter",
	})
	if err != nil {
		t.Error(err)
	}
	t.Logf("%#v\n", doc)

	doci := dc.List()
	for {
		docs, err := doci.Next()
		if err != nil {
			t.Error(err)
		}
		if docs == nil {
			break
		}
		t.Logf("%#v\n", docs)
	}

	doc, err = dc.Get(personid, personid)
	if err != nil {
		t.Error(err)
	}
	t.Logf("%#v\n", doc)

	doci = dc.Query(&cosmosdb.Query{
		Query: "SELECT * FROM people WHERE people.surname = @surname",
		Parameters: []cosmosdb.Parameter{
			{
				Name:  "@surname",
				Value: "Minter",
			},
		},
	})
	for {
		docs, err := doci.Next()
		if err != nil {
			t.Error(err)
		}
		if docs == nil {
			break
		}
		t.Logf("%#v\n", docs)
	}

	oldETag := doc.ETag
	doc, err = dc.Replace(personid, &types.Person{
		ID:      personid,
		ETag:    doc.ETag,
		Surname: "Morrison",
	})
	if err != nil {
		t.Error(err)
	}
	t.Logf("%#v\n", doc)

	_, err = dc.Replace(personid, &types.Person{
		ID:      personid,
		ETag:    oldETag,
		Surname: "Henson",
	})
	if !cosmosdb.IsErrorStatusCode(err, http.StatusPreconditionFailed) {
		t.Error(err)
	}

	err = dc.Delete(personid, doc)
	if err != nil {
		t.Error(err)
	}

	err = collc.Delete(coll)
	if err != nil {
		t.Error(err)
	}

	err = dbc.Delete(db)
	if err != nil {
		t.Error(err)
	}
}
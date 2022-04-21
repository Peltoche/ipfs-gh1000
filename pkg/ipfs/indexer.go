package ipfs

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"time"

	"github.com/Peltoche/ipfs-gh1000/pkg/metadata"
	cid "github.com/ipfs/go-cid"
	shell "github.com/ipfs/go-ipfs-api"
	"github.com/ipld/go-ipld-prime/codec/dagjson"
	"github.com/ipld/go-ipld-prime/datamodel"
	"github.com/ipld/go-ipld-prime/fluent/qp"
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"
	"github.com/ipld/go-ipld-prime/node/basicnode"
)

type Indexer struct {
	shell      *shell.Shell
	indexName  string
	indexKeyID string
}

func NewIndexer(shell *shell.Shell, indexName string) (*Indexer, error) {
	keyList, err := shell.KeyList(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve the key list: %w", err)
	}

	var indexKeyID string

	for _, key := range keyList {
		if key.Name == indexName {
			indexKeyID = key.Id
		}
	}

	if indexKeyID == "" {
		return nil, fmt.Errorf("key %q not found in the key list", indexName)
	}

	return &Indexer{shell, indexName, indexKeyID}, nil
}

func (i *Indexer) RetrieveIndex(ctx context.Context) (map[string]metadata.RepoMetadata, error) {
	indexCID, err := i.shell.Resolve(i.indexKeyID)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve the index id: %w", err)
	}

	raw, err := i.shell.Cat(indexCID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve the raw index: %w", err)
	}

	np := basicnode.Prototype.Any // Pick a stle for the in-memory data.
	nb := np.NewBuilder()         // Create a builder.
	dagjson.Decode(nb, raw)       // Hand the builder to decoding -- decoding will fill it in!
	n := nb.Build()               // Call 'Build' to get the resulting Node.  (It's immutable!)

	it := n.MapIterator()

	res := map[string]metadata.RepoMetadata{}

	for {
		if it.Done() {
			break
		}

		var meta metadata.RepoMetadata

		mapKeyN, mapValueN, err := it.Next()
		if err != nil {
			return nil, fmt.Errorf("failed to decode map: %w", err)
		}
		metaKey, _ := mapKeyN.AsString()

		it2 := mapValueN.MapIterator()

		for {
			if it2.Done() {
				break
			}

			keyN, valueN, err := it2.Next()
			if err != nil {
				return nil, fmt.Errorf("failed to decode map: %w", err)
			}
			key, _ := keyN.AsString()

			switch key {
			case "url":
				meta.RepositoryURL, err = valueN.AsString()
			case "rank":
				var rank int64
				rank, err = valueN.AsInt()
				if err == nil {
					meta.Rank = int(rank)
				}
			case "stars":
				var nbStars int64
				nbStars, err = valueN.AsInt()
				if err == nil {
					meta.NbStars = int(nbStars)
				}
			case "lastMetadataFetch":
				rawDate, err := valueN.AsString()
				if err == nil {
					meta.LastMetadataFetch, err = time.Parse(time.RFC3339, rawDate)
				}
			case "repo":
				var c datamodel.Link
				c, err = valueN.AsLink()
				if err == nil {
					var repoCid cid.Cid
					repoCid, err = cid.Parse(c.String())
					meta.Repo = &repoCid
				}
			}
			if err != nil {
				return nil, fmt.Errorf("failed to parse %q fields: %w", key, err)
			}
		}

		res[metaKey] = meta
	}

	return res, nil
}

func (i *Indexer) SaveIndex(ctx context.Context, index map[string]metadata.RepoMetadata) error {

	n, err := qp.BuildMap(basicnode.Prototype.Any, int64(len(index)), func(ma datamodel.MapAssembler) {
		for name, data := range index {
			log.Printf("index: %s", name)
			qp.MapEntry(ma, name, qp.Map(5, func(ma datamodel.MapAssembler) {
				qp.MapEntry(ma, "url", qp.String(data.RepositoryURL))
				qp.MapEntry(ma, "rank", qp.Int(int64(data.Rank)))
				qp.MapEntry(ma, "stars", qp.Int(int64(data.NbStars)))
				qp.MapEntry(ma, "lastMetadataFetch", qp.String(data.LastMetadataFetch.Format(time.RFC3339)))

				lp := cidlink.Link{Cid: *data.Repo}
				qp.MapEntry(ma, "repo", qp.Link(lp))
			}))
		}
	})
	if err != nil {
		return fmt.Errorf("failed to compose the index: %w", err)
	}

	rawIndex := []byte{}
	rawIndexBuf := bytes.NewBuffer(rawIndex)

	err = dagjson.Encode(n, rawIndexBuf)
	if err != nil {
		return fmt.Errorf("failed to encode the index: %w", err)
	}

	indexCID, err := i.shell.Add(rawIndexBuf, shell.Pin(true))
	if err != nil {
		return fmt.Errorf("failed to save the new index: %s", err)
	}

	log.Println("start publishing the new index")
	lifetime, _ := time.ParseDuration("2400H") // 100 days
	ttl, _ := time.ParseDuration("1H")

	_, err = i.shell.PublishWithDetails(indexCID, i.indexName, lifetime, ttl, true)
	if err != nil {
		return fmt.Errorf("failed to publish the new index: %w", err)
	}
	log.Println("new index successfully published")

	return nil
}

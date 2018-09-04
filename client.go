package notion

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/tmc/notion/notiontypes"
)

const defaultBaseURL = "https://www.notion.so/api/v3/"

// Client is the primary type that implements an interface to the notion.so API.
type Client struct {
	baseURL string
	token   string
	client  *http.Client
	logger  Logger
}

// NewClient initializes a new Client.
func NewClient(opts ...ClientOption) (*Client, error) {
	c := &Client{
		baseURL: defaultBaseURL,
		logger:  &WrapLogrus{logrus.New()},
	}
	for _, o := range opts {
		o(c)
	}
	if c.client == nil {
		c.client = http.DefaultClient
	}
	return c, nil
}

func (c *Client) url(path string) string {
	return fmt.Sprintf("%s%s", c.baseURL, path)
}

func (c *Client) get(pattern string, args ...interface{}) ([]byte, error) {
	return c.do("GET", nil, pattern, args...)
}

func (c *Client) post(payload interface{}, pattern string, args ...interface{}) ([]byte, error) {
	buf := new(bytes.Buffer)
	if err := json.NewEncoder(buf).Encode(payload); err != nil {
		return nil, err
	}
	c.logger.WithField("fn", "post").Debugln(buf.String())
	return c.do("POST", buf, pattern, args...)
}

func (c *Client) do(method string, body io.Reader, pattern string, args ...interface{}) ([]byte, error) {
	path := c.url(fmt.Sprintf(pattern, args...))
	req, err := http.NewRequest(method, path, body)
	if err != nil {
		return nil, errors.Wrap(err, "creating request")
	}
	req.Header.Set("cookie", fmt.Sprintf("token=%v", c.token))
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "performing request")
	}
	defer resp.Body.Close()
	logger := c.logger.WithField("method", method).WithField("path", path).WithField("status_code", resp.StatusCode)
	buf, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.Warnln("error reading body")
		return nil, err
	}
	logger.WithField("body", string(buf)).Debugln("api call finished")
	if resp.StatusCode != http.StatusOK {
		return buf, &Error{
			URL:        path,
			StatusCode: resp.StatusCode,
			Body:       string(buf),
		}
	}
	return buf, nil
}

type getRecordValuesRequest struct {
	Requests []Record `json:"requests,omitempty"`
}

type getRecordValuesResponse struct {
	Results []*notiontypes.BlockWithRole `json:"results"`
}

// Record describes a type of notion.no entity.
//
// Example: block1 := Record{Table:"block","ID":"aa8fc12667704e83ad6c3968dcfc9b82"}
type Record struct {
	ID    string `json:"id"`
	Table string `json:"table"`
}

// GetRecordValues returns details about the given record types.
func (c *Client) GetRecordValues(records ...Record) ([]*notiontypes.BlockWithRole, error) {
	gr := getRecordValuesRequest{
		Requests: records,
	}
	r := &getRecordValuesResponse{}
	b, err := c.post(gr, "getRecordValues")
	if err != nil {
		return nil, err
	}
	c.logger.Debugln(string(b))
	if err := json.Unmarshal(b, r); err != nil {
		return nil, errors.Wrap(err, "unmarshaling getRecordValuesResponse")
	}
	return r.Results, nil
}

type loadPageChunkRequest struct {
	PageID          string `json:"pageId"`
	Limit           int64  `json:"limit,omitempty"`
	Cursor          Cursor `json:"cursor"`
	VerticalColumns bool   `json:"verticalColumns"`
}

type loadPageChunkResponse struct {
	RecordMap notiontypes.RecordMap `json:"recordMap"`
	Cursor    Cursor                `json:"cursor"`
}

// GetPage returns a Page given an id.
func (c *Client) GetPage(pageId string) (*Page, error) {
	b, err := c.GetBlock(pageId)
	return &Page{Block: b}, err
}

// GetBlock returns a Block given an id.
func (c *Client) GetBlock(blockID string) (*notiontypes.Block, error) {
	lp := loadPageChunkRequest{
		PageID: blockID,
		Limit:  50,
		Cursor: Cursor{
			Stack: [][]StackPosition{},
		},
	}
	results := []notiontypes.RecordMap{}
	for {
		r := &loadPageChunkResponse{}
		b, err := c.post(lp, "loadPageChunk")
		if err != nil {
			return nil, err
		}
		c.logger.WithField("blockID", blockID).Debugln(string(b))
		if err := json.Unmarshal(b, r); err != nil {
			return nil, errors.Wrap(err, "unmarshaling loadPageChunkResponse")
		}

		results = append(results, r.RecordMap)
		lp.Cursor = r.Cursor
		if len(r.Cursor.Stack) == 0 {
			break
		}
	}
	return c.parseBlockFromRecordMaps(blockID, results)
}

func mergeRecordMaps(rms ...notiontypes.RecordMap) (notiontypes.RecordMap, error) {
	result := notiontypes.RecordMap{
		Blocks:          make(map[string]*notiontypes.BlockWithRole, 50*len(rms)-1),
		Space:           make(map[string]*notiontypes.SpaceWithRole, 0),
		Users:           make(map[string]*notiontypes.UserWithRole, 0),
		Collections:     make(map[string]*notiontypes.CollectionWithRole, 0),
		CollectionViews: make(map[string]*notiontypes.CollectionViewWithRole, 0),
	}
	// TODO: consider merging into first recordmap as a heap optimization.

	for _, rm := range rms {
		for k, v := range rm.Blocks {
			result.Blocks[k] = v
		}
		for k, v := range rm.Space {
			result.Space[k] = v
		}
		for k, v := range rm.Users {
			result.Users[k] = v
		}
		for k, v := range rm.Collections {
			result.Collections[k] = v
		}
		for k, v := range rm.CollectionViews {
			result.CollectionViews[k] = v
		}
	}
	return result, nil
}

func (c *Client) parseBlockFromRecordMaps(blockID string, responses []notiontypes.RecordMap) (*notiontypes.Block, error) {
	rm, err := mergeRecordMaps(responses...)
	if err != nil {
		return nil, err
	}
	blockBlock, ok := rm.Blocks[blockID]
	if !ok {
		return nil, fmt.Errorf("notion: missing block id in block list")
	}
	block := blockBlock.Value
	blocks := make(map[string]*notiontypes.Block, len(rm.Blocks))
	for k, v := range rm.Blocks {
		blocks[k] = v.Value
	}
	if err := notiontypes.ResolveBlock(block, blocks); err != nil {
		return nil, errors.Wrap(err, "resolveBlock failed")
	}
	return block, nil
}

type operation struct {
	ID      string     `json:"id"`
	Table   string     `json:"table"`
	Path    []string   `json:"path"`
	Command string     `json:"command"`
	Args    [][]string `json:"args"`
}

type submitTransactionRequest struct {
	Operations []*operation `json:"operations"`
}
type submitTransactionResponse map[string]interface{}

// UpdateBlock returns a Block given an id.
func (c *Client) UpdateBlock(blockID string, path string, value string) error {
	lp := submitTransactionRequest{
		Operations: []*operation{
			&operation{
				ID:      blockID,
				Table:   "block",
				Path:    strings.Split(path, "."),
				Command: "set",
				Args: [][]string{
					[]string{value},
				},
			},
		},
	}
	r := &submitTransactionResponse{}
	b, err := c.post(lp, "submitTransaction")
	if err != nil {
		return err
	}
	c.logger.WithField("blockID", blockID).Debugln(string(b))
	c.logger.Debugln("resp:", r)
	return nil
}

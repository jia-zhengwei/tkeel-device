package service

import (
	"bytes"
	"context"
	json "encoding/json"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/tkeel-io/kit/log"
	transportHTTP "github.com/tkeel-io/kit/transport/http"
	"io/ioutil"
	"net/http"
	"strings"
)

const coreUrl string = "http://192.168.123.5:30707/v0.1.0/core/plugins/device/entities"
const authUrl string = "http://192.168.123.5:30707/v0.1.0/auth/token/parse" // /invoke/keel/method
const tokenKey string = "Authentication"

type CoreClient struct {
	//entity_create = "id={entity_id}&type={entity_type}&owner={user_id}&source={source}".format(**query)
	url 	   string
	id  	   string
	entityType string
	owner      string
	source     string
}

func NewCoreClient() *CoreClient {
	return &CoreClient{}
}

// get core url
func (c *CoreClient) GetCoreUrl(midUrl string) string {
	url :=  fmt.Sprintf(coreUrl + midUrl + "?" +  "id=%s&type=%s&owner=%s&source=%s", c.id, c.entityType, c.owner, c.source)
	c.url = url
	return url
}

//get token
func (c *CoreClient) GetToken(ctx context.Context) (string, error) {
	header := transportHTTP.HeaderFromContext(ctx)
	token, ok := header[tokenKey]
	if !ok {
		return "", errors.New("invalid Authentication")
	}
	// only use the first one
	if err := c.parseToken(token[0]); nil != err{
		return "", err
	}
	return token[0], nil
}

func (c *CoreClient) parseToken(token string) (error) {
	data := map[string]string{
		"entity_token": token,
	}
	dt, err := json.Marshal(data)
	if nil != err {
		log.Error("Marshal err", err)
		return err
	}
	resp, err := http.Post(authUrl, "application/json", bytes.NewBuffer(dt))

	resp2, err2 := c.ParseResp(resp, err)
	if nil != err2 {
		log.Debug("error parse token, ", err)
		return err2
	}
	var ar interface{}
	if err3 := json.Unmarshal(resp2, &ar); nil != err3 {
		log.Error("resp Unmarshal error", err3)
		return err3
	}
	//log.Debug("Unmarshal res:", ar)
	res, err4 := ar.(map[string]interface{})
	if false == err4 {
		return errors.New("auth error")
	}
	if res["ret"].(float64) != 0 {
		return errors.New(res["msg"].(string))

	}
	tokenMap := res["data"].(map[string]interface{})

	// save token, map[entity_id:406c79543e0245a994a742e69ce48e71 entity_type:device tenant_id: token_id:de25624a-1d0a-4ab0-b1f1-5b0db5a12c30 user_id:abc]
	c.owner = tokenMap["user_id"].(string)
	c.id    = tokenMap["entity_id"].(string)
	c.entityType    = tokenMap["entity_type"].(string)
	c.source        = "device"

	return nil
}

func (c *CoreClient) Post(ctx context.Context, data []byte) ([]byte, error) {
	resp, err := http.Post(c.url, "application/json", bytes.NewBuffer(data))

	return c.ParseResp(resp, err)
}

func (c *CoreClient) Get(ctx context.Context, id string) ([]byte, error) {
	log.Debug("fixme id", id)
	resp, err := http.Get(c.url)

	return c.ParseResp(resp, err)
}

func (c *CoreClient) Put(ctx context.Context, data []byte) ([]byte, error) {
	payload := strings.NewReader(string(data))
	req, _ := http.NewRequest("PUT", c.url, payload)

	req.Header.Add("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	return c.ParseResp(resp, err)
}

func (c *CoreClient) Delete(ctx context.Context, data []byte) ([]byte, error) {
	req, _ := http.NewRequest("DELETE", c.url, nil)
	//req.Header.Add("Authorization", "xxxx")

	resp, err := http.DefaultClient.Do(req)
	return c.ParseResp(resp, err)
}

func (c *CoreClient) ParseResp(resp *http.Response, err error) ([]byte, error) {
	if err != nil {
		log.Error("error get", err, c.url)
		return nil, err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Error("error ReadAll", err)
		return body, err
	}

	log.Debug("receive resp, ", string(body))
	if resp.StatusCode != 200 {
		log.Error("bad status", resp.StatusCode)
		return body, errors.New(resp.Status)
	}
	return body, nil
}

// generate uuid
func GetUUID() string {
	id := uuid.New()
	return id.String()
}

package main

import (
	"bytes"
	"net/http"
	"strings"

	"encoding/json"
	//"github.com/CodisLabs/codis/pkg/models"
	"github.com/CodisLabs/codis/pkg/utils/errors"
	"github.com/CodisLabs/codis/pkg/utils/log"
	//"github.com/juju/errors"
	"io/ioutil"
)

const (
	METHOD_GET    HttpMethod = "GET"
	METHOD_POST   HttpMethod = "POST"
	METHOD_PUT    HttpMethod = "PUT"
	METHOD_DELETE HttpMethod = "DELETE"
)

type HttpMethod string

func jsonify(v interface{}) string {
	b, _ := json.MarshalIndent(v, "", "  ")
	return string(b)
}

func callApi(method HttpMethod, apiPath string, params interface{}, retVal interface{}) error {
	if apiPath[0] != '/' {
		return errors.Errorf("api path must starts with /")
	}
	url := "http://" + args.apiServer + apiPath
	client := &http.Client{Transport: http.DefaultTransport}

	b, err := json.Marshal(params)
	if err != nil {
		return errors.Trace(err)
	}

	req, err := http.NewRequest(string(method), url, strings.NewReader(string(b)))
	if err != nil {
		return errors.Trace(err)
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Errorf("can't connect to dashboard, please check 'dashboard_addr' is corrent in config file")
		return errors.Trace(err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return errors.Trace(err)
	}

	if resp.StatusCode == 200 {
		err := json.Unmarshal(body, retVal)
		if err != nil {
			return errors.Trace(err)
		}
		return nil
	}
	return errors.Errorf("http status code %d, %s", resp.StatusCode, string(body))
}

//call http url and get json, then decode to objptr
func httpCall(objPtr interface{}, url string, method string, arg interface{}) error {
	client := &http.Client{Transport: http.DefaultTransport}
	rw := &bytes.Buffer{}
	if arg != nil {
		buf, err := json.Marshal(arg)
		if err != nil {
			return errors.Trace(err)
		}
		rw.Write(buf)
	}

	req, err := http.NewRequest(method, url, rw)
	if err != nil {
		return errors.Trace(err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return errors.Trace(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode/100 != 2 {
		msg, _ := ioutil.ReadAll(resp.Body)
		return errors.Errorf("error: %d, message: %s", resp.StatusCode, string(msg))
	}

	if objPtr != nil {
		return json.NewDecoder(resp.Body).Decode(objPtr)
	}

	return nil
}

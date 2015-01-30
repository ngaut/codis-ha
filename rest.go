package main

import (
	"bytes"
	"net/http"

	"encoding/json"
	"github.com/juju/errors"
	"io/ioutil"
)

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

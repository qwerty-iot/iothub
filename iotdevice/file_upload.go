package iotdevice

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"
)

type FileUpload struct {
	c             *Client
	CorrelationId string `json:"correlationId"`
	HostName      string `json:"hostName"`
	ContainerName string `json:"containerName"`
	BlobName      string `json:"blobName"`
	SasToken      string `json:"sasToken"`
	fuc           FileUploadComplete
}

type FileUploadComplete struct {
	CorrelationId     string `json:"correlationId"`
	IsSuccess         bool   `json:"isSuccess"`
	StatusCode        int    `json:"statusCode"`
	StatusDescription string `json:"statusDescription"`
}

func (fu *FileUpload) Upload(data []byte) error {
	var bodyBytes bytes.Buffer
	bodyBytes.Write(data)
	req, err := http.NewRequest("PUT", "https://"+fu.HostName+"/"+fu.ContainerName+"/"+fu.BlobName+fu.SasToken, &bodyBytes)
	if err != nil {
		return err
	}

	req.Header.Set("x-ms-blob-type", "BlockBlob")

	rsp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer rsp.Body.Close()

	fu.fuc.StatusCode = rsp.StatusCode
	fu.fuc.StatusDescription = rsp.Status
	fu.fuc.IsSuccess = true
	fu.fuc.CorrelationId = fu.CorrelationId

	if rsp.StatusCode >= 400 {
		fu.fuc.IsSuccess = false
		body, _ := ioutil.ReadAll(rsp.Body)
		return fmt.Errorf("HTTP Error: (%d:%s) %s", rsp.StatusCode, rsp.Status, body)
	}

	return nil
}

func (fu *FileUpload) Complete() error {
	audience := fu.c.creds.GetHostName() + "/devices/" + url.QueryEscape(fu.c.creds.GetDeviceID())
	sas, err := fu.c.creds.Token(audience, time.Hour)
	if err != nil {
		return err
	}

	var bodyBytes bytes.Buffer
	data, _ := json.Marshal(fu.fuc)
	bodyBytes.Write(data)

	req, err := http.NewRequest("POST", "https://"+audience+"/files/notifications?api-version=2020-09-30", &bodyBytes)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", sas.String())

	rsp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer rsp.Body.Close()

	if rsp.StatusCode >= 400 {
		body, _ := ioutil.ReadAll(rsp.Body)
		return fmt.Errorf("HTTP Error: (%d:%s) %s", rsp.StatusCode, rsp.Status, body)
	}

	body, _ := ioutil.ReadAll(rsp.Body)

	fu.c.logger.Debugf("uploadBlob complete: %#v", body)

	return nil
}

func (c *Client) FileUpload(ctx context.Context, fileName string) (*FileUpload, error) {
	audience := c.creds.GetHostName() + "/devices/" + url.QueryEscape(c.creds.GetDeviceID())
	sas, err := c.creds.Token(audience, time.Hour)
	if err != nil {
		return nil, err
	}

	var bodyBytes bytes.Buffer
	bodyBytes.Write([]byte(fmt.Sprintf("{ \"blobName\": \"%s\" }", fileName)))

	req, err := http.NewRequest("POST", "https://"+audience+"/files?api-version=2020-09-30", &bodyBytes)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", sas.String())

	rsp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer rsp.Body.Close()

	if rsp.StatusCode >= 400 {
		body, _ := ioutil.ReadAll(rsp.Body)
		return nil, fmt.Errorf("HTTP Error: (%d:%s) %s", rsp.StatusCode, rsp.Status, body)
	}

	body, _ := ioutil.ReadAll(rsp.Body)

	c.logger.Debugf("uploadBlob init: %#v", body)

	var fu FileUpload
	fu.c = c
	err = json.Unmarshal(body, &fu)
	if err != nil {
		return nil, err
	}

	return &fu, nil
}

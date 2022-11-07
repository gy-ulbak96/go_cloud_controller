package cloudclient

import (
	"encoding/json"
	"fmt"
	"strings"
  "net/http"
  "io/ioutil"
	"k8s.io/klog/v2"
	// "reflect"
)

func (c *CloudClient) CreateServer(spec *ServerSpec) (Server, error)  {
	return c.createServer(spec)
}

func (c *CloudClient) createServer(spec *ServerSpec) (Server, error) {
	url := c.Url + "/servers"
	method := "POST"
	payload, err := json.Marshal(spec)
	if err != nil {
		fmt.Println(err)
		return Server{}, err
	}
	klog.Infof(string(payload))
	klog.Infof("id print")
  client := &http.Client {}
  req, err := http.NewRequest(method, url, strings.NewReader(string(payload)))
  if err != nil {
    fmt.Println(err)
    return Server{}, err
  }
  req.Header.Add("Content-Type", "application/json")

  res, err := client.Do(req)
  if err != nil {
    fmt.Println(err)
		klog.Infof("thistime3")
    return Server{}, err
  }
  defer res.Body.Close()

  body, err := ioutil.ReadAll(res.Body)
  if err != nil {
    fmt.Println(err)
    return Server{}, err
  }
	var data map[string]interface{}
	json.Unmarshal([]byte(body), &data)
	// Realserverid := ""
	// for key, value := range data {
	// 	fmt.Println(key, value)
	// 	if key == "id" {
	// 		Realserverid = string(value)
	// 	}
		
	// }
	fmt.Println(string(body))

	// fmt.Println(string(body.id))
	klog.Infof("this is result")

	if res.StatusCode != 201 {
		err := HttpError{
			Status: res.StatusCode,
			Message: data["message"].(string),
		}
		return Server{}, fmt.Errorf("%+v", err)
	}

	server := Server{
		Id: data["id"].(string),
	}

	return server, nil
}

func (c *CloudClient) GetServer(id string) (Server, error)  {
	return c.getServer(id)
}

func (c *CloudClient) getServer(id string) (Server, error) {
	url := c.Url + "/servers/" + id
	method := "GET"
  client := &http.Client {}
  req, err := http.NewRequest(method, url, nil)
  if err != nil {
    fmt.Println(err)
    return Server{}, err
  }
  req.Header.Add("Content-Type", "application/json")

  res, err := client.Do(req)
  if err != nil {
    fmt.Println(err)
    return Server{}, err
  }
  defer res.Body.Close()

  body, err := ioutil.ReadAll(res.Body)
  if err != nil {
    fmt.Println(err)
    return Server{}, err
  }
	var data map[string]interface{}
	json.Unmarshal([]byte(body), &data)
	//fmt.Println(string(body))

	if res.StatusCode != 200 {
		err := HttpError{
			Status: res.StatusCode,
			Message: data["message"].(string),
		}
		return Server{}, fmt.Errorf("%+v", err)
	}

	server := Server{
		Name: data["name"].(string),
		Id: data["id"].(string),
	}
	
	return server, nil
}

func (c *CloudClient) DeleteServer(id string) error  {
	return c.deleteServer(id)
}

func (c *CloudClient) deleteServer(id string) error {
	url := c.Url + "/servers/" + id
	method := "DELETE"
  client := &http.Client {}
  req, err := http.NewRequest(method, url, nil)
  if err != nil {
    fmt.Println(err)
    return err
  }
  req.Header.Add("Content-Type", "application/json")

  res, err := client.Do(req)
  if err != nil {
    fmt.Println(err)
    return err
  }
  defer res.Body.Close()

  body, err := ioutil.ReadAll(res.Body)
  if err != nil {
    fmt.Println(err)
    return err
  }
	var data map[string]interface{}
	json.Unmarshal([]byte(body), &data)
	//fmt.Println(string(body))

	if res.StatusCode != 204 {
		err := HttpError{
			Status: res.StatusCode,
			Message: data["message"].(string),
		}
		return fmt.Errorf("%+v", err)
	}
	
	return nil
}

func CreateClient(url string) (Client, error) {
	if url == "" {
		return nil, fmt.Errorf("Cloud url must be specified")
	}

	return &CloudClient{
		Url: url,
	}, nil
}


/*
Copyright 2019 The Knative Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	python3 "github.com/DataDog/go-python3"
	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/cloudevents/sdk-go/v2/event"
	"github.com/cloudevents/sdk-go/v2/protocol"
	"github.com/cloudevents/sdk-go/v2/protocol/http"
	// "github.com/kelseyhightower/envconfig"
)

// Define a small struct representing the data for our expected data.
type payload struct {
	Sequence int     `json:"id"`
	Message  Message `json:"message"`
}

type Message map[string]string

func ImportModule(dir, name string) *python3.PyObject {
	sysModule := python3.PyImport_ImportModule("sys")
	path := sysModule.GetAttrString("path")
	python3.PyList_Insert(path, 0, python3.PyUnicode_FromString(""))
	python3.PyList_Insert(path, 0, python3.PyUnicode_FromString(dir))
	return python3.PyImport_ImportModule(name)
}

func pythonRepr(o *python3.PyObject) (string, error) {
	if o == nil {
		return "", fmt.Errorf("object is nil")
	}

	s := o.Repr()
	if s == nil {
		python3.PyErr_Clear()
		return "", fmt.Errorf("failed to call Repr object method")
	}
	defer s.DecRef()

	return python3.PyUnicode_AsUTF8(s), nil
}

func handleReq(req map[string]string) string {
	path, _ := os.Getwd()
	handleModule := ImportModule(path+"/pyfiles", "index")
	handleFunc := handleModule.GetAttrString("handler")

	var args = python3.PyTuple_New(1)
	var params = python3.PyDict_New()
	for key, value := range req {
		py_key := python3.PyUnicode_FromString(key)
		py_value := python3.PyUnicode_FromString(value)

		python3.PyDict_SetItem(params, py_key, py_value)
	}
	python3.PyTuple_SetItem(args, 0, params)
	resStr := handleFunc.Call(args, python3.Py_None)
	funcResultStr, _ := pythonRepr(resStr)

	return funcResultStr
}

func gotEvent(inputEvent event.Event) (*event.Event, protocol.Result) {
	input_data := &payload{}
	if err := inputEvent.DataAs(input_data); err != nil {
		log.Printf("Got error while unmarshalling data: %s", err.Error())
		return nil, http.NewResult(400, "got error while unmarshalling data: %w", err)
	}

	log.Println("Received a new event: ")
	log.Printf("[%v] %s %s: %+v", inputEvent.Time(), inputEvent.Source(), inputEvent.Type(), input_data)
	res := strings.Trim(strings.Replace(handleReq(input_data.Message), `'`, `"`, -1), "\"")

	log.Printf("After handler: %s", res)
	resdata_raw := make(map[string]interface{})
	resdata_str := make(map[string]string)
	if err := json.Unmarshal([]byte(res), &resdata_raw); err != nil {
		log.Printf("Got error while unmarshalling data: %s", err.Error())
		return nil, http.NewResult(400, "got error while unmarshalling data: %w", err)
	}
	for key, value := range resdata_raw {
		strKey := fmt.Sprintf("%v", key)
		strValue := fmt.Sprintf("%v", value)
		resdata_str[strKey] = strValue
	}

	input_data.Message = resdata_str

	// Create output event
	outputEvent := inputEvent.Clone()

	if err := outputEvent.SetData(cloudevents.ApplicationJSON, input_data); err != nil {
		log.Printf("Got error while marshalling data: %s", err.Error())
		return nil, http.NewResult(500, "got error while marshalling data: %w", err)
	}

	log.Println("Transform the event to: ")
	log.Printf("[%s] %s %s: %+v", outputEvent.Time(), outputEvent.Source(), outputEvent.Type(), input_data)

	return &outputEvent, nil
}

func main() {
	python3.Py_Initialize()
	c, err := cloudevents.NewClientHTTP()
	if err != nil {
		log.Fatalf("failed to create client, %v", err)
	}

	log.Println("listening on 8080.")
	log.Fatal(c.StartReceiver(context.Background(), gotEvent))
}

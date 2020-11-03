package main

import (
	"fmt"
	"log"

	cloudevents "github.com/cloudevents/sdk-go/v2"
)

func receiver(event cloudevents.Event) {
	log.Printf("☁️  cloudevents.Event\n%s", event.String())

	data := make(map[string]interface{})
	if err := event.DataAs(&data); err != nil {
		log.Printf("Error while extracting cloudevent Data: %s", err.Error())
	}

	parsedData := make(map[string]string)
	for k, v := range data {
		parsedData[k] = fmt.Sprint(v)
	}
	log.Printf("parsedData: %+v", parsedData)

	jobs, err := jobGenerator.GenerateJob(parsedData)
	if err != nil {
		log.Print(err)
	}

	for _, job := range jobs {
		log.Printf("generated job: %s", job.Name)
	}
}

package main

import (
	"context"
  "fmt"
	"log"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
)

func main() {
	// Load the Shared AWS Configuration (~/.aws/config)
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatal(err)
	}
  client := ec2.NewFromConfig(cfg)
  res, err := client.StopInstances(context.TODO(),&ec2.StopInstancesInput{
    InstanceIds: []string{"i-08725fdb6f33ea8dd"},
  })
	if err != nil {
		log.Fatal(err)
	}
  fmt.Println(res)
}

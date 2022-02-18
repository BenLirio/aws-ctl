package main

import (
	"context"
  "os"
  "time"
  "errors"
  "fmt"
	"log"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
)

var EInvalidState error = errors.New("Invalid State")
var ENotReachable error = errors.New("State not reachable during wait")


type Action int
const (
  Start Action = 1<<iota
  Stop
  Show
)

func strToAction(actionStr string) (Action,bool) {
  if actionStr == "start" { return Start,true }
  if actionStr == "stop" { return Stop,true }
  if actionStr == "show" { return Show,true }
  return Action(0), false
}

type State int
const (
  Pending State = 1<<iota
  Running
  Stopping
  Stopped
)

func strToState(ss string) (State,bool) {
  if ss == "running" { return Running,true }
  if ss == "pending" { return Pending,true }
  if ss == "stopped" { return Stopped,true }
  if ss == "stopping" { return Stopping,true }
  return State(0), false
}

func stateToString(st State) string {
  if st == Running { return "running" }
  if st == Pending { return "pending" }
  if st == Stopped { return "stopped" }
  if st == Stopping { return "stopping" }
  panic("Invalid State")
}

type Instance struct {
  client *ec2.Client
  Id string
}

func (ins Instance) GetState() (State, error) {
  args := &ec2.DescribeInstancesInput{InstanceIds: []string{ins.Id}}
  res, err := ins.client.DescribeInstances(context.TODO(),args)
  if err != nil { return State(0), err }
  if len(res.Reservations) != 1 {
    return State(0), errors.New("Not exactly one reservation found")
  }
  if len(res.Reservations[0].Instances) != 1 {
    return State(0), errors.New("Not exactly one instance found")
  }
  st,ok := strToState(string(res.Reservations[0].Instances[0].State.Name))
  if !ok {
    return State(0), EInvalidState
  }
  return st,nil
}

func (ins Instance) WaitUntil(st State) error {
  if st != Running && st != Stopped { 
    return EInvalidState
  }
  var validSt State
  if st == Running { validSt = Pending }
  if st == Stopped { validSt = Stopping }
  i := 0
  for {
    curSt, err := ins.GetState()
    if err != nil { return err }
    if curSt == st { return nil }
    if curSt != validSt { return ENotReachable }
    time.Sleep(time.Second)
    if i%4 != 3 {
      fmt.Print(".")
    } else {
      fmt.Print("\b\b\b   \b\b\b")
    }
    i += 1
  }
}

func (ins Instance) MustStart() error {
  err := ins.WaitUntil(Stopped)
  if err != nil { 
    if err == ENotReachable {
      return errors.New("Already Pending or Running")
    } else {
      return err
    }
  }
  args := &ec2.StartInstancesInput{InstanceIds: []string{ins.Id}}
  _, err = ins.client.StartInstances(context.TODO(), args)
	if err != nil { return err }
  return nil
}

func (ins Instance) MustStop() error {
  err := ins.WaitUntil(Running)
  if err != nil {
    if err == ENotReachable {
      return errors.New("Already Stopping or Stopped")
    } else {
      return err
    }
  }
  hibernate := true
  args := &ec2.StopInstancesInput{Hibernate: &hibernate, InstanceIds: []string{ins.Id}}
  _, err = ins.client.StopInstances(context.TODO(),args)
	if err != nil { log.Fatal(err) }
  return nil
}

func main() {
	// Load the Shared AWS Configuration (~/.aws/config)
  if len(os.Args) != 2 {
    fmt.Println("Must supply one of [start|stop|show]")
    return
  }
  action, ok := strToAction(os.Args[1])
  if !ok {
    fmt.Println("Action must be one of [start|stop|show]")
    return
  }
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatal(err)
	}
  client := ec2.NewFromConfig(cfg)
  instance := Instance{client: client, Id: "i-08725fdb6f33ea8dd"}

  if action == Show {
    st, err := instance.GetState()
    if err != nil { log.Fatal(err) }
    fmt.Println(stateToString(st))
    return
  } else if action == Start {
    err := instance.MustStart()
    if err != nil { fmt.Println(err) }
    return
  } else if action == Stop {
    err := instance.MustStop()
    if err != nil { fmt.Println(err) }
    return
  }
  panic("Unmatched Action")
}

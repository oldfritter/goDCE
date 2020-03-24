package sneakerWorkers

import (
	"io/ioutil"
	"log"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

type Worker struct {
	Name       string            `yaml:"name"`
	Exchange   string            `yaml:"exchange"`
	RoutingKey string            `yaml:"routing_key"`
	Queue      string            `yaml:"queue"`
	Durable    bool              `yaml:"durable"`
	Ack        bool              `yaml:"ack"`
	Options    map[string]string `yaml:"options"`
	Arguments  map[string]string `yaml:"arguments"`
	Delays     []int32           `yaml:"delays"`
	Steps      []int32           `yaml:"steps"`
	Threads    int               `yaml:"threads"`
}

var AllWorkers []Worker

func InitWorkers() {
	path_str, _ := filepath.Abs("config/workers.yml")
	content, err := ioutil.ReadFile(path_str)
	if err != nil {
		log.Fatal(err)
	}
	yaml.Unmarshal(content, &AllWorkers)
}

func (worker Worker) GetName() string {
	return worker.Name
}
func (worker Worker) GetExchange() string {
	return worker.Exchange
}
func (worker Worker) GetRoutingKey() string {
	return worker.RoutingKey
}
func (worker Worker) GetQueue() string {
	return worker.Queue
}
func (worker Worker) GetDurable() bool {
	return worker.Durable
}
func (worker Worker) GetAck() bool {
	return worker.Ack
}
func (worker Worker) GetOptions() map[string]string {
	return worker.Options
}
func (worker Worker) GetArguments() map[string]string {
	return worker.Arguments
}
func (worker Worker) GetDelays() []int32 {
	return worker.Delays
}
func (worker Worker) GetSteps() []int32 {
	return worker.Steps
}
func (worker Worker) GetThreads() int {
	return worker.Threads
}

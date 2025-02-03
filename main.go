package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"flag"
	"github.com/eclipse/paho.mqtt.golang"
	"github.com/gen2brain/beeep"
	"gopkg.in/ini.v1"
)
func main() {
	// get config path
	configPath := flag.String("c", "", "Path to config file")
	flag.Parse()

	// default path
	if *configPath == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			log.Fatalf("ERROR: getting home directory: %v", err)
		}
		*configPath = filepath.Join(homeDir, ".config", "mqttpushnotify.ini")
	}

	fmt.Printf("INFO: Load config from: %s\n", *configPath)

	// load config
	cfg, err := ini.Load(*configPath)
	if err != nil {
		log.Fatalf("ERROR: loading config: %v", err)
	}

	// Read MQTT Config
	log.Printf("INFO: Load MQTT Config")
	mqttSection := cfg.Section("mqtt")
	server := mqttSection.Key("server").String()
	port := mqttSection.Key("port").String()
	topic := mqttSection.Key("topic").String()

	// Read Notify Config
	log.Printf("INFO: Load Notify Config")
	notificationSection := cfg.Section("notification")
	title := notificationSection.Key("title").String()
	text := notificationSection.Key("text").String()

	// Connect MQTT Client
	log.Printf("INFO: Connect MQTT Client")
	opts := mqtt.NewClientOptions().AddBroker(fmt.Sprintf("tcp://%s:%s", server, port))
	client := mqtt.NewClient(opts)

	if token := client.Connect(); token.Wait() && token.Error() != nil {
		log.Fatalf("ERROR: connecting to MQTT broker: %v", token.Error())
	}
	defer client.Disconnect(250)

	// Subscribe to Topic
	if token := client.Subscribe(topic, 0, func(client mqtt.Client, msg mqtt.Message) {
		// Generate Push
		err := beeep.Notify(title, text, "assets/information.png")
		log.Printf("INFO: Send out notification")
		if err != nil {
			log.Printf("ERROR: rendering notification: %v", err)
		}
	}); token.Wait() && token.Error() != nil {
		log.Fatalf("ERROR: subscribing topic: %v", token.Error())
	}

	// Wait
	select {}
}

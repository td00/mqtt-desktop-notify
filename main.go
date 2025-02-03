package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	mqtt "github.com/eclipse/paho.mqtt.golang"
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

	// check if config file exists
	if _, err := os.Stat(*configPath); os.IsNotExist(err) {
		// No config file found, ask user to create a new one
		fmt.Println("No config file found at", *configPath)
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Do you want to create a new config file? (Y/n): ")
		answer, _ := reader.ReadString('\n')
		answer = strings.TrimSpace(answer)
		if answer == "" {
			answer = "y" // default to "y"
		}

		if answer == "y" || answer == "Y" {
			// Create a new config
			createConfig(*configPath)
		} else {
			log.Fatalf("ERROR: No config file found and user chose not to create one.")
		}
	}

	// load config
	cfg, err := ini.Load(*configPath)
	if err != nil {
		log.Fatalf("ERROR: loading config: %v", err)
	}

	// Read MQTT Config
	log.Printf("INFO: Load MQTT Config")
	mqttSection := cfg.Section("mqtt")
	server := mqttSection.Key("server").String()
	if server == "" {
		server = "127.0.0.1" // Default value for server
	}
	port := mqttSection.Key("port").String()
	if port == "" {
		port = "1883" // Default value for port
	}
	topic := mqttSection.Key("topic").String()
	if topic == "" {
		topic = "mqtt-desktop-notify/default" // Default value for topic
	}
	username := mqttSection.Key("username").String()
	password := mqttSection.Key("password").String()

	// Ask user if notification should be configured
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Do you want to configure notification settings? (y/N): ")
	notificationAnswer, _ := reader.ReadString('\n')
	notificationAnswer = strings.TrimSpace(notificationAnswer)
	if notificationAnswer == "" {
		notificationAnswer = "n" // Default to "n"
	}

	// Read Notify Config
	var title, text string
	if notificationAnswer == "y" || notificationAnswer == "Y" {
		// User wants to configure notifications
		fmt.Print("Enter the notification title: ")
		title = getInput()
		fmt.Print("Enter the notification text: ")
		text = getInput()
	} else {
		// Use default notification settings
		title = "mqtt-desktop-notify"
		text = "new notification"
	}

	// Save config to file
	cfg.Section("mqtt").Key("server").SetValue(server)
	cfg.Section("mqtt").Key("port").SetValue(port)
	cfg.Section("mqtt").Key("topic").SetValue(topic)
	cfg.Section("mqtt").Key("username").SetValue(username)
	cfg.Section("mqtt").Key("password").SetValue(password)
	cfg.Section("notification").Key("title").SetValue(title)
	cfg.Section("notification").Key("text").SetValue(text)

	err = cfg.SaveTo(*configPath)
	if err != nil {
		log.Fatalf("ERROR: saving config: %v", err)
	}

	// Connect MQTT Client
	log.Printf("INFO: Connect MQTT Client")
	opts := mqtt.NewClientOptions().AddBroker(fmt.Sprintf("tcp://%s:%s", server, port))
	if username != "" && password != "" {
		opts.SetUsername(username)
		opts.SetPassword(password)
	}
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

func createConfig(configPath string) {
	// Create the config file
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		log.Fatalf("ERROR: creating config directory: %v", err)
	}

	// Create new config file
	file, err := os.Create(configPath)
	if err != nil {
		log.Fatalf("ERROR: creating config file: %v", err)
	}
	defer file.Close()

	writer := bufio.NewWriter(file)

	// MQTT Config
	fmt.Println("Enter MQTT Configuration:")
	fmt.Print("Server (default: 127.0.0.1): ")
	server := getInput()
	if server == "" {
		server = "127.0.0.1" // Default to 127.0.0.1 if no input
	}
	fmt.Print("Port (default: 1883): ")
	port := getInput()
	if port == "" {
		port = "1883" // Default to 1883 if no input
	}
	fmt.Print("Topic (default: mqtt-desktop-notify/default): ")
	topic := getInput()
	if topic == "" {
		topic = "mqtt-desktop-notify/default" // Default to mqtt-desktop-notify/default if no input
	}

	// Ask for username & password
	fmt.Print("Enter MQTT Username (press Enter for no username): ")
	username := getInput()
	fmt.Print("Enter MQTT Password (press Enter for no password): ")
	password := getInput()

	// Write the MQTT config to the file
	writer.WriteString("[mqtt]\n")
	writer.WriteString("server = " + server + "\n")
	writer.WriteString("port = " + port + "\n")
	writer.WriteString("topic = " + topic + "\n")
	if username != "" {
		writer.WriteString("username = " + username + "\n")
	}
	if password != "" {
		writer.WriteString("password = " + password + "\n")
	}

	// Write the notification config section
	writer.WriteString("[notification]\n")
	writer.WriteString("title = mqtt-desktop-notify\n")
	writer.WriteString("text = new notification\n")

	// Save the file
	writer.Flush()

	fmt.Println("Config file created successfully at", configPath)
}

func getInput() string {
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	return strings.TrimSpace(input)
}

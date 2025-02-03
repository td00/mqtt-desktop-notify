package main

import (
	"bufio"
	"encoding/json"
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

// define version
const version = "0.0.2"

func main() {
	// Global flags
	versionFlag := flag.Bool("v", false, "Show version information")
	configPath := flag.String("c", "", "Path to config file")

	// Parse the command-line arguments
	flag.Parse()

	// If the version flag is set, show version info and exit
	if *versionFlag {
		showVersion()
		return
	}

	// Check if the createconfig command is provided
	if len(os.Args) > 1 && os.Args[1] == "createconfig" {
		// Create the config interactively
		createConfig(*configPath)
	} else {
		// Start the application directly if no command is specified
		runApp(configPath)
	}
}

func runApp(configPath *string) {
	// Default path
	if *configPath == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			log.Fatalf("ERROR: getting home directory: %v", err)
		}
		*configPath = filepath.Join(homeDir, ".config", "mqttpushnotify.ini")
	}

	// Check if config file exists
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

	// Load config
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

	// Read Notify Config
	title := cfg.Section("notification").Key("title").String()
	text := cfg.Section("notification").Key("text").String()
	notificationType := cfg.Section("notification").Key("type").String()
	if notificationType == "" {
		log.Printf("WARN: notification type not set, defaulting to static")
		notificationType = "static" // Default to "static"
	}

	// Ensure default values for static notification type
	if notificationType == "static" {
		if title == "" {
			title = "mqtt-desktop-notify" // Default title
		}
		if text == "" {
			text = "your notification text could be here" // Default text
		}
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
		// Variables to hold the final notification title and text
		var notificationText string
		var notificationTitle string

		// Handle different notification types
		switch notificationType {
		case "static":
			// Ensure default values if title or text are empty
			notificationTitle = title
			notificationText = text
		case "dynamic":
			// Use only the message payload for text, title stays static
			notificationTitle = title
			notificationText = string(msg.Payload()) // Set the topic message content as text
		case "json":
			// Parse the payload as JSON to extract title and text
			var jsonData map[string]string
			err := json.Unmarshal(msg.Payload(), &jsonData)
			if err != nil {
				log.Printf("ERROR: parsing JSON payload: %v", err)
				return
			}
			notificationTitle = jsonData["title"]
			notificationText = jsonData["text"]
		default:
			log.Printf("ERROR: unknown notification type: %s", notificationType)
			return
		}

		// Generate Push Notification
		err := beeep.Notify(notificationTitle, notificationText, "assets/information.png")
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

	// Ask for notification configuration
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Do you want to configure notification settings? (y/N): ")
	notificationAnswer, _ := reader.ReadString('\n')
	notificationAnswer = strings.TrimSpace(notificationAnswer)
	if notificationAnswer == "" {
		notificationAnswer = "n" // Default to "n"
	}

	// Write the notification config section
	writer.WriteString("[notification]\n")
	if notificationAnswer == "y" || notificationAnswer == "Y" {
		fmt.Print("Enter the notification title: ")
		title := getInput()
		fmt.Print("Enter the notification text: ")
		text := getInput()
		fmt.Print("Enter the notification type (static/dynamic/json, default: static): ")
		notificationType := getInput()
		if notificationType == "" {
			notificationType = "static" // Default to "static"
		}
		writer.WriteString("title = " + title + "\n")
		writer.WriteString("text = " + text + "\n")
		writer.WriteString("type = " + notificationType + "\n")
	} else {
		// Use default notification settings
		writer.WriteString("title = mqtt-desktop-notify\n")
		writer.WriteString("text = new notification\n")
		writer.WriteString("type = static\n")
	}

	// Save the file
	writer.Flush()

	fmt.Println("Config file created successfully at", configPath)
}

func showVersion() {
	// Output the version information
	fmt.Println("mqtt-desktop-notify is running in version", version)
	fmt.Println("\nmqtt-desktop-notify is licensed under AGPLv3.")
	fmt.Println("\nFind out more: https://github.com/td00/mqtt-desktop-notify")
}

func getInput() string {
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	return strings.TrimSpace(input)
}

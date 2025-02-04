# MQTT Desktop Notify

A simple tool that subscribes to an MQTT topic and sends a push notification when the topic's value changes/updates.

[![Go](https://github.com/td00/mqtt-desktop-notify/actions/workflows/go.yml/badge.svg?branch=main)](https://github.com/td00/mqtt-desktop-notify/actions/workflows/go.yml)

## Default Config Location

By default, the configuration file is located at:

- **macOS/Linux**: `~/.config/mqttpushnotify.ini`
- **Windows**: `C:\Users\<YourUser>\.config\mqttpushnotify.ini`

If you wish to use a different location for the configuration file, you can specify it with the `-c` flag.

If no config is found at the location you'll be asked to create the config via the interactive CLI. (this can be combined with `-c`)

If you want to overwrite a config you can do so by specifying `createconfig` (this can be combined with `-c`)

### Examples

`./mqtt-desktop-notify -c example.ini` will load the file `example.ini` in your current directory

`./mqtt-desktop-notify createconfig -c example.ini` will **_overwrite_** the file `example.ini` in your current directory. **This will happen without warning!**
## Config Options

The configuration file (`mqttpushnotify.ini`) should contain the following sections and options:

```ini
[mqtt]
server = 127.0.0.1
port = 1883
username = ""  # Optional, leave empty if not needed
password = ""  # Optional, leave empty if not needed
topic = "topic/to/subscribe"  # MQTT topic to subscribe to

[notification]
type = static
title = "Ring Ring! 🔔"  # Notification title
text = "Your Text"  # Notification text
```

- **[mqtt]**
  - `server`: MQTT broker server address (IP or hostname).
  - `port`: MQTT broker port (default is 1883).
  - `username`: Optional MQTT username.
  - `password`: Optional MQTT password.
  - `topic`: The MQTT topic to subscribe to.

- **[notification]**
  - `type`: Type of notification parsing
    - `static`: shows `title` & `text` as notification
    - `dynamic`: shows `title` & gets `text` from payload
    - `json`: gets `title` & `text` from json
  - `title`: The title of the notification.
  - `text`: The content of the notification.

json formatting:

```json
{
  "title": "notification title",
  "text": "notification body"
}
```

## Downloading the Files

You can download the source code by cloning the GitHub repository:

```bash
git clone https://github.com/td00/mqtt-desktop-notify.git
cd mqtt-desktop-notify
```

Alternatively, you can download a precompiled binary for your platform (macOS, Windows, Linux) from the [releases page](https://github.com/td00/mqtt-desktop-notify/releases).

## Running on macOS (Non-Signed Application)

To run the application on macOS, you might encounter security restrictions because the application is not signed. Follow these steps to run the program:

1. Open **System Preferences** → **Security & Privacy** → **General**.
2. Click on "Allow Anyway" for the application if it was blocked.
3. Alternatively, run the application from the terminal using the following command to bypass security:

```bash
sudo spctl --master-disable
```

This will allow unsigned apps to run. Be sure to re-enable the Gatekeeper security after running the app:

```bash
sudo spctl --master-enable
```

## License

This project is licensed under the AGPLv3 License - see the [LICENSE](LICENSE) file for details.

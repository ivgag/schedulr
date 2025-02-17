# Schedulr

## About
Schedulr is a Telegram bot that lets users forward messages, extract event details using OpenAI and Deepseek, and schedule events on their calendars. Currently, only Google Calendar is supported.

## Features
- **Message Forwarding:** Users can forward messages to the bot.
- **Event Extraction:** Utilizes OpenAI and Deepseek to extract event details.
- **Calendar Scheduling:** Automatically schedules events on Google Calendar.

## License
This project is licensed under the EPL-2.0 License. See the [LICENSE](LICENSE) file for details.

## Installation

### Prerequisites
- Docker and Docker Compose must be installed.

### Setup

1. **Clone the repository:**

    ```bash
    git clone git@github.com:ivgag/schedulr.git
    cd schedulr
    ```
2. **Configure environment variables:**

    ```bash
    export TELEGRAM_BOT_TOKEN=""
    export OPEN_AI_API_KEY=""
    export DEEPSEEK_API_KEY=""
    export GOOGLE_CLIENT_ID=""
    export GOOGLE_CLIENT_SECRET=""
    export GOOGLE_REDIRECT_URL=""
    ```
3. **Build and run the containers:**
    
    ```bash
    docker compose build
    docker compose up
    ```
import os
import json
import logging
import asyncio
import random
from telethon.sync import TelegramClient
from telethon.sessions import StringSession

SESSION_STRING = os.getenv("TELEGRAM_SESSION_STRING", None)
API_ID = os.getenv("TELEGRAM_API_ID", None)
API_HASH = os.getenv("TELEGRAM_API_HASH", None)
DESTINATION_CHANNEL = "@V2RootConfigPilot"
LOG_DIR = "Logs"

if not os.path.exists(LOG_DIR):
    os.makedirs(LOG_DIR)

logger = logging.getLogger()
logger.setLevel(logging.INFO)
logger.handlers = []
file_handler = logging.FileHandler(os.path.join(LOG_DIR, "post_configs.log"), mode='w', encoding='utf-8')
file_handler.setFormatter(logging.Formatter("%(asctime)s - %(levelname)s - %(message)s"))
logger.addHandler(file_handler)

async def post_best_configs_to_channel(client):
    try:
        # Read BestConfigs_scored.json
        json_path = os.path.join("output", "BestConfigs_scored.json")
        if not os.path.exists(json_path):
            logger.error(f"File {json_path} does not exist")
            return

        with open(json_path, "r", encoding="utf-8") as f:
            configs = json.load(f)

        if not configs:
            logger.warning("No configs found in BestConfigs_scored.json")
            return

        # Select top 2 configs
        top_configs = configs[:2] if len(configs) >= 2 else configs

        # Post each config to the channel
        for config in top_configs:
            config_type = config.get("Protocol", "").capitalize()
            config_url = config.get("URL", "")
            if not config_type or not config_url:
                logger.warning(f"Invalid config format: {config}")
                continue

            message = f"⚙️🌐 {config_type} Config\n\n```{config_url}```\n\n🆔 @V2RootConfigPilot"

            try:
                await client.send_message(DESTINATION_CHANNEL, message, parse_mode="markdown")
                logger.info(f"Posted {config_type} config to {DESTINATION_CHANNEL}: {config_url}")
            except Exception as e:
                logger.error(f"Failed to post config to {DESTINATION_CHANNEL}: {str(e)}")

    except Exception as e:
        logger.error(f"Error in post_best_configs_to_channel: {str(e)}")

async def main():
    logger.info("Starting config posting process")

    if not SESSION_STRING:
        logger.error("No session string provided.")
        print("Please set TELEGRAM_SESSION_STRING in environment variables.")
        return
    if not API_ID or not API_HASH:
        logger.error("API ID or API Hash not provided.")
        print("Please set TELEGRAM_API_ID and TELEGRAM_API_HASH in environment variables.")
        return

    try:
        api_id = int(API_ID)
    except ValueError:
        logger.error("Invalid TELEGRAM_API_ID format. It must be a number.")
        print("Invalid TELEGRAM_API_ID format. It must be a number.")
        return

    session = StringSession(SESSION_STRING)

    try:
        async with TelegramClient(session, api_id, API_HASH) as client:
            if not await client.is_user_authorized():
                logger.error("Invalid session string.")
                print("Invalid session string. Generate a new one using generate_session.py.")
                return

            await post_best_configs_to_channel(client)

    except Exception as e:
        logger.error(f"Error in main loop: {str(e)}")
        print(f"Error in main loop: {str(e)}")
        return

    logger.info("Config posting process completed")

if __name__ == "__main__":
    asyncio.run(main())

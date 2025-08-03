import os
import re
import json
import logging
import asyncio
import argparse
import base64
from datetime import datetime
from telethon.sync import TelegramClient
from telethon.tl.types import Message
from telethon.errors import ChannelInvalidError, PeerIdInvalidError
from telethon.sessions import StringSession

SESSION_STRING = os.getenv("TELEGRAM_SESSION_STRING", None)
API_ID = os.getenv("TELEGRAM_API_ID", None)
API_HASH = os.getenv("TELEGRAM_API_HASH", None)
CHANNELS_FILE = "telegram_channels.json"
LOG_DIR = "Logs"
INVALID_CHANNELS_FILE = os.path.join(LOG_DIR, "invalid_channels.txt")
STATS_FILE = os.path.join(LOG_DIR, "channel_stats.json")
CONFIG_PATTERNS = {
    "vless": r"vless://[^\s]+",
    "vmess": r"vmess://[^\s]+",
    "shadowsocks": r"ss://[^\s]+",
    "trojan": r"trojan://[^\s]+"
}

if not os.path.exists(LOG_DIR):
    os.makedirs(LOG_DIR)

logger = logging.getLogger()
logger.setLevel(logging.INFO)
logger.handlers = []
file_handler = logging.FileHandler(os.path.join(LOG_DIR, "collector.log"), mode='w', encoding='utf-8')
file_handler.setFormatter(logging.Formatter("%(asctime)s - %(levelname)s - %(message)s"))
logger.addHandler(file_handler)

def load_channels():
    with open(CHANNELS_FILE, "r", encoding="utf-8") as f:
        channels = json.load(f)
    logger.info(f"Loaded {len(channels)} channels from {CHANNELS_FILE}")
    return channels

def update_channels(channels):
    with open(CHANNELS_FILE, "w", encoding="utf-8") as f:
        json.dump(channels, f, ensure_ascii=False, indent=4)
    logger.info(f"Updated {CHANNELS_FILE} with {len(channels)} channels")

if not os.path.exists("Config"):
    logger.info(f"Creating directory: Config")
    os.makedirs("Config")

def extract_server_address(config, protocol):
    try:
        if protocol == "vmess":
            config_data = config.split("vmess://")[1]
            decoded = base64.b64decode(config_data).decode("utf-8")
            config_json = json.loads(decoded)
            return config_json.get("add", "")
        else:
            match = re.search(r"@([^\s:]+):", config)
            if match:
                return match.group(1)
            match = re.search(r"{}://[^\s@]+?([^\s:]+):".format(protocol), config)
            if match:
                return match.group(1)
        return None
    except Exception as e:
        logger.error(f"Failed to extract server address from {config}: {str(e)}")
        return None

async def fetch_configs_from_channel(client, channel):
    configs = {"vless": [], "vmess": [], "shadowsocks": [], "trojan": []}
    try:
        await client.get_entity(channel)
    except (ChannelInvalidError, PeerIdInvalidError, ValueError) as e:
        logger.error(f"Channel {channel} does not exist or is inaccessible: {str(e)}")
        return configs, False

    try:
        message_count = 0
        today = datetime.now().date()

        async for message in client.iter_messages(channel, limit=200):
            message_count += 1
            if message.date:
                message_date = message.date.date()
            else:
                continue

            if message_date != today:
                continue

            if isinstance(message, Message) and message.message:
                text = message.message
                for protocol, pattern in CONFIG_PATTERNS.items():
                    matches = re.findall(pattern, text)
                    if matches:
                        logger.info(f"Found {len(matches)} {protocol} configs in message from {channel}: {matches}")
                        configs[protocol].extend(matches)

        logger.info(f"Processed {message_count} messages from {channel}, found {sum(len(v) for v in configs.values())} configs")
        return configs, True
    except Exception as e:
        logger.error(f"Failed to fetch from {channel}: {str(e)}")
        return configs, False

def save_configs(configs, protocol):
    output_file = os.path.join("Config", f"{protocol}.txt")
    logger.info(f"Saving configs to {output_file}")
    with open(output_file, "w", encoding="utf-8") as f:
        if configs:
            for config in configs:
                f.write(config + "\n")
            logger.info(f"Saved {len(configs)} {protocol} configs to {output_file}")
        else:
            f.write("No configs found for this protocol.\n")
            logger.info(f"No {protocol} configs found, wrote placeholder to {output_file}")

def save_invalid_channels(invalid_channels):
    logger.info(f"Saving invalid channels to {INVALID_CHANNELS_FILE}")
    with open(INVALID_CHANNELS_FILE, "w", encoding="utf-8") as f:
        if invalid_channels:
            for channel in invalid_channels:
                f.write(f"{channel}\n")
            logger.info(f"Saved {len(invalid_channels)} invalid channels to {INVALID_CHANNELS_FILE}")
        else:
            f.write("No invalid channels found.\n")
            logger.info(f"No invalid channels found, wrote placeholder to {INVALID_CHANNELS_FILE}")

def save_channel_stats(stats):
    logger.info(f"Saving channel stats to {STATS_FILE}")
    stats_list = [{"channel": channel, **data} for channel, data in stats.items()]
    sorted_stats = sorted(stats_list, key=lambda x: x["score"], reverse=True)
    with open(STATS_FILE, "w", encoding="utf-8") as f:
        json.dump(sorted_stats, f, ensure_ascii=False, indent=4)
    logger.info(f"Saved channel stats to {STATS_FILE}")

async def main():
    parser = argparse.ArgumentParser(description="Fetch VPN configs from Telegram channels")
    parser.add_argument("--output", default="configs.json", help="Output JSON file for configs")
    args = parser.parse_args()

    logger.info("Starting config collection process")
    invalid_channels = []
    channel_stats = {}

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

    TELEGRAM_CHANNELS = load_channels()
    session = StringSession(SESSION_STRING)

    try:
        async with TelegramClient(session, api_id, API_HASH) as client:
            if not await client.is_user_authorized():
                logger.error("Invalid session string.")
                print("Invalid session string. Generate a new one using generate_session.py.")
                return

            all_configs = {"vless": [], "vmess": [], "shadowsocks": [], "trojan": []}
            valid_channels = []
            for channel in TELEGRAM_CHANNELS:
                logger.info(f"Fetching configs from {channel}...")
                print(f"Fetching configs from {channel}...")
                try:
                    channel_configs, is_valid = await fetch_configs_from_channel(client, channel)
                    if not is_valid:
                        invalid_channels.append(channel)
                        channel_stats[channel] = {
                            "vless_count": 0,
                            "vmess_count": 0,
                            "shadowsocks_count": 0,
                            "trojan_count": 0,
                            "total_configs": 0,
                            "score": 0,
                            "error": "Channel does not exist or is inaccessible"
                        }
                        continue

                    valid_channels.append(channel)
                    total_configs = sum(len(configs) for configs in channel_configs.values())
                    score = total_configs

                    channel_stats[channel] = {
                        "vless_count": len(channel_configs["vless"]),
                        "vmess_count": len(channel_configs["vmess"]),
                        "shadowsocks_count": len(channel_configs["shadowsocks"]),
                        "trojan_count": len(channel_configs["trojan"]),
                        "total_configs": total_configs,
                        "score": score
                    }
                    for protocol in all_configs:
                        all_configs[protocol].extend(channel_configs[protocol])
                except Exception as e:
                    invalid_channels.append(channel)
                    channel_stats[channel] = {
                        "vless_count": 0,
                        "vmess_count": 0,
                        "shadowsocks_count": 0,
                        "trojan_count": 0,
                        "total_configs": 0,
                        "score": 0,
                        "error": str(e)
                    }
                    logger.error(f"Channel {channel} is invalid: {str(e)}")

            for protocol in all_configs:
                all_configs[protocol] = list(set(all_configs[protocol]))
                logger.info(f"Found {len(all_configs[protocol])} unique {protocol} configs")
            
            with open(args.output, "w", encoding="utf-8") as f:
                json.dump(all_configs, f, ensure_ascii=False, indent=4)
            logger.info(f"Saved configs to {args.output}")

            for protocol in all_configs:
                save_configs(all_configs[protocol], protocol)
            save_invalid_channels(invalid_channels)
            save_channel_stats(channel_stats)
            update_channels(valid_channels)

    except Exception as e:
        logger.error(f"Error in main loop: {str(e)}")
        print(f"Error in main loop: {str(e)}")
        return

    logger.info("Config collection process completed")

if __name__ == "__main__":
    asyncio.run(main())

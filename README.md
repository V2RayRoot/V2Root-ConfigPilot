# V2Root-ConfigPilot (Beta)

**Status: This project is in Beta and under active development (dev). Features and performance may change as we continue to improve the repository.**

V2Root-ConfigPilot is a fully automated VPN configuration optimizer running in GitHub Actions, designed to test and rank VLESS, VMess, Shadowsocks, and Trojan configs for performance, optimized for use in Iran.

**[نسخه فارسی (Persian)](README.fa.md)**

## Overview
V2Root-ConfigPilot fetches VPN configurations, tests them for connectivity and performance, and outputs the top 10 configurations. Results are published every 2 hours to the GitHub repository and the Telegram channel `@V2RootConfigPilot`.

## Features
- **Input**: Processes VPN configurations and outputs them to `configs.json`.
- **Output**:
  - `output/BestConfigs.txt`: Top 10 config URLs.
  - `output/BestConfigs_scored.json`: Top 10 configs with performance metrics.
- **Automation**: Runs every 2 hours via GitHub Actions, committing results to the repository and posting to `@V2RootConfigPilot`.
- **Tests**: Evaluates configs for connectivity, latency, throughput, and DNS resolution.

## Config Distribution
- The top 10 configurations are updated every 2 hours and available in:
  - The GitHub repository: `output/BestConfigs.txt` and `output/BestConfigs_scored.json`.
  - The Telegram channel: `@V2RootConfigPilot`.

## Notes
- This project is in **Beta** and under active **development (dev)**. Expect potential changes and improvements as we refine functionality.
- For optimal "Iran-tested" configs, use a self-hosted runner in an Iran-based environment with full network access (ICMP, TCP, HTTP).
- Ensure the `GITHUB_TOKEN` has `contents: write` permissions, or use a PAT stored as `REPO_TOKEN`.

## Contributing
Fork the repository, make changes, and submit a pull request. Community contributions are welcome to enhance the repository's functionality during this Beta phase.

## License
MIT License

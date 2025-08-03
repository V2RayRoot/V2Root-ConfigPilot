# V2Root-ConfigPilot

A fully automated VPN configuration optimizer that runs in GitHub Actions, testing and ranking VLESS, VMess, Shadowsocks, and Trojan configs for performance.

## Overview
V2Root-ConfigPilot fetches VPN configurations using a provided Python script (`FetchConfig.py`), tests them for connectivity, latency, throughput, and DNS resolution, and outputs the top 10 configurations in `BestConfigs.txt` and `BestConfigs_scored.json`.

## Features
- **Input**: Reads configs from `FetchConfig.py` output (`configs.json`).
- **Tests**:
  - TLS handshake success
  - ICMP ping latency (5 packets)
  - Throughput (~1 MiB download)
  - DNS resolution time
- **Scoring**: Weighted composite score based on test results.
- **Output**:
  - `output/BestConfigs.txt`: Top 10 config URLs.
  - `output/BestConfigs_scored.json`: Top 10 configs with scores and test results.
- **Automation**: Runs daily via GitHub Actions, commits results to the repository.

## Setup
1. **Clone the Repository**:
   ```bash
   git clone https://github.com/your-repo/v2root-configpilot.git
   cd v2root-configpilot
   ```

2. **Configure Secrets**:
   In your GitHub repository, add the following secrets under Settings > Secrets and variables > Actions:
   - `TELEGRAM_SESSION_STRING`
   - `TELEGRAM_API_ID`
   - `TELEGRAM_API_HASH`

3. **Customize Environment Variables** (optional):
   In `.github/workflows/config-pilot.yml`, adjust:
   - `TEST_TIMEOUT`: Duration for each test (default: 10s).
   - `MAX_CONCURRENCY`: Number of concurrent tests (default: 10).

4. **Run Workflow**:
   - The workflow runs daily at midnight UTC or can be triggered manually via GitHub Actions.

## Scoring Weights
The scoring algorithm in `scorer/scorer.go` uses:
- TLS handshake success: +25
- Latency <50 ms: +30, <100 ms: +15
- Throughput >5 Mbps: +30, >1 Mbps: +10
- DNS resolution <100 ms: +10
- No packet loss: +5

To adjust weights, modify `CalculateScore` in `scorer/scorer.go`.

## Extending Protocols
To add a new protocol:
1. Update `parser/parser.go` with a new regex pattern and validation logic.
2. Ensure `tester/tester.go` can handle the protocol's server/port extraction.

## Contributing
Fork the repository, make changes, and submit a pull request. Community contributions are welcome to add new metrics, protocols, or improve testing.

## License
MIT License
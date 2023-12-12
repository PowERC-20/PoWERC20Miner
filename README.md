# PoWERC20 Mining Tool

## Overview

This Go project is a fully-functional tool for participating in Ethereum's PoWERC20 mining process. It provides a practical implementation for connecting to the Ethereum blockchain, generating nonces, and submitting transactions once a valid nonce is found. This tool is ideal for those looking to engage in Ethereum mining or understand the underlying mechanics in a real-world scenario.

## Features

- **Mining Functionality**: Generates nonces and compares the resulting hash against a target, mimicking the actual mining process in Ethereum.
- **Parallel Mining Workers**: Utilizes Go's concurrency capabilities to deploy multiple mining workers, increasing the chance of finding a valid nonce.
- **Smart Contract Interaction**: Retrieves the current mining challenge and difficulty from a specified Ethereum smart contract.
- **Nonce Submission and Transaction Handling**: Submits the mining solution to the Ethereum network and handles the transaction process once a valid nonce is discovered.

## Installation and Setup

1. **Install Go**: Ensure Go 1.21.x is installed on your system.
2. **Clone the Repository**: Download this repository to your local machine.
3. **Dependency Installation**: Run `go get` to install the necessary dependencies.
4. **Build the Project**:
   - Navigate to the project directory.
   - Run `go build -o Powerc20Worker` to build the executable.

## Usage

1. **Configuration**:
   - Customize your Ethereum private key and the contract address in the `init()` function.
   - Adjust the number of mining workers as needed.

2. **Running the Tool**:
   - Launch the tool by executing `./Powerc20Worker` in your terminal.
   - You can use optional flags for specific configurations, for example: `./Powerc20Worker -privateKey YOUR_PRIVATE_KEY -contractAddress CONTRACT_ADDRESS -workerCount NUMBER_OF_WORKERS`.
  
## Declare

The project code is completely open source. The released version is compiled using Github Actions. If you have any questions about the code, please feel free to raise them or submit the code for security auditing anywhere.

## Contributing

Your contributions are welcome. Please adhere to the project's coding standards and include tests for any new features or fixes.

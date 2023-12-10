package main

import (
	"context"
	"crypto/rand"
	"flag"
	"fmt"
	"math/big"
	"sync"
	"time"

	"powerc20miner/abi"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/fatih/color"
	"github.com/gosuri/uilive"
	"github.com/sirupsen/logrus"
)

var (
	infuraURL       = "https://rpc.ankr.com/eth"
	privateKey      string
	contractAddress string
	workerCount     int
	logger          = logrus.New()
)

func init() {
	flag.StringVar(&privateKey, "privateKey", "", "Private key for the Ethereum account")
	flag.StringVar(&contractAddress, "contractAddress", "0xca9b78435Be8267922E7Ac5cDE70401e7502c9cc", "Address of the Ethereum contract")
	flag.IntVar(&workerCount, "workerCount", 10, "Number of concurrent mining workers")

	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
	})
}

func mineWorker(ctx context.Context, wg *sync.WaitGroup, contract *abi.PoWERC20, fromAddress common.Address, client *ethclient.Client, auth *bind.TransactOpts, resultChan chan<- *big.Int, errorChan chan<- error, challenge *big.Int, target *big.Int, hashCountChan chan<- int) {
	defer wg.Done()

	var nonce *big.Int
	var err error

	for {
		select {
		case <-ctx.Done():
			return
		default:
			nonce, err = rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 256))
			if err != nil {
				errorChan <- fmt.Errorf("failed to generate random nonce: %v", err)
				return
			}

			noncePadded := common.LeftPadBytes(nonce.Bytes(), 32)
			challengePadded := common.LeftPadBytes(challenge.Bytes(), 32)
			addressBytes := fromAddress.Bytes()
			data := append(challengePadded, append(addressBytes, noncePadded...)...)
			hash := crypto.Keccak256Hash(data)
			if hash.Big().Cmp(target) == -1 {
				resultChan <- nonce
				return
			}
			hashCountChan <- 1

		}
	}
}

func main() {
	banner := `
//  ____    __        _______ ____   ____ ____   ___    __  __ _                 
// |  _ \ __\ \      / / ____|  _ \ / ___|___ \ / _ \  |  \/  (_)_ __   ___ _ __ 
// | |_) / _ \ \ /\ / /|  _| | |_) | |     __) | | | | | |\/| | | '_ \ / _ \ '__|
// |  __/ (_) \ V  V / | |___|  _ <| |___ / __/| |_| | | |  | | | | | |  __/ |   
// |_|   \___/ \_/\_/  |_____|_| \_\\____|_____|\___/  |_|  |_|_|_| |_|\___|_|   
	`
	fmt.Println(banner)
	flag.Parse()
	writer := uilive.New()

	writer.Start()
	defer writer.Stop()

	logger.Info(color.GreenString("Establishing connection with Ethereum client..."))
	client, err := ethclient.Dial(infuraURL)
	if err != nil {
		logger.Fatalf("Failed to connect to the Ethereum client: %v", err)
	}
	logger.Info(color.GreenString("Successfully connected to Ethereum client."))
	privateKeyECDSA, err := crypto.HexToECDSA(privateKey)
	if err != nil {
		logger.Fatalf("Error in parsing private key: %v", err)
	}

	chainID, err := client.NetworkID(context.Background())
	if err != nil {
		logger.Fatalf("Failed to get chainID: %v", err)
	}
	logger.Infof(color.GreenString("Successfully connected to Ethereum network with Chain ID: %v"), chainID)

	auth, err := bind.NewKeyedTransactorWithChainID(privateKeyECDSA, chainID)
	if err != nil {
		logger.Fatalf("Failed to create transactor: %v", err)
	}

	contractAddr := common.HexToAddress(contractAddress)
	contract, err := abi.NewPoWERC20(contractAddr, client)
	if err != nil {
		logger.Fatalf("Failed to instantiate a Token contract: %v", err)
	}
	logger.Info(color.GreenString("PoWERC20 token contract successfully instantiated."))

	contractName, err := contract.Name(nil)
	if err != nil {
		logger.Fatalf("Failed to get contract name: %v", err)
	}
	logger.Infof(color.GreenString("Contract Name: %s"), color.RedString(contractName))

	challenge, err := contract.Challenge(nil)
	if err != nil {
		logger.Fatalf("Failed to get challenge: %v", err)
	}
	logger.Infof(color.GreenString("Current mining challenge number: %d"), challenge)

	difficulty, err := contract.Difficulty(nil)
	if err != nil {
		logger.Fatalf("Failed to get difficulty: %v", err)
	}
	logger.Infof(color.GreenString("Current mining difficulty level: %d"), difficulty)

	difficultyUint := uint(difficulty.Uint64())
	target := new(big.Int).Lsh(big.NewInt(1), 256-difficultyUint)
	logger.Infof(color.GreenString("Target number is: %d"), target)

	resultChan := make(chan *big.Int)
	errorChan := make(chan error)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	logger.Info(color.YellowString("Mining workers started..."))

	hashCountChan := make(chan int)
	totalHashCount := 0
	ticker := time.NewTicker(1 * time.Second)

	go func() {
		for {
			select {
			case <-ticker.C:
				timestamp := time.Now().Format("2006-01-02 15:04:05")
				hashesPerSecond := float64(totalHashCount) / 1000.0
				fmt.Fprintf(writer, "%s[%s] %s\n", color.BlueString("Mining"), timestamp, color.GreenString("Total hashes per second: %8.2f K/s", hashesPerSecond))
				totalHashCount = 0
			case count := <-hashCountChan:
				totalHashCount += count
			}
		}
	}()

	var wg sync.WaitGroup
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go mineWorker(ctx, &wg, contract, auth.From, client, auth, resultChan, errorChan, challenge, target, hashCountChan)
	}

	select {
	case nonce := <-resultChan:
		ticker.Stop()
		cancel()
		wg.Wait()
		logger.Infof(color.GreenString("Successfully discovered a valid nonce: %d"), nonce)
		logger.Info(color.YellowString("Submitting mining transaction with nonce..."))
		tx, err := contract.Mine(auth, nonce)
		if err != nil {
			logger.Fatalf("Failed to submit mine transaction: %v", err)
		}
		receipt, err := bind.WaitMined(context.Background(), client, tx)
		if err != nil {
			logger.Fatalf("Failed to mine the transaction: %v", err)
		}
		logger.Info(color.GreenString("Mining transaction successfully confirmed, Transaction Hash: %s"), color.CyanString(receipt.TxHash.Hex()))

	case err := <-errorChan:
		cancel()
		wg.Wait()
		logger.Fatalf("Mining operation failed due to an error: %v", err)
	}
	logger.Info(color.GreenString("Mining process successfully completed"))
}


package utils

import (
	"crypto"
	"time"
)

const HOPLIMIT_CONSTANT = uint32(10)
const RUNOR_TIMEOUT = time.Second
const DATA_REQUEST_TIMEOUT = 5 * time.Second
const SEARCH_REQUEST_TIMEOUT = 2 * time.Second
const ANTI_ENTROPY_FREQUENCY = time.Second
const CACHING_DURATION = 500 * time.Millisecond
const PRIV_FILE_REQ_TIMEOUT = 30 * time.Second
const FILE_EXCHANGE_TIMEOUT = 5 // Tolerated response time after sending out ACCEPT before receiving FIX FileExchangeRequest message
const CLIENT_IP = "127.0.0.1"
const UI_PORT = "8080"
const CHUNK_SIZE = 8192
const MSG_BUFFER = 100
const HASH_LENGTH = 32
const DEFAULT_BUDGET = uint64(2)
const MAX_BUDGET = uint64(32)
const MIN_THRESHOLD = uint32(2) // Threshold for answered search requests before stopping search
const TX_PUBLISH_HOP_LIMIT = uint32(10)
const BLOCK_PUBLISH_HOP_LIMIT = uint32(20)
const LEADING_ZEROES = 2 // For PoW
const DOWNLOAD_FOLDER = "_Downloads"
const SHARED_FOLDER = "_SharedFiles"
const STATE_FOLDER = "_States"
const KEY_FOLDER = "_Keys"
const HASH_ALGO = crypto.SHA1
const RSA_KEY_LENGTH = 550
const AES_KEY_LENGTH = 32
const POSTPEND_LENGTH = 50         // For Challenges
const CHALLENGE_FAIL_THRESHOLD = 5 // Number of fails tolerated for Challenges before dropping exchange file
const AES_MARKUP = 28              // Byte markup generated through AES encryption

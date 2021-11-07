package shared

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"time"

	"p2p/crypto"
)

func init() {
	rand.Seed(time.Now().Unix())
}

func MessageIn(conn Conn, bytes []byte) (*Message, error) {
	message := &Message{}
	err := json.Unmarshal(bytes, message)

	// If there is an error, check if message is encrypted, if so, decrypt and unmarshal
	if err != nil {
		var secret [32]byte
		secret, err = conn.GetSecret()
		if err == nil {
			// Decrypt
			bytes, err = crypto.Decrypt(bytes, secret)
			if err != nil {
				return message, err
			}

			// unmarshal into Request struct
			err = json.Unmarshal(bytes, message)
		}

		// if there was an error unmarshalling initially and either the message wasn't encrypted or unmarshaling the unencrypted message failed
		if err != nil {
			log.Print(err)
			return message, err
		}
	}

	return message, nil
}

func MessageOut(conn Conn, message *Message) ([]byte, error) {
	bytes, err := json.Marshal(message)
	if err != nil {
		return bytes, err
	}

	if message.Encrypt {
		var secret [32]byte
		secret, err = conn.GetSecret()
		if err != nil {
			return bytes, fmt.Errorf("cannot encrypt with an empty secret")
		}

		// encrypt message content
		bytes, err = crypto.Encrypt(bytes, secret)
		if err != nil {
			return bytes, err
		}
	}

	return bytes, nil
}

func GenPort() string {
	return ":" + strconv.Itoa(rand.Intn(65535-10000)+10000)
}

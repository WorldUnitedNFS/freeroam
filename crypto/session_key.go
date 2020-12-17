package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/md5"
	"encoding/base64"
	"encoding/binary"
	"fmt"
)

type SessionKey struct {
	CryptKey    []byte
	ChecksumKey []byte
	HmacKey     []byte
}

// Russian-doll MD5; if rounds=2, output=md5(md5(part))
func deriveKeyPart(part []byte, rounds int) ([]byte, error) {
	hash := md5.New()
	hashInput := part

	for i := 0; i < rounds; i++ {
		hash.Reset()
		_, err := hash.Write(hashInput)
		if err != nil {
			return nil, err
		}

		hashInput = hash.Sum(nil)
	}

	finalResult := hash.Sum(nil)
	return finalResult, nil
}

// Decodes a byte array to a SessionKey object.
func DecodeSessionKey(inputKey []byte) (SessionKey, error) {
	masterKey, err := deriveKeyPart(inputKey[:6], 2)
	if err != nil {
		return SessionKey{}, err
	}

	checksumKey, err := deriveKeyPart(inputKey[6:12], 2)
	if err != nil {
		return SessionKey{}, err
	}
	hmacKey, err := deriveKeyPart(inputKey[12:], 2)
	if err != nil {
		return SessionKey{}, err
	}

	return SessionKey{
		CryptKey:    masterKey,
		ChecksumKey: checksumKey,
		HmacKey:     hmacKey,
	}, nil
}

// Decodes a Base64-encoded string to a SessionKey object.
func ParseSessionKey(inputKey string) (SessionKey, error) {
	keyDecoded, err := base64.StdEncoding.DecodeString(inputKey)

	if err != nil {
		return SessionKey{}, err
	}

	if len(keyDecoded) != 16 {
		return SessionKey{}, fmt.Errorf("expected 16 byte key but got %d byte key", len(keyDecoded))
	}

	return DecodeSessionKey(keyDecoded)
}

// Generates a 16 byte initialization vector based on a sequence number.
func (sk SessionKey) GenerateIv(seq uint16) ([]byte, error) {
	mac := hmac.New(md5.New, sk.HmacKey)
	n, err := mac.Write([]byte{byte(seq & 0xff), byte((seq >> 8) & 0xff)})

	if err != nil {
		return nil, err
	}

	if n != 2 {
		return nil, fmt.Errorf("wrote %d bytes to IV HMAC, wanted to write 2 bytes", n)
	}

	return mac.Sum(nil), nil
}

// Generates a 32-bit checksum for a packet body.
func (sk SessionKey) GenerateChecksum(input []byte) (uint32, error) {
	mac := hmac.New(md5.New, sk.ChecksumKey)
	n, err := mac.Write(input)

	if err != nil {
		return 0, err
	}

	if n != len(input) {
		return 0, fmt.Errorf("wrote %d bytes to checksum HMAC, wanted to write %d bytes", n, len(input))
	}

	return binary.LittleEndian.Uint32(mac.Sum(nil)[:4]), nil
}

// Performs the cipher operation for a packet body.
func (sk SessionKey) Cipher(input []byte, seq uint16) ([]byte, error) {
	block, err := aes.NewCipher(sk.CryptKey)
	if err != nil {
		return nil, err
	}
	iv, err := sk.GenerateIv(seq)
	if err != nil {
		return nil, err
	}
	output := make([]byte, len(input))
	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(output, input)
	return output, nil
}

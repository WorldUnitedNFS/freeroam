package crypto

import (
	"bytes"
	"encoding/hex"
	"testing"
)

func TestSessionKeyParsing(t *testing.T) {
	sessionKey, err := ParseSessionKey("FMb88lBCpYFqrEzsigPNVA==")
	if err != nil {
		t.Error(err)
	}
	t.Logf("Crypto key   : %s", hex.Dump(sessionKey.CryptKey))
	t.Logf("Checksum key : %s", hex.Dump(sessionKey.ChecksumKey))
	t.Logf("HMAC key     : %s", hex.Dump(sessionKey.HmacKey))
}

func TestPacketChecksum(t *testing.T) {
	sessionKey, err := ParseSessionKey("FMb88lBCpYFqrEzsigPNVA==")
	if err != nil {
		t.Error(err)
	}

	computedChecksum, err := sessionKey.GenerateChecksum([]byte{0x50, 0xd9, 0xc0})
	if err != nil {
		t.Error(err)
	}

	if computedChecksum != 0x63b2bebe {
		t.Errorf("Expected checksum to be 0x63b2bebe, got 0x%08x", computedChecksum)
	}
}

func TestPacketEncrypt(t *testing.T) {
	sessionKey, err := ParseSessionKey("FMb88lBCpYFqrEzsigPNVA==")
	if err != nil {
		t.Error(err)
	}
	testPayload := []byte{0x8f, 0x6c, 0x09}
	encryptedPayload, err := sessionKey.Cipher(testPayload, 0)
	if err != nil {
		t.Error(err)
	}
	if !bytes.Equal(encryptedPayload, []byte{0x00, 0x2f, 0x7a}) {
		t.Errorf("Expected encrypted payload to be [00 2F 7A], got: %s", hex.Dump(encryptedPayload))
	}
}

func TestPacketDecrypt(t *testing.T) {
	sessionKey, err := ParseSessionKey("FMb88lBCpYFqrEzsigPNVA==")
	if err != nil {
		t.Error(err)
	}
	testPayload := []byte{0x00, 0x2f, 0x7a}
	decryptedPayload, err := sessionKey.Cipher(testPayload, 0)
	if err != nil {
		t.Error(err)
	}
	if !bytes.Equal(decryptedPayload, []byte{0x8f, 0x6c, 0x09}) {
		t.Errorf("Expected decrypted payload to be [8f 6c 09], got: %s", hex.Dump(decryptedPayload))
	}
}

package main

import (
	"bytes"
	"crypto/sha512"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"math/rand"
	"time"

	"golang.org/x/crypto/pbkdf2"
)

const (
	// Base64Encoding flag
	Base64Encoding = iota
	// HexEncoding flag
	HexEncoding
)

// HashBytes size of the generated hash (to be chosen accordint the the chosen algo)
const HashBytes = 64

// SaltBytes sise of the salt : larger salt means hashed passwords are more resistant to rainbow table
const SaltBytes = 16

// Iterations tune so that hashing the password takes about 1 second
const Iterations = 100000

// Algo is pbkdf2 algorithm
var Algo = sha512.New

// Encoding hex is readable but base64 is shorter
var Encoding = Base64Encoding

// HashFrameBytesLength raw hash generated length
var HashFrameBytesLength = HashBytes + SaltBytes + 8

// HashFrameStrLength base64 hash generated length
var HashFrameStrLength = ((4 * HashFrameBytesLength / 3) + 3) & -4

// ErrUnsupportedEncoding error
var ErrUnsupportedEncoding = errors.New("Unsupported encoding")

func init() {
	rand.Seed(time.Now().UnixNano())
}

// VerifyPasswdHash verify pbkdf2 password hash
func VerifyPasswdHash(hash []byte, passwd []byte) (ok bool) {
	ok = false
	var err error
	var rawHash []byte
	var rawHashLen int
	switch Encoding {
	case Base64Encoding:
		{
			rawHash = make([]byte, base64.StdEncoding.DecodedLen(len(hash)))
			rawHashLen, err = base64.StdEncoding.Decode(rawHash, hash)
			if err != nil {
				return
			}
			rawHash = rawHash[:rawHashLen]
			break
		}
	case HexEncoding:
		{
			rawHash = make([]byte, hex.DecodedLen(len(hash)))
			rawHashLen, err = hex.Decode(rawHash, hash)
			if err != nil {
				return
			}
			rawHash = rawHash[:rawHashLen]
		}
	default:
		{
			return
		}
	}

	if rawHashLen < 8 {
		return
	}

	saltBytes := binary.BigEndian.Uint32(rawHash[:4])
	hashBytes := uint32(rawHashLen) - saltBytes - 8
	iterations := int(binary.BigEndian.Uint32(rawHash[4:8]))

	if uint32(rawHashLen) < saltBytes+hashBytes+8 {
		return
	}
	salt := rawHash[8 : saltBytes+8]
	realHash := rawHash[8+saltBytes : saltBytes+hashBytes+8]
	return bytes.Equal(pbkdf2.Key(passwd, salt, iterations, int(hashBytes), Algo), realHash)
}

// CreatePasswdHash create password pbkdf2 hash
func CreatePasswdHash(passwd []byte) (encodedHash []byte, err error) {
	salt, err := GenerateRandomBytes(SaltBytes)
	if err != nil {
		return
	}
	realHash := pbkdf2.Key(passwd, salt, Iterations, HashBytes, Algo)
	hash := make([]byte, 8)
	binary.BigEndian.PutUint32(hash, SaltBytes)
	binary.BigEndian.PutUint32(hash[4:], Iterations)
	hash = append(hash, salt...)
	hash = append(hash, realHash...)
	switch Encoding {
	case Base64Encoding:
		{
			encodedHash = make([]byte, base64.StdEncoding.EncodedLen(HashFrameBytesLength))
			base64.StdEncoding.Encode(encodedHash, hash)
			return
		}
	case HexEncoding:
		{
			encodedHash = make([]byte, hex.EncodedLen(HashFrameBytesLength))
			hex.Encode(encodedHash, hash)
			return
		}
	default:
		{
			err = ErrUnsupportedEncoding
			return
		}
	}
}

// GenerateRandomBytes random secure bytes
func GenerateRandomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	// Note that err == nil only if we read len(b) bytes.
	if err != nil {
		return nil, err
	}

	return b, nil
}

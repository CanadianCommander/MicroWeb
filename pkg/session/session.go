package session

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha512"
	"encoding/binary"
	"errors"
	"fmt"
	"strings"

	"github.com/CanadianCommander/MicroWeb/pkg/logger"
)

//Serializable is an object that can be serialized in to the session cookie.
type Serializable interface {
	//called to produce a string which will be stored in session cookie
	MarshalBinary() (data []byte, err error)
	//called to initialize object from a string stored in the session cookie
	UnmarshalBinary(data []byte) error
	//identifier used to identify this objects string in the session cookie
	GetIdentifier() string
}

/*
NewSession constructs a new session with the given encryption key.
*/
func NewSession(key string) (*Session, error) {
	newSession := Session{}

	// generate iv
	iv := make([]byte, aes.BlockSize)
	_, err := rand.Read(iv)
	if err != nil {
		logger.LogError("failed to generate IV with error: %s", err.Error())
		return nil, err
	}

	newSession.SetIV(string(iv))
	newSession.SetKey(key)

	return &newSession, nil
}

/*
Session represents a session retrieved from an http request.
*/
type Session struct {
	//raw data found http request session cookie
	encIV          string
	encKey         string
	rawSessionData byte
	sessionObjects []Serializable
}

/*
FromBuffer builds the session object from a session object buffer
*/
func (session *Session) FromBuffer(buffer []byte) error {
	if len(buffer) <= (sha512.Size + aes.BlockSize) {
		errorStr := fmt.Sprintf("buffer invalid. buffer must be at least %d bytes long", (sha512.Size + aes.BlockSize))
		return errors.New(errorStr)
	}

	iv := string(buffer[len(buffer)-aes.BlockSize:])
	buffer = buffer[:len(buffer)-aes.BlockSize]
	key := session.GetKey()

	//format key
	encKey := sha512.Sum512_256([]byte(key))

	//decrypt cookie
	aesBlockCipher, err := aes.NewCipher(encKey[:])
	if err != nil {
		logger.LogError("cipher creation error: %s", err.Error())
		return err
	}
	cfbEnc := cipher.NewCFBDecrypter(aesBlockCipher, []byte(iv))

	plainText := make([]byte, len(buffer))
	cfbEnc.XORKeyStream(plainText, buffer)

	// check checksum
	checksumFromCookie := plainText[len(plainText)-sha512.Size:]
	plainText = plainText[:len(plainText)-sha512.Size]
	checksum := sha512.Sum512(plainText)

	if !bytes.Equal(checksumFromCookie, checksum[:]) {
		return errors.New("cookie checksum is incorrect. cookie checksum: " +
			string(checksumFromCookie) + " calculated checksum: " + string(checksum[:]))
	}

	// build session objects
	sessionMap := make(map[string]Serializable)
	for _, sessionObj := range session.GetSessionObjects() {
		sessionMap[sessionObj.GetIdentifier()] = sessionObj
	}

	cookieReader := strings.NewReader(string(plainText))
	for i := 0; i < cookieReader.Len(); {
		var objLenBuff [binary.MaxVarintLen64]byte

		// get object length
		read, err := cookieReader.Read(objLenBuff[:])
		i += read
		if err != nil {
			return err
		}
		objLen, _ := binary.Uvarint(objLenBuff[:])

		// get object identifier length
		read, err = cookieReader.Read(objLenBuff[:])
		i += read
		if err != nil {
			return err
		}
		objIDLen, _ := binary.Uvarint(objLenBuff[:])

		// get object identifier
		objIdentifier := make([]byte, objIDLen)
		read, err = cookieReader.Read(objIdentifier)
		i += read
		if err != nil {
			return err
		}

		sessionObject, ok := sessionMap[string(objIdentifier)]
		if !ok {
			logger.LogWarning("could not map session information to object with id [%s]", string(objIdentifier))
			i += int(objLen - objIDLen)
			continue
		}

		objData := make([]byte, objLen-objIDLen)
		read, err = cookieReader.Read(objData)
		i += read
		if err != nil {
			return err
		}
		sessionObject.UnmarshalBinary(objData)
	}

	return nil
}

/*
GetBuffer "compile" session in to byte buffer
*/
func (session *Session) GetBuffer() ([]byte, error) {
	cookieBuff := bytes.Buffer{}
	iv := session.GetIV()
	key := session.GetKey()

	//build cookie buffer
	for _, sessionObj := range session.GetSessionObjects() {
		rawTxt, err := sessionObj.MarshalBinary()
		if err != nil {
			logger.LogError("Failed to serialize object with error: %s", err.Error())
			return nil, err
		}

		var lenBuff [binary.MaxVarintLen64]byte
		binary.PutUvarint(lenBuff[:], uint64(len([]byte(rawTxt))+len([]byte(sessionObj.GetIdentifier()))))
		cookieBuff.Write(lenBuff[:]) //len of serialized obj

		binary.PutUvarint(lenBuff[:], uint64(len([]byte(sessionObj.GetIdentifier()))))
		cookieBuff.Write(lenBuff[:]) //len of identifier

		cookieBuff.Write([]byte(sessionObj.GetIdentifier())) //obj identifier
		cookieBuff.Write(rawTxt)                             // serialized object.
	}

	//append checksum
	checkSum := sha512.Sum512(cookieBuff.Bytes())
	cookieBuff.Write(checkSum[:])

	//format key
	encKey := sha512.Sum512_256([]byte(key))

	//encrypt cookie
	aesBlockCipher, err := aes.NewCipher(encKey[:])
	if err != nil {
		logger.LogError("cipher creation error: %s", err.Error())
		return nil, err
	}
	cfbEnc := cipher.NewCFBEncrypter(aesBlockCipher, []byte(iv))

	outputBuffer := make([]byte, cookieBuff.Len()+aes.BlockSize)
	cfbEnc.XORKeyStream(outputBuffer, cookieBuff.Bytes())

	//append iv to cookie
	copy(outputBuffer[cookieBuff.Len():], []byte(iv))

	return outputBuffer, nil
}

//Add adds a serializable object to the session
func (session *Session) Add(obj Serializable) {
	session.sessionObjects = append(session.sessionObjects, obj)
}

//SetIV sets the iv for the session
func (session *Session) SetIV(iv string) {
	session.encIV = iv
}

//SetKey sets the key for the session
func (session *Session) SetKey(key string) {
	session.encKey = key
}

//GetIV gets the iv for the session
func (session *Session) GetIV() string {
	return session.encIV
}

//GetKey gets the key for the session
func (session *Session) GetKey() string {
	return session.encKey
}

//GetSessionObjects returns a list of all objects associated with the session
func (session *Session) GetSessionObjects() []Serializable {
	return session.sessionObjects
}

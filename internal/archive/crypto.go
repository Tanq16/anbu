package archive

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"io"

	"golang.org/x/crypto/pbkdf2"
)

const (
	archiveVersion        byte = 1
	archiveSaltLen             = 16
	archiveNoncePrefixLen      = 8
	archiveKDFIter             = 100000
	archiveChunkSize           = 64 * 1024
)

var archiveMagic = []byte("ANBUZ")

type encryptWriter struct {
	w           io.Writer
	gcm         cipher.AEAD
	noncePrefix []byte
	counter     uint32
	buf         []byte
}

func newEncryptWriter(w io.Writer, password string) (*encryptWriter, error) {
	salt := make([]byte, archiveSaltLen)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return nil, err
	}
	noncePrefix := make([]byte, archiveNoncePrefixLen)
	if _, err := io.ReadFull(rand.Reader, noncePrefix); err != nil {
		return nil, err
	}
	gcm, err := newArchiveGCM(password, salt)
	if err != nil {
		return nil, err
	}
	if _, err := w.Write(archiveMagic); err != nil {
		return nil, err
	}
	if _, err := w.Write([]byte{archiveVersion}); err != nil {
		return nil, err
	}
	if _, err := w.Write(salt); err != nil {
		return nil, err
	}
	if _, err := w.Write(noncePrefix); err != nil {
		return nil, err
	}
	return &encryptWriter{
		w:           w,
		gcm:         gcm,
		noncePrefix: noncePrefix,
		buf:         make([]byte, 0, archiveChunkSize),
	}, nil
}

func (e *encryptWriter) Write(p []byte) (int, error) {
	n := 0
	for len(p) > 0 {
		space := archiveChunkSize - len(e.buf)
		take := min(space, len(p))
		e.buf = append(e.buf, p[:take]...)
		p = p[take:]
		n += take
		if len(e.buf) == archiveChunkSize {
			if err := e.flush(); err != nil {
				return n, err
			}
		}
	}
	return n, nil
}

func (e *encryptWriter) Close() error {
	return e.flush()
}

func (e *encryptWriter) flush() error {
	if len(e.buf) == 0 {
		return nil
	}
	sealed, err := e.seal(e.buf)
	if err != nil {
		return err
	}
	e.buf = e.buf[:0]
	var lenbuf [4]byte
	binary.BigEndian.PutUint32(lenbuf[:], uint32(len(sealed)))
	if _, err := e.w.Write(lenbuf[:]); err != nil {
		return err
	}
	_, err = e.w.Write(sealed)
	return err
}

func (e *encryptWriter) seal(plain []byte) ([]byte, error) {
	if e.counter == ^uint32(0) {
		return nil, fmt.Errorf("archive chunk counter overflow")
	}
	nonce := make([]byte, archiveNoncePrefixLen+4)
	copy(nonce, e.noncePrefix)
	binary.BigEndian.PutUint32(nonce[archiveNoncePrefixLen:], e.counter)
	e.counter++
	return e.gcm.Seal(nil, nonce, plain, nil), nil
}

func decryptTo(dst io.Writer, src io.Reader, password string) error {
	magic := make([]byte, len(archiveMagic))
	if _, err := io.ReadFull(src, magic); err != nil {
		return err
	}
	if !bytes.Equal(magic, archiveMagic) {
		return fmt.Errorf("not an encrypted archive")
	}
	var ver [1]byte
	if _, err := io.ReadFull(src, ver[:]); err != nil {
		return err
	}
	if ver[0] != archiveVersion {
		return fmt.Errorf("unsupported archive version %d", ver[0])
	}
	salt := make([]byte, archiveSaltLen)
	if _, err := io.ReadFull(src, salt); err != nil {
		return err
	}
	noncePrefix := make([]byte, archiveNoncePrefixLen)
	if _, err := io.ReadFull(src, noncePrefix); err != nil {
		return err
	}
	gcm, err := newArchiveGCM(password, salt)
	if err != nil {
		return err
	}
	maxSealed := archiveChunkSize + gcm.Overhead()
	var counter uint32
	for {
		var lenbuf [4]byte
		_, err := io.ReadFull(src, lenbuf[:])
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		n := binary.BigEndian.Uint32(lenbuf[:])
		if n < uint32(gcm.Overhead()) || n > uint32(maxSealed) {
			return fmt.Errorf("invalid archive chunk size")
		}
		sealed := make([]byte, n)
		if _, err := io.ReadFull(src, sealed); err != nil {
			return err
		}
		nonce := make([]byte, archiveNoncePrefixLen+4)
		copy(nonce, noncePrefix)
		binary.BigEndian.PutUint32(nonce[archiveNoncePrefixLen:], counter)
		counter++
		plain, err := gcm.Open(nil, nonce, sealed, nil)
		if err != nil {
			return err
		}
		if _, err := dst.Write(plain); err != nil {
			return err
		}
	}
}

func newArchiveGCM(password string, salt []byte) (cipher.AEAD, error) {
	key := pbkdf2.Key([]byte(password), salt, archiveKDFIter, 32, sha256.New)
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	if gcm.NonceSize() != archiveNoncePrefixLen+4 {
		return nil, fmt.Errorf("unexpected GCM nonce size")
	}
	return gcm, nil
}

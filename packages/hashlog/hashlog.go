// Implements funtions to initialize and add data to the loggers implemented by the log package, with additional checks and error handling against an SHA256 checksum to detect file tampering.
// HashLogs are concurrency-safe by virtue of implement log.Loggers, which are themselves concurrency safe.
// Initialized in init(), and updated throughout the application in real time.
package hashlog

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
)

type HashLog struct {
	Name         string      // Name of logger, determinds log and checksum file directories.
	L            *log.Logger // Embeds a *log.Logger
	LogPath      string      // File directory of associated .txt file. Relative filepaths only.
	ChecksumPath string      // File directory of associated checksum. Relative filepaths only.
}

var (
	errTampered = errors.New("file tampering detected")
)

// Init initializes a HashLog's embedded log.Logger for writing to its associated log file with checking at its associated checksum file.
func Init(logname string) *HashLog {
	hl := &HashLog{ // Logs all login and logout attempts, account creation, modification and deletion.
		Name:         logname,
		L:            nil,
		LogPath:      fmt.Sprintf("./logs/%s.txt", logname),
		ChecksumPath: fmt.Sprintf("./logs/%sChecksum.txt", logname),
	}

	RecordFile, errLP := os.OpenFile(hl.LogPath,
		os.O_WRONLY|os.O_APPEND, 0666)
	if errLP != nil {
		RecordFile, _ = os.OpenFile(hl.LogPath,
			os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	}
	// defer RecordFile.Close()

	_, errCS := os.OpenFile(hl.ChecksumPath,
		os.O_RDWR, 0666)
	if errCS != nil {
		os.OpenFile(hl.ChecksumPath,
			os.O_CREATE|os.O_RDWR, 0666)
	}
	// defer ChecksumFile.Close()

	hl.L = log.New(io.MultiWriter(RecordFile, os.Stdout),
		fmt.Sprintf("[LOG] %s: ", logname), log.Ldate|log.Ltime)
	if errLP != nil {
		hl.updateHash()
	} else {
		check := hl.checkHash()
		if check != nil {
			log.Fatal("Hashlog init error: ", check, hl.Name)
		}
	}
	return hl
}

// AddtoLog checks a log file's SHA256 checksum to ensure no tampering has occured, before writing a message to the log and saving its updated hash value.
func (hl *HashLog) AddLog(message string) error {
	err := hl.checkHash()
	if err != nil {
		log.Fatal(hl.Name, ": ", err)
		return err
	}
	hl.L.Println(message)
	hl.updateHash()
	return nil
}

// Checks the file at a HashLog's LogPath against its saved hash at ChecksumPath. Returns nil if hashes match up, else returns errTampered.
func (hl *HashLog) checkHash() error {
	logFile, err := ioutil.ReadFile(hl.LogPath)
	if err != nil {
		log.Fatal("Failed to open user log file:", err)
		return err
	}
	// Compute current log's SHA256 hash
	b := sha256.Sum256(logFile)
	hash := string(b[:])
	// Read in saved checksum value
	checksumFile, err := ioutil.ReadFile(hl.ChecksumPath)
	if err != nil {
		log.Fatal("Failed to open checksum file:", err)
		return err
	}
	savedHash := string(checksumFile)
	// Compare hash values
	if hash != savedHash {
		return errTampered
	}
	return nil
}

// Take the file at a HashLog's LogPath, hash it, and save its hash at the HashLog's ChecksumPath.
func (hl *HashLog) updateHash() error {
	logFile, err := ioutil.ReadFile(hl.LogPath)
	if err != nil {
		log.Fatal("Failed to open user log file:", err)
		return err
	}
	// Compute current log's SHA256 hash
	b := sha256.Sum256(logFile)

	// Create file to store the hash, if it doesn't already exist.
	checksumFile, err := os.OpenFile(hl.ChecksumPath, os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		log.Fatal("Failed to open checksum file:", err)
		return err
	}
	// defer checksumFile.Close()
	checksumFile.Seek(0, 0) // Suppose a previous checksum already exists, rewind to prepare for overwrite.
	checksumFile.Write(b[:])
	return nil
}

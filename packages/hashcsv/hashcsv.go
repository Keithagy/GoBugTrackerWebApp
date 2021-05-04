// Implements funtions to initialize and read to / write from csv files using the csv package, with additional checks and error handling against an SHA256 checksum to detect file tampering.
// HashCSVs are safe for concurrent use via the inclusion of a Mutex with each struct.
// Also includes a file to track the last saved time and date of each CSV file.
// Functions to update the CSV save files are called whenever a user logs out of their session.
package hashcsv

import (
	"crypto/sha256"
	"encoding/csv"
	"errors"
	"fmt"
	"goInAction2/assignment/packages/dsa"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"sync"
	"time"
)

const (
	csvpath = "./csv/"
)

var (
	errTampered = errors.New("file tampering detected")
)

type HashCSV struct {
	Name          string
	FilePath      string
	ChecksumPath  string
	LastsavedPath string
	Reader        *csv.Reader
	Writer        *csv.Writer
	mu            sync.Mutex
}

// Init creates a new csv file as well as its associated hash, and returns a pointer to a HashCSV for that file.
func Init(name string) *HashCSV {
	filePath := fmt.Sprintf("%s%s.csv", csvpath, name)
	checksumPath := fmt.Sprintf("%s%schecksum.txt", csvpath, name)
	lastsavedPath := fmt.Sprintf("%s%sLS.txt", csvpath, name)
	csvFile, errFile := os.OpenFile(filePath,
		os.O_RDWR, 0666)
	if errFile != nil {
		csvFile, _ = os.OpenFile(filePath,
			os.O_CREATE|os.O_RDWR, 0666)
	}
	// defer csvFile.Close()
	_, errCS := os.OpenFile(checksumPath,
		os.O_RDWR, 0666)
	if errCS != nil {
		os.OpenFile(checksumPath,
			os.O_CREATE|os.O_RDWR, 0666)
	}

	// defer checksumFile.Close()

	lastsavedFile, err := os.OpenFile(lastsavedPath,
		os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		log.Fatal("Failed to open lastsaved file:", err, name)
	}
	// defer lastsavedFile.Close()

	hcsvReader := csv.NewReader(csvFile)
	hcsvWriter := csv.NewWriter(csvFile)
	lastsavedFile.Seek(0, 0)
	lastsavedFile.Write([]byte(time.Now().Format(time.RFC3339)))
	result := &HashCSV{
		Name:          name,
		FilePath:      filePath,
		ChecksumPath:  checksumPath,
		LastsavedPath: lastsavedPath,
		Reader:        hcsvReader,
		Writer:        hcsvWriter,
	}

	if errFile != nil {
		result.updateHash()
	} else {
		check := result.checkHash()
		if check != nil {
			log.Fatal("HashCSV init error: ", check, result.Name)
		}
	}
	return result
}

// SaveSubmissions saves an existing submissions heap (implemented in the dsa package) to an existing csv file, overwriting any existing data in the file, and updates the associated hash.
func (hcsv *HashCSV) SaveSubmissions(submissions *[]dsa.Ticket) {
	hcsv.mu.Lock()
	defer hcsv.mu.Unlock()
	// Check the hash
	err := hcsv.checkHash()
	if err != nil {
		log.Fatal("Checkhash failed: ", err, hcsv.Name)
	}

	// Remove existing version of the log and update HashCSV fields
	os.Remove(hcsv.FilePath)
	updatedCSV, err := os.OpenFile(hcsv.FilePath,
		os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		log.Fatal("Failed to open csv file: ", err, hcsv.Name)
	}
	// defer updatedCSV.Close()

	hcsv.Reader = csv.NewReader(updatedCSV)
	hcsv.Writer = csv.NewWriter(updatedCSV)

	// Write to updatedCSV
	var records [][]string
	for _, ticket := range *submissions {
		records = append(records, []string{
			fmt.Sprint(ticket.TicketID),
			fmt.Sprint(ticket.Product),
			fmt.Sprint(ticket.Status),
			fmt.Sprint(ticket.Category),
			fmt.Sprint(ticket.Priority),
			fmt.Sprint(ticket.EstHours),
			ticket.StartDate.Format(time.RFC3339),
			ticket.DueDate.Format(time.RFC3339),
			ticket.Creator,
			ticket.Title,
			ticket.Description,
			ticket.Assignee,
		})
	}
	err = (hcsv.Writer).WriteAll(records)
	if err != nil {
		log.Fatal("Failed to write record to csv file: ", err, hcsv.Name)
	}

	// Update hash checksum and last saved
	err = hcsv.updateHash()
	if err != nil {
		log.Fatal("Failed to update hash: ", err, hcsv.Name)
	}
	err = hcsv.updateLastSaved()
	if err != nil {
		log.Fatal("Failed to update last saved: ", err, hcsv.Name)
	}
}

// LoadSubmissions loads a submissions heap (implemented in the dsa package) from an existing csv file, and returns that newly-loaded heap's address.
func (hcsv *HashCSV) LoadSubmissions() *[]dsa.Ticket {
	hcsv.mu.Lock()
	defer hcsv.mu.Unlock()
	// Check the hash
	err := hcsv.checkHash()
	if err != nil {
		log.Fatal("Checkhash failed: ", err, hcsv.Name)
	}
	// Read from the file
	records, err := (hcsv.Reader).ReadAll()
	if err != nil {
		log.Fatal("Failed to read csv file: ", err, hcsv.Name)
	}

	submissions := make([]dsa.Ticket, 0)
	for _, ticket := range records {
		rebuilt := rebuildTicket(ticket)
		dsa.Addsubmission(&submissions, rebuilt)
	}
	return &submissions
}

// SaveTickets saves an existing ticket AVL tree (implemented in the dsa package) to an existing csv file, overwriting any existing data in the file, and updates the associated hash.
func (hcsv *HashCSV) SaveTickets(Tickets *dsa.AVLtree) {
	hcsv.mu.Lock()
	defer hcsv.mu.Unlock()
	// Check the hash
	err := hcsv.checkHash()
	if err != nil {
		log.Fatal("Checkhash failed: ", err, hcsv.Name)
	}
	// Remove existing version of the log and update HashCSV fields
	os.Remove(hcsv.FilePath)
	updatedCSV, err := os.OpenFile(hcsv.FilePath,
		os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		log.Fatal("Failed to open csv file: ", err, hcsv.Name)
	}
	// defer updatedCSV.Close()
	hcsv.Reader = csv.NewReader(updatedCSV)
	hcsv.Writer = csv.NewWriter(updatedCSV)

	// Write to updatedCSV
	records := saveAVLTree(Tickets.Root, [][]string{})
	err = (hcsv.Writer).WriteAll(records)
	if err != nil {
		log.Fatal("Failed to write record to csv file: ", err, hcsv.Name)
	}

	// Update hash checksum and last saved
	err = hcsv.updateHash()
	if err != nil {
		log.Fatal("Failed to update hash: ", err, hcsv.Name)
	}
	err = hcsv.updateLastSaved()
	if err != nil {
		log.Fatal("Failed to update last saved: ", err, hcsv.Name)
	}
}

// LoadTickets loads a tickets AVL tree (implemented in the dsa package) from an existing csv file, and returns that newly-loaded AVL tree's address.
func (hcsv *HashCSV) LoadTickets() *dsa.TicketNode {
	hcsv.mu.Lock()
	defer hcsv.mu.Unlock()
	// Check the hash
	err := hcsv.checkHash()
	if err != nil {
		log.Fatal("Checkhash failed: ", err, hcsv.Name)
	}
	// Read from the file
	records, err := (hcsv.Reader).ReadAll()
	if err != nil {
		log.Fatal("Failed to read csv file: ", err, hcsv.Name)
	}

	var tickets *dsa.TicketNode
	for _, ticket := range records {
		rebuilt := rebuildTicket(ticket)
		tickets = dsa.AVLinsert(&dsa.TicketNode{
			Ticket: rebuilt,
			Height: 0,
			Left:   nil,
			Right:  nil,
		}, dsa.ByTicketID, tickets)
	}
	return tickets
}

// SaveProducts saves a products slice to an existing csv file, overwriting any existing data in the file, and updates the associated hash.
func (hcsv *HashCSV) SaveProducts(products *[]string) {
	hcsv.mu.Lock()
	defer hcsv.mu.Unlock()
	// Check the hash
	err := hcsv.checkHash()
	if err != nil {
		log.Fatal("Checkhash failed: ", err, hcsv.Name)
	}
	// Remove existing version of the log and update HashCSV fields
	os.Remove(hcsv.FilePath)
	updatedCSV, err := os.OpenFile(hcsv.FilePath,
		os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		log.Fatal("Failed to open csv file: ", err, hcsv.Name)
	}
	// defer updatedCSV.Close()
	hcsv.Reader = csv.NewReader(updatedCSV)
	hcsv.Writer = csv.NewWriter(updatedCSV)

	// Write to updatedCSV
	records := make([][]string, 0)
	for _, product := range *products {
		records = append(records, []string{product})
	}
	err = (hcsv.Writer).WriteAll(records)
	if err != nil {
		log.Fatal("Failed to write record to csv file: ", err, hcsv.Name)
	}

	// Update hash checksum and last saved
	err = hcsv.updateHash()
	if err != nil {
		log.Fatal("Failed to update hash: ", err, hcsv.Name)
	}
	err = hcsv.updateLastSaved()
	if err != nil {
		log.Fatal("Failed to update last saved: ", err, hcsv.Name)
	}
}

// LoadProducts loads a products slice from an existing csv file, and returns that newly-loaded products slice's address.
func (hcsv *HashCSV) LoadProducts() *[]string {
	hcsv.mu.Lock()
	defer hcsv.mu.Unlock()
	// Check the hash
	err := hcsv.checkHash()
	if err != nil {
		log.Fatal("Checkhash failed: ", err, hcsv.Name)
	}
	// Read from the file
	records, err := (hcsv.Reader).ReadAll()
	if err != nil {
		log.Fatal("Failed to read csv file: ", err, hcsv.Name)
	}

	products := make([]string, 0)
	for _, product := range records {
		products = append(products, product[0])
	}
	return &products
}

// SaveUsers saves a users hash table to an existing csv file, overwriting any existing data in the file, and updates the associated hash.
func (hcsv *HashCSV) SaveUsers(users *[]*dsa.UserNode) {
	hcsv.mu.Lock()
	defer hcsv.mu.Unlock()
	// Check the hash
	err := hcsv.checkHash()
	if err != nil {
		log.Fatal("Checkhash failed: ", err, hcsv.Name)
	}
	// Remove existing version of the log and update HashCSV fields
	os.Remove(hcsv.FilePath)
	updatedCSV, err := os.OpenFile(hcsv.FilePath,
		os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		log.Fatal("Failed to open csv file: ", err, hcsv.Name)
	}
	// defer updatedCSV.Close()
	hcsv.Reader = csv.NewReader(updatedCSV)
	hcsv.Writer = csv.NewWriter(updatedCSV)

	// Write to updatedCSV
	printfunc := func(SLL *dsa.UserNode) []string {
		result := make([]string, 0)
		result = append(result, SLL.User.Name)
		result = append(result, string(SLL.User.Pw))
		result = append(result, strconv.FormatBool(SLL.User.Admin))
		return result
	}
	records := dsa.PrintHT(users, printfunc)

	err = (hcsv.Writer).WriteAll(records)
	if err != nil {
		log.Fatal("Failed to write record to csv file: ", err, hcsv.Name)
	}

	// Update hash checksum and last saved
	err = hcsv.updateHash()
	if err != nil {
		log.Fatal("Failed to update hash: ", err, hcsv.Name)
	}
	err = hcsv.updateLastSaved()
	if err != nil {
		log.Fatal("Failed to update last saved: ", err)
	}
}

// LoadUsers loads a users hash table (implemented in the dsa package) from an existing csv file, and returns that newly-loaded users hash table's address.
func (hcsv *HashCSV) LoadUsers() *[]*dsa.UserNode {
	hcsv.mu.Lock()
	defer hcsv.mu.Unlock()
	// Check the hash
	err := hcsv.checkHash()
	if err != nil {
		log.Fatal("Checkhash failed: ", err, hcsv.Name)
	}
	// Read from the file
	records, err := (hcsv.Reader).ReadAll()
	if err != nil {
		log.Fatal("Failed to read csv file: ", err, hcsv.Name)
	}

	users := make([]*dsa.UserNode, dsa.Hashbuckets)
	for _, record := range records {
		admin, _ := strconv.ParseBool(record[2])
		user := dsa.User{
			Name:  record[0],
			Pw:    []byte(record[1]),
			Admin: admin,
		}
		dsa.AddUser(&users, user)
	}
	return &users
}

// Check the SHA256 hash of a CSV file against its associated checksum file.
func (hcsv *HashCSV) checkHash() error {
	csvFile, err := ioutil.ReadFile(hcsv.FilePath)
	if err != nil {
		log.Fatal("Failed to open csv file: ", err, hcsv.Name)
		return err
	}
	// Compute current log's SHA256 hash
	b := sha256.Sum256(csvFile)
	hash := string(b[:])
	// Read in saved checksum value
	checksumFile, err := ioutil.ReadFile(hcsv.ChecksumPath)
	if err != nil {
		log.Fatal("Failed to open checksum file: ", err, hcsv.Name)
		return err
	}
	savedHash := string(checksumFile)
	// Compare hash values
	if hash != savedHash {
		return errTampered
	}
	return nil
}

// Take the file at a HashCSV's FilePath, hash it, and save its hash at the HashCSV's ChecksumPath.
func (hcsv *HashCSV) updateHash() error {
	csvFile, err := ioutil.ReadFile(hcsv.FilePath)
	if err != nil {
		log.Fatal("Failed to open csv file: ", err, hcsv.Name)
		return err
	}
	// Compute current log's SHA256 hash
	b := sha256.Sum256(csvFile)

	// Create file to store the hash, if it doesn't already exist.
	checksumFile, err := os.OpenFile(hcsv.ChecksumPath, os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		log.Fatal("Failed to open checksum file: ", err, hcsv.Name)
		return err
	}
	// defer checksumFile.Close()
	checksumFile.Seek(0, 0) // Suppose a previous checksum already exists, rewind to prepare for overwrite.
	checksumFile.Write(b[:])
	return nil
}

// Update HashCSV's lastsaved file.
func (hcsv *HashCSV) updateLastSaved() error {
	// Create file to store the lastsaved, if it doesn't already exist.
	lastsavedFile, err := os.OpenFile(hcsv.LastsavedPath, os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		log.Fatal("Failed to open lastsaved file: ", err, hcsv.Name)
		return err
	}
	// defer lastsavedFile.Close()
	lastsavedFile.Seek(0, 0) // Suppose a previous checksum already exists, rewind to prepare for overwrite.
	lastsavedFile.Write([]byte(time.Now().Format(time.RFC3339)))
	return nil
}

// Prepares AVL tree for saving into csv file.
func saveAVLTree(avlroot *dsa.TicketNode, result [][]string) [][]string {
	if avlroot == nil {
		return result
	}
	result = saveAVLTree(avlroot.Left, result)
	result = append(result, []string{
		fmt.Sprint(avlroot.Ticket.TicketID),
		fmt.Sprint(avlroot.Ticket.Product),
		fmt.Sprint(avlroot.Ticket.Status),
		fmt.Sprint(avlroot.Ticket.Category),
		fmt.Sprint(avlroot.Ticket.Priority),
		fmt.Sprint(avlroot.Ticket.EstHours),
		avlroot.Ticket.StartDate.Format(time.RFC3339),
		avlroot.Ticket.DueDate.Format(time.RFC3339),
		avlroot.Ticket.Creator,
		avlroot.Ticket.Title,
		avlroot.Ticket.Description,
		avlroot.Ticket.Assignee,
	})
	result = saveAVLTree(avlroot.Right, result)
	return result
}

// Recomposes the ticket from the record ([]string) read from a CSV file.
func rebuildTicket(ticket []string) dsa.Ticket {
	ticketid, _ := strconv.ParseInt(ticket[0], 10, 64)
	product, _ := strconv.Atoi(ticket[1])
	status, _ := strconv.Atoi(ticket[2])
	category, _ := strconv.Atoi(ticket[3])
	priority, _ := strconv.Atoi(ticket[4])
	esthours, _ := strconv.Atoi(ticket[5])
	startdate, _ := time.Parse(time.RFC3339, ticket[6])
	duedate, _ := time.Parse(time.RFC3339, ticket[7])
	creator := ticket[8]
	title := ticket[9]
	desc := ticket[10]
	assignee := ticket[11]

	return dsa.Ticket{
		TicketID:    ticketid,
		Product:     product,
		Status:      status,
		Category:    category,
		EstHours:    esthours,
		Priority:    priority,
		StartDate:   startdate,
		DueDate:     duedate,
		Creator:     creator,
		Title:       title,
		Description: desc,
		Assignee:    assignee}
}

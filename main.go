package main

import (
	"errors"
	"fmt"
	"goInAction2/assignment/packages/dsa"
	"goInAction2/assignment/packages/hashcsv"
	"goInAction2/assignment/packages/hashlog"
	"goInAction2/assignment/packages/test"
	"html/template"
	"log"
	"net/http"
	"sync"
)

// Used for pre-loading AVLtree pivots
type pivotItem struct {
	label   string
	pivoted *dsa.AVLtree
}

var (
	// Networking-related variables
	tpl         *template.Template
	mapSessions map[string]dsa.User

	// User-tracking variable
	users    *[]*dsa.UserNode
	loggedin dsa.User

	// Ticket-tracking
	ticketIDcounter int64
	submissions     *[]dsa.Ticket
	ticketlog       *dsa.AVLtree
	mytickets       *dsa.AVLtree
	myassigns       *dsa.AVLtree

	// Categories
	products   = &([]string{})
	statuses   = &([]string{"Not Started", "In Progress", "Paused"})
	categories = &([]string{"New feature", "Bug", "Enhancement"})
	priorities = &([]string{"High", "Medium", "Low"})

	// Concurrency
	testdataloader = make(chan test.Testdata)
	demodata       test.Testdata

	// Error Handling
	// errExisting signals that option inputted overlaps with an already-existing key
	errExisting = errors.New("input already exists -- please input unique value")
	// errBlank signals that a disallowed blank input was provided
	errBlank = errors.New("blank input not allowed -- please try again")
	// errInvalid signals that a disallowed blank input was provided
	errInvalid = errors.New("invalid input -- please try again")

	// Initialize Loggers
	userRecord       = hashlog.Init("UserRecord")
	generalRecord    = hashlog.Init("GeneralRecord")
	ticketRecord     = hashlog.Init("TicketRecord")
	submissionRecord = hashlog.Init("SubmissionRecord")

	// // Initialize Persistent Storage (CSV)
	submissionsCSV = hashcsv.Init("submissions")
	ticketsCSV     = hashcsv.Init("tickets")
	productsCSV    = hashcsv.Init("products")
	usersCSV       = hashcsv.Init("users")
)

func init() {
	// Parse HTML Templates
	tpl = template.Must(template.ParseGlob("templates/*"))

	// Initialize Data Structures
	mapSessions = make(map[string]dsa.User)
	users = dsa.NewHT()
	submissions = &[]dsa.Ticket{}
	ticketlog = dsa.NewAVLT(dsa.ByTicketID)

	// Read in data from persistent storage, if any
	submissions = submissionsCSV.LoadSubmissions()
	ticketlog.Root = ticketsCSV.LoadTickets()
	products = productsCSV.LoadProducts()
	users = usersCSV.LoadUsers()
}

func main() {
	msg := "Panic Trapped!"
	defer func() {
		if err := recover(); err != nil {
			generalRecord.AddLog(fmt.Sprintf("%s: %s", msg, err))
			generalRecord.AddLog("Resetting data structure states. Previous data not saved.")
			submissions = &[]dsa.Ticket{}
			users = dsa.NewHT()
			ticketlog = dsa.NewAVLT(dsa.ByTicketID)
			products = &([]string{})
			statuses = &([]string{"Not Started", "In Progress", "Paused"})
			categories = &([]string{"New feature", "Bug", "Enhancement"})
			priorities = &([]string{"High", "Medium", "Low"})
		} else {
			submissionsCSV.SaveSubmissions(submissions)
			ticketsCSV.SaveTickets(ticketlog)
			productsCSV.SaveProducts(products)
			usersCSV.SaveUsers(users)
			generalRecord.AddLog("Exited safely.")
		}
	}()

	// Concurrently pre-load demo mode data for population
	var wg sync.WaitGroup
	wg.Add(1)
	go test.Demomode(&ticketIDcounter, testdataloader, &wg)
	demodata = <-testdataloader

	// Populating the default multiplexer
	// Login Screen
	http.HandleFunc("/", index)
	http.Handle("/favicon.ico", http.NotFoundHandler())
	http.HandleFunc("/signup", signup)
	http.HandleFunc("/login", login)
	http.HandleFunc("/viewusers", viewusers)
	http.HandleFunc("/demo", demo)

	// Admin features
	http.HandleFunc("/adduser", adduser)
	http.HandleFunc("/edituser", edituser)
	http.HandleFunc("/deleteuser", deleteuser)
	http.HandleFunc("/manprods", manprods)
	http.HandleFunc("/addproducts", addproducts)
	http.HandleFunc("/editproducts", editproducts)
	http.HandleFunc("/deleteproducts", deleteproducts)
	http.HandleFunc("/managesubmissions", managesubmissions)

	// Non-Admin features
	http.HandleFunc("/submitticket", submitticket)
	http.HandleFunc("/submitted", submitted)
	http.HandleFunc("/viewmytickets", viewmytickets)
	http.HandleFunc("/deletemytickets", deletemytickets)
	http.HandleFunc("/viewmyassignments", viewmyassignments)
	http.HandleFunc("/markmyassignments", markmyassignments)
	http.HandleFunc("/viewalltickets", viewalltickets)
	http.HandleFunc("/ressorttickets", resorttickets)
	http.HandleFunc("/viewsubmissions", viewsubmissions)

	http.HandleFunc("/logout", logout)
	wg.Wait()
	err := http.ListenAndServeTLS(":8081", "./cert/cert.pem", "./cert/key.pem", nil)
	// err := http.ListenAndServe(":8081", nil)
	if err != nil {
		log.Fatal(err)
	}
}

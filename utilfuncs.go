// Utility functions called by handler functions.
package main

import (
	"fmt"
	"goInAction2/assignment/packages/dsa"
	"net/http"
	"time"

	uuid "github.com/satori/go.uuid"
	"golang.org/x/crypto/bcrypt"
)

// func keepAlive(res http.ResponseWriter, req *http.Request) {
// 	res.Header().Set("Content-Type", "text/event-stream")
// 	res.Header().Set("Cache-Control", "no-cache")
// 	res.Header().Set("Connection", "keep-alive")

// 	flusher, ok := res.(http.Flusher)
// 	if !ok {
// 		if alreadyLoggedIn(req) {
// 			generalRecord.AddLog(fmt.Sprintf("User %s's browser doesn't support server-sent events", loggedin.Name))
// 		} else {
// 			generalRecord.AddLog("User (not logged in)'s browser doesn't support server-sent events")
// 		}
// 	}

// 	// Send a comment every second to prevent connection timeout.
// 	for {
// 		_, err := fmt.Fprint(res, "")
// 		if err != nil {
// 			if alreadyLoggedIn(req) {
// 				submissionsCSV.SaveSubmissions(submissions)
// 				ticketsCSV.SaveTickets(ticketlog)
// 				generalRecord.AddLog(fmt.Sprintf("User %s closed their browser tab. Submissions and Ticketlog saved.", loggedin.Name))
// 			} else {
// 				generalRecord.AddLog("User (not logged in) closed their browser tab.")
// 			}
// 		}
// 		flusher.Flush()
// 		time.Sleep(time.Second)
// 	}
// }

func dltProductsHeap(input int, products *[]string, submissionsToDlt *[]int64, submissions *[]dsa.Ticket) *[]int64 {
	for i := range *submissions {
		compareval := (*submissions)[i].Product
		if compareval == input {
			*submissionsToDlt = append(*submissionsToDlt, (*submissions)[i].TicketID)
		} else if compareval > input {
			(*submissions)[i].Product = compareval - 1
		}
	}
	return submissionsToDlt
}

func dltProductsAVL(input int, products *[]string, ticketsToDlt *[]int64, avlroot *dsa.TicketNode) *[]int64 {

	if avlroot == nil {
		return ticketsToDlt
	}

	compareval := avlroot.Ticket.Product
	if compareval == input {
		*ticketsToDlt = append(*ticketsToDlt, avlroot.Ticket.TicketID)
	} else if compareval > input {
		avlroot.Ticket.Product = compareval - 1
	}

	ticketsToDlt = dltProductsAVL(input, products, ticketsToDlt, avlroot.Left)
	ticketsToDlt = dltProductsAVL(input, products, ticketsToDlt, avlroot.Right)

	return ticketsToDlt

}

func searchSlice(slice *[]string, input string) bool {
	var found bool
	for i := 0; i < len(*slice); i++ {
		if (*slice)[i] == input {
			found = true
		}
	}
	return found
}

// Creates a new account and creates an active session using the newly created account.
func newacc(res http.ResponseWriter, req *http.Request) (dsa.User, error) {
	var myUser dsa.User
	// process form submission
	if req.Method == http.MethodPost {
		// get form values
		usernameChan := make(chan string)
		passwordChan := make(chan string)
		repeatChan := make(chan string)
		adminrawChan := make(chan string)

		go func() {
			usernameChan <- req.FormValue("username")
		}()

		go func() {
			passwordChan <- req.FormValue("password")
		}()

		go func() {
			repeatChan <- req.FormValue("repeat")
		}()

		go func() {
			adminrawChan <- req.FormValue("admin")
		}()

		username := <-usernameChan
		password := <-passwordChan
		repeat := <-repeatChan
		adminraw := <-adminrawChan

		admin := false

		if adminraw != "" {
			admin = true
		}
		if username != "" && password != "" {
			// check if username exist/ taken
			if ok, _ := dsa.SearchUser(users, username); ok {
				userRecord.AddLog(fmt.Sprintf("Attempted account creation(non-admin), but username %s already taken.", username))
				http.Error(res, "Username already taken", http.StatusForbidden)
				return dsa.EmptyUser, errExisting
			}
			// validate password against repeat
			if password != repeat {
				userRecord.AddLog("Attempted account creation(non-admin), but passwords entered did not match.")
				http.Error(res, "Passwords entered do not match.", http.StatusForbidden)
				return dsa.EmptyUser, errInvalid
			}
			// create session
			id := uuid.NewV4()
			myCookie := &http.Cookie{
				Name:  "myCookie",
				Value: id.String(),
			}
			http.SetCookie(res, myCookie)
			generalRecord.AddLog("MyCookie: New session created (Non-admin signup page).")

			bPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
			if err != nil {
				http.Error(res, "Internal server error", http.StatusInternalServerError)
				return dsa.EmptyUser, errInvalid
			}

			myUser = dsa.User{
				Name:  username,
				Pw:    bPassword,
				Admin: admin}
			mapSessions[myCookie.Value] = myUser
			dsa.AddUser(users, myUser)
			userRecord.AddLog(fmt.Sprintf("Successful account creation(non-admin). Username: %s, Admin: %t.", username, admin))
		} else {
			userRecord.AddLog("Attempted account creation(non-admin), but blank username and/or password entered.")
			http.Error(res, errBlank.Error(), http.StatusForbidden)
			return dsa.EmptyUser, errBlank
		}
		// redirect to main index
		http.Redirect(res, req, "/", http.StatusSeeOther)
		return myUser, nil
	}
	return dsa.EmptyUser, nil
}

// Creates a new account without switching the active session. Called by admin when creating new account.
func newaccadmin(res http.ResponseWriter, req *http.Request) (dsa.User, error) {
	var myUser dsa.User
	// process form submission
	if req.Method == http.MethodPost {
		// get form values
		username := req.FormValue("username")
		password := req.FormValue("password")
		repeat := req.FormValue("repeat")
		adminraw := req.FormValue("admin")
		admin := false
		if adminraw != "" {
			admin = true
		}
		if username != "" && password != "" {
			// check if username exist/ taken
			if ok, _ := dsa.SearchUser(users, username); ok {
				userRecord.AddLog(fmt.Sprintf("Attempted account creation(admin), but username %s already taken.", username))
				http.Error(res, "Username already taken", http.StatusForbidden)
				return dsa.EmptyUser, errExisting
			}

			// validate password against repeat
			if password != repeat {
				userRecord.AddLog("Attempted account creation(admin), but passwords entered did not match.")
				http.Error(res, "Passwords entered do not match.", http.StatusForbidden)
				return dsa.EmptyUser, errInvalid
			}

			bPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
			if err != nil {
				http.Error(res, "Internal server error", http.StatusInternalServerError)
				return dsa.EmptyUser, errInvalid
			}

			myUser = dsa.User{
				Name:  username,
				Pw:    bPassword,
				Admin: admin}
			dsa.AddUser(users, myUser)
			userRecord.AddLog(fmt.Sprintf("Successful account creation(admin). Username: %s, Admin: %t.", username, admin))
		} else {
			userRecord.AddLog("Attempted account creation(admin), but blank username and/or password entered.")
			http.Error(res, errBlank.Error(), http.StatusForbidden)
			return dsa.EmptyUser, errBlank
		}
		// redirect to main index
		http.Redirect(res, req, "/", http.StatusSeeOther)
		return myUser, nil
	}
	return dsa.EmptyUser, nil
}

func getUser(res http.ResponseWriter, req *http.Request) dsa.User {
	// get current session cookie
	myCookie, err := req.Cookie("myCookie")
	if err != nil {
		id := uuid.NewV4()
		myCookie = &http.Cookie{
			Name:  "myCookie",
			Value: id.String(),
		}
		generalRecord.AddLog("MyCookie: New session created.")
	}
	http.SetCookie(res, myCookie)

	// if the user exists already, get user
	var myUserNode *dsa.UserNode
	if user, ok := mapSessions[myCookie.Value]; ok {
		_, myUserNode = dsa.SearchUser(users, user.Name)
	}

	if myUserNode == nil {
		return dsa.EmptyUser
	}
	return myUserNode.User
}

func alreadyLoggedIn(req *http.Request) bool {
	myCookie, err := req.Cookie("myCookie")
	if err != nil {
		return false
	}
	user := mapSessions[myCookie.Value]
	ok, _ := dsa.SearchUser(users, user.Name)
	return ok
}

func newSubmission(title, desc, creator, assignee string,
	startdate, duedate time.Time,
	priority, dueyears, duemonths, duedays, product, status, category, esthours int,
	priorities, products, statuses, categories *[]string,
	ticketID int64) <-chan dsa.Ticket {
	newticket := make(chan dsa.Ticket)
	go func() {
		newticket <- dsa.Newticket(
			ticketID, product, status, category, priority, esthours,
			startdate, duedate,
			creator, title, desc, assignee,
			priorities, products, statuses, categories)
	}()
	return newticket
}

func preloadPivots(index int, label string) pivotItem {
	pivoted := &dsa.AVLtree{
		Root:     nil,
		Sortfunc: dsa.ByTicketID,
	}
	switch index {
	case 0: // Product
		pivoted.Sortfunc = dsa.ByProduct
		pivoted.Root = dsa.AVLpivot(ticketlog.Root, pivoted.Root, pivoted.Sortfunc)
	case 1: // Status
		pivoted.Sortfunc = dsa.ByStatus
		pivoted.Root = dsa.AVLpivot(ticketlog.Root, pivoted.Root, pivoted.Sortfunc)
	case 2: // Category
		pivoted.Sortfunc = dsa.ByCategory
		pivoted.Root = dsa.AVLpivot(ticketlog.Root, pivoted.Root, pivoted.Sortfunc)
	case 3: // Estimated Hours to Complete
		pivoted.Sortfunc = dsa.ByEstHours
		pivoted.Root = dsa.AVLpivot(ticketlog.Root, pivoted.Root, pivoted.Sortfunc)
	case 4: // Priority
		pivoted.Sortfunc = dsa.ByPriority
		pivoted.Root = dsa.AVLpivot(ticketlog.Root, pivoted.Root, pivoted.Sortfunc)
	case 5: // Start Date
		pivoted.Sortfunc = dsa.ByStartDate
		pivoted.Root = dsa.AVLpivot(ticketlog.Root, pivoted.Root, pivoted.Sortfunc)
	case 6: // Due Date
		pivoted.Sortfunc = dsa.ByDueDate
		pivoted.Root = dsa.AVLpivot(ticketlog.Root, pivoted.Root, pivoted.Sortfunc)
	case 7: // Creator
		pivoted.Sortfunc = dsa.ByCreator
		pivoted.Root = dsa.AVLpivot(ticketlog.Root, pivoted.Root, pivoted.Sortfunc)
	case 8: // Assignee
		pivoted.Sortfunc = dsa.ByAssignee
		pivoted.Root = dsa.AVLpivot(ticketlog.Root, pivoted.Root, pivoted.Sortfunc)
	case 9: // Title
		pivoted.Sortfunc = dsa.ByTitle
		pivoted.Root = dsa.AVLpivot(ticketlog.Root, pivoted.Root, pivoted.Sortfunc)
	case 10: // Description
		pivoted.Sortfunc = dsa.ByDescription
		pivoted.Root = dsa.AVLpivot(ticketlog.Root, pivoted.Root, pivoted.Sortfunc)
	}
	return pivotItem{label, pivoted}
}

func passPreload(options []string) <-chan pivotItem {
	c := make(chan pivotItem)
	go func() {
		for index, label := range options {
			c <- preloadPivots(index, label)
		}
	}()
	return c
}

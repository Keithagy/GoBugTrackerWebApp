// Handler functions used to serve client requests.

package main

import (
	"errors"
	"fmt"
	"goInAction2/assignment/packages/dsa"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	uuid "github.com/satori/go.uuid"
	"golang.org/x/crypto/bcrypt"
)

// Main Menu

func index(res http.ResponseWriter, req *http.Request) {

	var wg sync.WaitGroup
	checkuser := getUser(res, req)
	myCookie, _ := req.Cookie("myCookie")
	if myCookie == nil {
		http.Redirect(res, req, "/", http.StatusSeeOther)
		return
	} else {
		loggedin = mapSessions[myCookie.Value]
	}

	if alreadyLoggedIn(req) {
		generalRecord.AddLog(fmt.Sprintf("Username %s (Admin: %t) accessed main menu.", loggedin.Name, loggedin.Admin))
	} else {
		generalRecord.AddLog("New User (not logged in) accessed main menu.")
	}

	if alreadyLoggedIn(req) && !loggedin.Admin {
		wg.Add(1)
		go func() {
			defer wg.Done()
			mytickets = &dsa.AVLtree{
				Root:     nil,
				Sortfunc: dsa.ByTicketID,
			}
			mytickets.Root = dsa.Mytickets(ticketlog.Root, mytickets.Root, dsa.ByTicketID, loggedin.Name)
		}()

		wg.Add(1)
		go func() {
			defer wg.Done()
			myassigns = &dsa.AVLtree{
				Root:     nil,
				Sortfunc: dsa.ByTicketID,
			}
			myassigns.Root = dsa.Myassigns(ticketlog.Root, myassigns.Root, dsa.ByTicketID, loggedin.Name)
		}()
	}
	tpl.ExecuteTemplate(res, "index.gohtml", checkuser)
	wg.Wait()
}

// Login Screen

func signup(res http.ResponseWriter, req *http.Request) {

	if alreadyLoggedIn(req) {
		generalRecord.AddLog(fmt.Sprintf("Username %s (Admin: %t) accessed sign up. Redirected to main menu.", loggedin.Name, loggedin.Admin))
	} else {
		generalRecord.AddLog("New User (not logged in) accessed sign up page.")
	}

	if alreadyLoggedIn(req) {
		http.Redirect(res, req, "/", http.StatusSeeOther)
		return
	}

	myUser, err := newacc(res, req)
	if err == nil {
		tpl.ExecuteTemplate(res, "signup.gohtml", myUser)
	}
}

func login(res http.ResponseWriter, req *http.Request) {

	if alreadyLoggedIn(req) {
		http.Redirect(res, req, "/", http.StatusSeeOther)
		return
	}
	// process form submission
	if req.Method == http.MethodPost {

		// usernameChan := make(chan string)

		// go func() {
		// usernameChan <- req.FormValue("username")
		// }()

		// go func() {
		// passwordChan <- req.FormValue("password")
		// }()

		// username := <-usernameChan
		// password := <-passwordChan

		username := req.FormValue("username")

		passwordChan := make(chan string)
		go func() {
			passwordChan <- req.FormValue("password")
		}()

		password := <-passwordChan

		// check if user exist with username
		ok, myUserNode := dsa.SearchUser(users, username)
		if !ok {
			userRecord.AddLog(fmt.Sprintf("Sign-in attempt using username %s (invalid username).", username))
			http.Error(res, "Username and/or password do not match", http.StatusUnauthorized)
			return
		}
		// check if user already has existing session, i.e logged in already
		for _, user := range mapSessions {
			if user.Name == username {
				userRecord.AddLog(fmt.Sprintf("Sign-in attempt using username %s (already logged in).", username))
				http.Error(res, "Inputted user already logged in!", http.StatusUnauthorized)
				return
			}
		}
		// Matching of password entered
		err := bcrypt.CompareHashAndPassword(myUserNode.User.Pw, []byte(password))
		if err != nil {
			userRecord.AddLog(fmt.Sprintf("Sign-in attempt using username %s (wrong password).", username))
			http.Error(res, "Username and/or password do not match", http.StatusForbidden)
			return
		}
		// create session
		id := uuid.NewV4()
		myCookie := &http.Cookie{
			Name:  "myCookie",
			Value: id.String(),
		}
		http.SetCookie(res, myCookie)
		mapSessions[myCookie.Value] = myUserNode.User
		generalRecord.AddLog(fmt.Sprintf("MyCookie: New session created for username %s.", myUserNode.User.Name))
		userRecord.AddLog(fmt.Sprintf("Successful sign-in using username %s (Admin: %t).", myUserNode.User.Name, myUserNode.User.Admin))
		http.Redirect(res, req, "/", http.StatusSeeOther)
		return
	}
	tpl.ExecuteTemplate(res, "login.gohtml", nil)
}

func viewusers(res http.ResponseWriter, req *http.Request) {

	str := dsa.PrintHT(users, dsa.PrintSLLusername)
	tpl.ExecuteTemplate(res, "viewusers.gohtml", str)
}

func demo(res http.ResponseWriter, req *http.Request) {

	generalRecord.AddLog("Demo mode activated, test state populated containing:")
	ticketRecord.AddLog("Demomode, 4 tickets loaded into ticket log.")
	submissionRecord.AddLog("Demomode, 4 tickets loaded into submissions priority queue.")
	userRecord.AddLog("Demomode, 3 users added to hash table.")
	userRecord.AddLog("Username: admin")
	userRecord.AddLog("Admin account")
	userRecord.AddLog("Username: user1")
	userRecord.AddLog("Non-Admin account")
	userRecord.AddLog("Username: user2")
	userRecord.AddLog("Non-Admin account")

	submissions = demodata.Testsubs
	users = demodata.Testusers
	ticketlog = demodata.Testticketlog
	products = demodata.Testproducts

	tpl.ExecuteTemplate(res, "demo.gohtml", nil)
}

// Admin Features

func adduser(res http.ResponseWriter, req *http.Request) {

	if !alreadyLoggedIn(req) {
		generalRecord.AddLog("New User (not logged in) accessed admin add user. Redirected to main menu.")
	} else if !loggedin.Admin {
		generalRecord.AddLog(fmt.Sprintf("Username %s (Admin: %t) accessed admin add user. Redirected to main menu.", loggedin.Name, loggedin.Admin))
	} else {
		generalRecord.AddLog(fmt.Sprintf("Username %s (Admin: %t) accessed admin add user.", loggedin.Name, loggedin.Admin))
	}

	if !alreadyLoggedIn(req) || !loggedin.Admin {
		http.Redirect(res, req, "/", http.StatusSeeOther)
		return
	}
	currsesh, _ := req.Cookie("myCookie")
	_, err := newaccadmin(res, req)
	if err == nil {
		tpl.ExecuteTemplate(res, "signup.gohtml", mapSessions[currsesh.Value])
	}
}

func edituser(res http.ResponseWriter, req *http.Request) {

	if !alreadyLoggedIn(req) {
		generalRecord.AddLog("New User (not logged in) accessed admin edit user. Redirected to main menu.")
	} else if !loggedin.Admin {
		generalRecord.AddLog(fmt.Sprintf("Username %s (Admin: %t) accessed admin edit user. Redirected to main menu.", loggedin.Name, loggedin.Admin))
	} else {
		generalRecord.AddLog(fmt.Sprintf("Username %s (Admin: %t) accessed admin edit user.", loggedin.Name, loggedin.Admin))
	}

	if !alreadyLoggedIn(req) || !loggedin.Admin {
		http.Redirect(res, req, "/", http.StatusSeeOther)
		return
	}
	var retrieved, edited dsa.User
	var newname, newpw string
	str := dsa.PrintHT(users, dsa.PrintSLLusername)

	// Process form submission
	if req.Method == http.MethodPost {
		userchoice, _ := strconv.Atoi(req.FormValue("account"))
		username := strings.Fields(str[userchoice][0])[1]
		_, myUserNode := dsa.SearchUser(users, username)
		retrieved = myUserNode.User

		newname = req.FormValue("username")
		newpw = req.FormValue("password")

		edited = retrieved

		if newname != "" {
			if exists, _ := dsa.SearchUser(users, newname); exists {
				userRecord.AddLog(fmt.Sprintf("Admin User %s attempted user editing, but new username %s matches existing account.", loggedin.Name, newname))
				http.Error(res, "New username cannot be identical to existing account.", http.StatusUnauthorized)
				return
			}
			edited.Name = newname
			userRecord.AddLog(fmt.Sprintf("Admin user %s changed username %s to %s.", loggedin.Name, retrieved.Name, edited.Name))
		}
		if newpw != "" {
			input, _ := bcrypt.GenerateFromPassword([]byte(newpw), bcrypt.MinCost)
			edited.Pw = input
			userRecord.AddLog(fmt.Sprintf("Admin user %s changed password for username %s.", loggedin.Name, edited.Name))
		}
	}

	// Add form submission to data structure and exit to main menu
	if retrieved.Name != "" {
		dsa.EditUser(users, retrieved, edited)
		userRecord.AddLog("Hash table updated with edited account.")
		http.Redirect(res, req, "/", http.StatusSeeOther)
	}
	if currsesh, _ := req.Cookie("myCookie"); retrieved.Name == (mapSessions[currsesh.Value]).Name {
		mapSessions[currsesh.Value] = edited
	}
	tpl.ExecuteTemplate(res, "edituser.gohtml", str)
}

func deleteuser(res http.ResponseWriter, req *http.Request) {

	if !alreadyLoggedIn(req) {
		generalRecord.AddLog("New User (not logged in) accessed admin delete user. Redirected to main menu.")
	} else if !loggedin.Admin {
		generalRecord.AddLog(fmt.Sprintf("Username %s (Admin: %t) accessed admin delete user. Redirected to main menu.", loggedin.Name, loggedin.Admin))
	} else {
		generalRecord.AddLog(fmt.Sprintf("Username %s (Admin: %t) accessed admin delete user.", loggedin.Name, loggedin.Admin))
	}

	if !alreadyLoggedIn(req) || !loggedin.Admin {
		http.Redirect(res, req, "/", http.StatusSeeOther)
		return
	}
	var todelete dsa.User
	str := dsa.PrintHT(users, dsa.PrintSLLusername)

	// Process form submission
	if req.Method == http.MethodPost {
		userchoice, _ := strconv.Atoi(req.FormValue("account"))
		username := strings.Fields(str[userchoice][0])[1]
		_, myUserNode := dsa.SearchUser(users, username)
		todelete = myUserNode.User
	}

	// Delete user from hash table, delete cookie from active sessions, exit to main menu
	if todelete.Name != "" {
		dsa.DeleteUser(users, todelete.Name)
		http.Redirect(res, req, "/", http.StatusSeeOther)
	}
	if currsesh, _ := req.Cookie("myCookie"); todelete.Name == (mapSessions[currsesh.Value]).Name {
		delete(mapSessions, currsesh.Value)
		userRecord.AddLog(fmt.Sprintf("Admin user %s deleted account %s from hash table.", loggedin.Name, todelete.Name))
	}
	tpl.ExecuteTemplate(res, "deleteuser.gohtml", str)
}

func manprods(res http.ResponseWriter, req *http.Request) {

	if !alreadyLoggedIn(req) {
		generalRecord.AddLog("New User (not logged in) accessed admin manage products. Redirected to main menu.")
	} else if !loggedin.Admin {
		generalRecord.AddLog(fmt.Sprintf("Username %s (Admin: %t) accessed admin manage products. Redirected to main menu.", loggedin.Name, loggedin.Admin))
	} else {
		generalRecord.AddLog(fmt.Sprintf("Username %s (Admin: %t) accessed admin manage products.", loggedin.Name, loggedin.Admin))
	}

	if !alreadyLoggedIn(req) || !loggedin.Admin {
		http.Redirect(res, req, "/", http.StatusSeeOther)
		return
	}
	tpl.ExecuteTemplate(res, "manprods.gohtml", products)
}

func addproducts(res http.ResponseWriter, req *http.Request) {

	if !alreadyLoggedIn(req) {
		generalRecord.AddLog("New User (not logged in) accessed admin add products. Redirected to main menu.")
	} else if !loggedin.Admin {
		generalRecord.AddLog(fmt.Sprintf("Username %s (Admin: %t) accessed admin add products. Redirected to main menu.", loggedin.Name, loggedin.Admin))
	} else {
		generalRecord.AddLog(fmt.Sprintf("Username %s (Admin: %t) accessed admin add products.", loggedin.Name, loggedin.Admin))
	}

	if !alreadyLoggedIn(req) || !loggedin.Admin {
		http.Redirect(res, req, "/", http.StatusSeeOther)
		return
	}
	if req.Method == http.MethodPost {
		newproduct := req.FormValue("productname")
		if unique := !searchSlice(products, newproduct); unique && newproduct != "" {
			*products = append((*products), newproduct)
		} else if !unique {
			http.Error(res, "Product Name must be unique.", http.StatusUnauthorized)
			return
		}
	}
	tpl.ExecuteTemplate(res, "addproducts.gohtml", *products)
}

func editproducts(res http.ResponseWriter, req *http.Request) {

	if !alreadyLoggedIn(req) {
		generalRecord.AddLog("New User (not logged in) accessed admin edit products. Redirected to main menu.")
	} else if !loggedin.Admin {
		generalRecord.AddLog(fmt.Sprintf("Username %s (Admin: %t) accessed admin edit products. Redirected to main menu.", loggedin.Name, loggedin.Admin))
	} else {
		generalRecord.AddLog(fmt.Sprintf("Username %s (Admin: %t) accessed admin edit products.", loggedin.Name, loggedin.Admin))
	}

	if !alreadyLoggedIn(req) || !loggedin.Admin {
		http.Redirect(res, req, "/", http.StatusSeeOther)
		return
	}
	// Process form submission
	if req.Method == http.MethodPost {
		editindex, _ := strconv.Atoi(req.FormValue("product"))
		newname := req.FormValue("newname")
		if editindex >= len(*products) {
			panic(errors.New("selection exceeds slice length"))
		} else {
			if newname != "" {
				if exists := searchSlice(products, newname); exists {
					http.Error(res, "New Name must be unique.", http.StatusUnauthorized)
					return
				}
				(*products)[editindex] = newname
				http.Redirect(res, req, "/manprods", http.StatusSeeOther)
			} else {
				http.Error(res, "New Name cannot be blank.", http.StatusUnauthorized)
				return
			}
		}
	}
	tpl.ExecuteTemplate(res, "editproducts.gohtml", *products)
}

func deleteproducts(res http.ResponseWriter, req *http.Request) {

	if !alreadyLoggedIn(req) {
		generalRecord.AddLog("New User (not logged in) accessed admin delete products. Redirected to main menu.")
	} else if !loggedin.Admin {
		generalRecord.AddLog(fmt.Sprintf("Username %s (Admin: %t) accessed admin delete products. Redirected to main menu.", loggedin.Name, loggedin.Admin))
	} else {
		generalRecord.AddLog(fmt.Sprintf("Username %s (Admin: %t) accessed admin delete products.", loggedin.Name, loggedin.Admin))
	}

	if !alreadyLoggedIn(req) || !loggedin.Admin {
		http.Redirect(res, req, "/", http.StatusSeeOther)
		return
	}

	// Process form submission
	if req.Method == http.MethodPost {
		dltname := req.FormValue("product")
		dltindex, _ := strconv.Atoi(dltname)
		if dltname != "" {
			if dltindex >= len(*products) {
				panic(errors.New("selection exceeds slice length"))
			} else {
				// Delete all products of that category in ticketlog
				ticketsToDlt := &[]int64{}
				ticketsToDlt = dltProductsAVL(dltindex, products, ticketsToDlt, ticketlog.Root)

				for index := range *ticketsToDlt {
					ticketlog.Root = dsa.AVLdelete(ticketlog.Root, (*ticketsToDlt)[index])
				}

				// Delete all products of that category in submissions
				submissionsToDlt := &[]int64{}
				submissionsToDlt = dltProductsHeap(dltindex, products, submissionsToDlt, submissions)

				for index := range *submissionsToDlt {
					_, heapindex := dsa.Searchsubmissions(submissions, (*submissionsToDlt)[index])
					copy((*submissions)[heapindex:], (*submissions)[heapindex+1:])
					(*submissions) = (*submissions)[:len(*submissions)-1]
				}
				dsa.Makeheap(submissions, 0)

				// Delete product element from products slice
				(*products)[dltindex] = ""
				copy((*products)[dltindex:], (*products)[dltindex+1:])
				(*products) = (*products)[:len(*products)-1]
			}
		} else {
			panic(errors.New("invalid selection"))
		}
	}
	tpl.ExecuteTemplate(res, "deleteproducts.gohtml", *products)
}

func managesubmissions(res http.ResponseWriter, req *http.Request) {

	if !alreadyLoggedIn(req) {
		generalRecord.AddLog("New User (not logged in) accessed admin manage products. Redirected to main menu.")
	} else if !loggedin.Admin {
		generalRecord.AddLog(fmt.Sprintf("Username %s (Admin: %t) accessed admin manage products. Redirected to main menu.", loggedin.Name, loggedin.Admin))
	} else {
		generalRecord.AddLog(fmt.Sprintf("Username %s (Admin: %t) accessed admin manage products.", loggedin.Name, loggedin.Admin))
	}

	if !alreadyLoggedIn(req) || !loggedin.Admin {
		http.Redirect(res, req, "/", http.StatusSeeOther)
		return
	}
	var s [][]string
	var popped dsa.Ticket
	var apprej string
	s = dsa.LOtraversal(submissions, priorities, products, statuses, categories)

	if req.Method == http.MethodPost {
		popped = dsa.Popsubmission(submissions)
		apprej = req.FormValue("apprej")

		if apprej == "Approve" {
			newticket := &dsa.TicketNode{
				Ticket: popped,
				Height: 0,
				Left:   nil,
				Right:  nil,
			}
			ticketlog.Root = dsa.AVLinsert(newticket, ticketlog.Sortfunc, ticketlog.Root)
			s = dsa.LOtraversal(submissions, priorities, products, statuses, categories)
			ticketRecord.AddLog(fmt.Sprintf("Admin user %s approved submission (ID %v).", loggedin.Name, newticket.Ticket.TicketID))
		} else {
			s = dsa.LOtraversal(submissions, priorities, products, statuses, categories)
			ticketRecord.AddLog(fmt.Sprintf("Admin user %s rejected submission (ID %v).", loggedin.Name, popped.TicketID))
		}
	}
	tpl.ExecuteTemplate(res, "managesubmissions.gohtml", s)
}

// Non-Admin Features

func submitticket(res http.ResponseWriter, req *http.Request) {
	if !alreadyLoggedIn(req) {
		http.Redirect(res, req, "/", http.StatusSeeOther)
		return
	}
	var titleChan, descChan, creatorChan, assigneeChan chan string
	var startdateChan chan time.Time
	var priorityChan, dueyearsChan, duemonthsChan, duedaysChan, productChan, statusChan, categoryChan, esthoursChan chan int
	var ticketIDChan chan int64
	var erryChan, errmChan, errdChan, errEHChan chan error

	titleChan = make(chan string)
	descChan = make(chan string)
	creatorChan = make(chan string)
	assigneeChan = make(chan string)

	startdateChan = make(chan time.Time)

	priorityChan = make(chan int)
	dueyearsChan = make(chan int)
	duemonthsChan = make(chan int)
	duedaysChan = make(chan int)
	productChan = make(chan int)
	statusChan = make(chan int)
	categoryChan = make(chan int)
	esthoursChan = make(chan int)

	ticketIDChan = make(chan int64)

	erryChan = make(chan error)
	errmChan = make(chan error)
	errdChan = make(chan error)
	errEHChan = make(chan error)

	var title, desc, creator, assignee string
	var startdate, duedate time.Time
	var priority, dueyears, duemonths, duedays, product, status, category, esthours int
	var ticketID int64
	var erry, errm, errd, errEH error

	var mu sync.Mutex

	if req.Method == http.MethodGet {
		go func() {
			startdateChan <- time.Now()
			// close(startdateChan)
		}()

		go func() {
			ticketIDChan <- ticketIDcounter
			// close(ticketIDChan)
		}()

		go func() {
			currsesh, _ := req.Cookie("myCookie")
			mu.Lock()
			creatorChan <- mapSessions[currsesh.Value].Name
			mu.Unlock()
			// close(creatorChan)
		}()

		ticketID = <-ticketIDChan
		startdate = <-startdateChan
		creator = <-creatorChan
	}

	if req.Method == http.MethodPost {

		go func() {
			startdateChan <- time.Now()
		}()

		go func() {
			ticketIDChan <- ticketIDcounter
		}()

		go func() {
			titleChan <- req.FormValue("title")
		}()

		go func() {
			descChan <- req.FormValue("desc")
		}()

		go func() {
			currsesh, _ := req.Cookie("myCookie")
			mu.Lock()
			creatorChan <- mapSessions[currsesh.Value].Name
			mu.Unlock()
		}()

		go func() {
			assigneeRaw := req.FormValue("assignee")
			assignee := (strings.Fields(assigneeRaw))[1]
			assigneeChan <- assignee
		}()

		go func() {
			esthours, errEH := strconv.Atoi(req.FormValue("esthours"))
			esthoursChan <- esthours

			errEHChan <- errEH
		}()

		go func() {
			dueyears, erry := strconv.Atoi(req.FormValue("dueyears"))
			dueyearsChan <- dueyears

			erryChan <- erry
		}()

		go func() {
			duemonths, errm := strconv.Atoi(req.FormValue("duemonths"))
			duemonthsChan <- duemonths

			errmChan <- errm
		}()

		go func() {
			duedays, errd := strconv.Atoi(req.FormValue("duedays"))
			duedaysChan <- duedays

			errdChan <- errd
		}()

		go func() {
			priority, _ := strconv.Atoi(req.FormValue("priority"))
			priorityChan <- priority
		}()

		go func() {
			product, _ := strconv.Atoi(req.FormValue("product"))
			productChan <- product
		}()

		go func() {
			status, _ := strconv.Atoi(req.FormValue("status"))
			statusChan <- status
		}()

		go func() {
			category, _ := strconv.Atoi(req.FormValue("category"))
			categoryChan <- category
		}()

		ticketID = <-ticketIDChan
		atomic.AddInt64(&ticketIDcounter, 1)

		title = <-titleChan
		desc = <-descChan
		creator = <-creatorChan
		assignee = <-assigneeChan
		startdate = <-startdateChan
		esthours = <-esthoursChan
		errEH = <-errEHChan
		dueyears = <-dueyearsChan
		erry = <-erryChan
		duemonths = <-duemonthsChan
		errm = <-errmChan
		duedays = <-duedaysChan
		errd = <-errdChan
		priority = <-priorityChan
		product = <-productChan
		status = <-statusChan
		category = <-categoryChan

		if title == "" {
			submissionRecord.AddLog(fmt.Sprintf("User %s attempted ticket submission with one or more invalid inputs.", loggedin.Name))
			http.Error(res, "Title cannot be empty.", http.StatusForbidden)
			return
		}

		if desc == "" {
			submissionRecord.AddLog(fmt.Sprintf("User %s attempted ticket submission with one or more invalid inputs.", loggedin.Name))
			http.Error(res, "Description cannot be empty.", http.StatusForbidden)
			return
		}

		if ((erry == nil) && (dueyears > 0)) && ((errm == nil) && (duemonths > 0)) && ((errd == nil) && (duedays > 0)) {
			duedate = startdate.AddDate(dueyears, duemonths, duedays)
		} else {
			submissionRecord.AddLog(fmt.Sprintf("User %s attempted ticket submission with one or more invalid inputs.", loggedin.Name))
			http.Error(res, "Invalid ticket duration.", http.StatusForbidden)
			return
		}

		if errEH != nil || esthours <= 0 {
			submissionRecord.AddLog(fmt.Sprintf("User %s attempted ticket submission with one or more invalid inputs.", loggedin.Name))
			http.Error(res, "Invalid estimated hours entry.", http.StatusForbidden)
			return
		}

		if title != "" {
			submissionRecord.AddLog(fmt.Sprintf("Ticket ID %v submitted by user %v.", ticketID, creator))
			newticket := newSubmission(title, desc, creator, assignee, startdate, duedate, priority, dueyears, duemonths, duedays, product, status, category, esthours, priorities, products, statuses, categories, ticketID)

			var mu sync.Mutex
			mu.Lock()
			dsa.Addsubmission(submissions, <-newticket)
			mu.Unlock()

			http.Redirect(res, req, "/submitted", http.StatusSeeOther)
			return
		}
	}
	str := dsa.PrintHT(users, dsa.PrintSLLnoadmin)
	data := struct {
		Loggedinuser string
		Users        [][]string
		Priorities   []string
		Startdate    time.Time
		Products     []string
		Statuses     []string
		Categories   []string
		TicketID     int64
	}{
		creator,
		str,
		*priorities,
		startdate,
		*products,
		*statuses,
		*categories,
		ticketID,
	}
	tpl.ExecuteTemplate(res, "submitticket.gohtml", data)
}

func submitted(res http.ResponseWriter, req *http.Request) {

	tpl.ExecuteTemplate(res, "submitted.gohtml", nil)
}

func viewsubmissions(res http.ResponseWriter, req *http.Request) {

	if !alreadyLoggedIn(req) {
		http.Redirect(res, req, "/", http.StatusSeeOther)
		return
	}
	s := dsa.LOtraversal(submissions, priorities, products, statuses, categories)
	tpl.ExecuteTemplate(res, "viewsubmissions.gohtml", s)
}

func viewmytickets(res http.ResponseWriter, req *http.Request) {

	if !alreadyLoggedIn(req) {
		http.Redirect(res, req, "/", http.StatusSeeOther)
		return
	}
	str := make([][]string, 0)
	str = dsa.IOtraversal(mytickets.Root, priorities, products, statuses, categories, str)
	owner := "My"
	object := "Tickets"

	data := struct {
		Tickets [][]string
		Owner   string
		Object  string
	}{
		str,
		owner,
		object,
	}
	tpl.ExecuteTemplate(res, "viewtickets.gohtml", data)
}

func viewmyassignments(res http.ResponseWriter, req *http.Request) {

	if !alreadyLoggedIn(req) {
		http.Redirect(res, req, "/", http.StatusSeeOther)
		return
	}
	str := make([][]string, 0)
	str = dsa.IOtraversal(myassigns.Root, priorities, products, statuses, categories, str)
	owner := "My"
	object := "Assignments"

	data := struct {
		Tickets [][]string
		Owner   string
		Object  string
	}{
		str,
		owner,
		object,
	}
	tpl.ExecuteTemplate(res, "viewtickets.gohtml", data)
}

func viewalltickets(res http.ResponseWriter, req *http.Request) {

	if !alreadyLoggedIn(req) {
		http.Redirect(res, req, "/", http.StatusSeeOther)
		return
	}
	str := make([][]string, 0)
	str = dsa.IOtraversal(ticketlog.Root, priorities, products, statuses, categories, str)
	owner := "All"
	object := "Tickets"

	data := struct {
		Tickets [][]string
		Owner   string
		Object  string
	}{
		str,
		owner,
		object,
	}
	tpl.ExecuteTemplate(res, "viewtickets.gohtml", data)
}

func deletemytickets(res http.ResponseWriter, req *http.Request) {

	if !alreadyLoggedIn(req) {
		http.Redirect(res, req, "/", http.StatusSeeOther)
		return
	}

	var deleteID int64

	if req.Method == http.MethodPost {
		deleteIDraw, deleteerr := strconv.Atoi(req.FormValue("deleteID"))
		deleteID = int64(deleteIDraw)
		if deleteerr != nil || req.FormValue("deleteID") == "" || deleteID < 0 || dsa.AVLsearch(mytickets.Root, deleteID) == nil {
			ticketRecord.AddLog(fmt.Sprintf("Attempted ticket deletion by user %v, but invalid ticket ID input.", loggedin.Name))
			http.Error(res, "Invalid ticket ID input.", http.StatusForbidden)
			return
		}
		todelete := dsa.AVLsearch(ticketlog.Root, deleteID)
		ticketlog.Root = dsa.AVLdelete(ticketlog.Root, todelete.Ticket.TicketID)
		ticketRecord.AddLog(fmt.Sprintf("Ticket ID %v has been deleted by user %v.", todelete.Ticket.TicketID, loggedin.Name))
	}
	str := make([][]string, 0)
	str = dsa.IOtraversal(mytickets.Root, priorities, products, statuses, categories, str)
	owner := "My"
	object := "Tickets"

	data := struct {
		Tickets [][]string
		Owner   string
		Object  string
	}{
		str,
		owner,
		object,
	}

	tpl.ExecuteTemplate(res, "deletetickets.gohtml", data)
}

func markmyassignments(res http.ResponseWriter, req *http.Request) {

	if !alreadyLoggedIn(req) {
		http.Redirect(res, req, "/", http.StatusSeeOther)
		return
	}

	var deleteID int64

	if req.Method == http.MethodPost {
		deleteIDraw, deleteerr := strconv.Atoi(req.FormValue("deleteID"))
		deleteID = int64(deleteIDraw)
		if deleteerr != nil || req.FormValue("deleteID") == "" || deleteID < 0 || dsa.AVLsearch(mytickets.Root, deleteID) == nil {
			submissionRecord.AddLog(fmt.Sprintf("Attempted assignment clearing by user %v, but invalid ticket ID input.", loggedin.Name))
			http.Error(res, "Invalid ticket ID input.", http.StatusForbidden)
			return
		}
		todelete := dsa.AVLsearch(ticketlog.Root, deleteID)
		ticketlog.Root = dsa.AVLdelete(ticketlog.Root, todelete.Ticket.TicketID)
		ticketRecord.AddLog(fmt.Sprintf("Ticket ID %v has been marked complete by user %v.", todelete.Ticket.TicketID, loggedin.Name))
	}
	str := make([][]string, 0)
	str = dsa.IOtraversal(myassigns.Root, priorities, products, statuses, categories, str)
	owner := "My"
	object := "Assignments"

	data := struct {
		Tickets [][]string
		Owner   string
		Object  string
	}{
		str,
		owner,
		object,
	}

	tpl.ExecuteTemplate(res, "deletetickets.gohtml", data)
}

func resorttickets(res http.ResponseWriter, req *http.Request) {
	if !alreadyLoggedIn(req) {
		http.Redirect(res, req, "/", http.StatusSeeOther)
		return
	}

	// Possible pivot options
	options := []string{
		"Product",
		"Status",
		"Category",
		"Estimated Hours to Complete",
		"Priority",
		"Start Date",
		"Due Date",
		"Creator",
		"Assignee",
		"Title",
		"Description",
	}
	prepivots := passPreload(options)

	// Concurrently pre-pivot tickets for all approaches
	if req.Method == http.MethodPost {
		label := req.FormValue("criteria")
		var pivotedtickets *dsa.AVLtree
		for i := 0; i < len(options); i++ {
			pivoted := <-prepivots
			if pivoted.label == label {
				pivotedtickets = pivoted.pivoted
			}
		}

		str := make([][]string, 0)
		str = dsa.IOtraversal(pivotedtickets.Root, priorities, products, statuses, categories, str)

		owner := "All"
		object := fmt.Sprintf("Tickets (resorted by %s)", label)

		data := struct {
			Tickets [][]string
			Owner   string
			Object  string
		}{
			str,
			owner,
			object,
		}

		tpl.ExecuteTemplate(res, "viewtickets.gohtml", data)
		return
	}

	tpl.ExecuteTemplate(res, "resorttickets.gohtml", options)
}

func logout(res http.ResponseWriter, req *http.Request) {

	if !alreadyLoggedIn(req) {
		http.Redirect(res, req, "/", http.StatusSeeOther)
		return
	}
	// Cookie management
	currsesh, _ := req.Cookie("myCookie")
	delete(mapSessions, currsesh.Value)
	currsesh.MaxAge = -1
	http.SetCookie(res, currsesh)
	userRecord.AddLog(fmt.Sprintf("User %v has logged out. Session cookie has been deleted.", loggedin.Name))

	// Saving Submissions to CSV
	submissionsCSV.SaveSubmissions(submissions)
	ticketsCSV.SaveTickets(ticketlog)
	productsCSV.SaveProducts(products)
	usersCSV.SaveUsers(users)
	generalRecord.AddLog("Submissions, Tickets, Products and Users Saved.")

	http.Redirect(res, req, "/", http.StatusSeeOther)
}
